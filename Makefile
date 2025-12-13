.PHONY: proto generate run build migrateup migratedown migratecreate

# Load environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Database connection string
DB_URL = postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable&timezone=$(DB_TIMEZONE)

# Setup
setup:
	@echo "Setting up order service..."
	which migrate > /dev/null || go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	which sqlc > /dev/null || go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	which protoc > /dev/null || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	which protoc-gen-go-grpc > /dev/null || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go mod download
	go mod tidy
	go mod verify
	@echo "✅ Order service setup complete!"

# Generate Go code from proto files
proto:
	@echo "Generating Go code from proto files..."
	@mkdir -p go-proto
	protoc --go_out=. --go_opt=paths=import \
		--go-grpc_out=. --go-grpc_opt=paths=import \
		--proto_path=proto \
		proto/modules/*.proto proto/services/*.proto
	@if [ -d "order-service/go-proto" ]; then \
		mv order-service/go-proto/* go-proto/ 2>/dev/null || true; \
		rmdir order-service/go-proto 2>/dev/null || true; \
		rmdir order-service 2>/dev/null || true; \
	fi
	@echo "Fixing import paths in generated files..."
	@find go-proto -name "*.pb.go" -type f -exec sed -i '' 's|"go-proto/|"order-service/go-proto/|g' {} \;
	@echo "✅ Proto files generated!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "✅ Dependencies installed!"

# Generate Go code from SQL queries
sqlc:
	@echo "Generating Go code from SQL queries..."
	@if ! command -v sqlc > /dev/null; then \
		echo "Installing sqlc..."; \
		go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest; \
	fi
	sqlc generate
	@echo "✅ SQL code generated!"

generate:
	make proto
	make sqlc
	@echo "✅ Generate complete!"

db-up:
	make run-db
	make migrateup
	@echo "✅ DB complete!"

db-down:
	make stop-db
	make remove-db
	@echo "✅ DB down complete!"

# Build the service
build:
	@echo "Building order service..."
	go build -o bin/order-service cmd/server/main.go
	@echo "✅ Build complete!"

run-db:
	docker run -d --name $(DB_CONTAINER_NAME) -p $(DB_PORT):5432 \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		$(POSTGRES_IMAGE)
	@echo "✅ Database running!"

remove-db:
	docker rm $(DB_CONTAINER_NAME)
	@echo "✅ Database removed!"

stop-db:
	docker stop $(DB_CONTAINER_NAME)
	@echo "✅ Database stopped!"

start-db:
	docker start $(DB_CONTAINER_NAME)
	@echo "✅ Database started!"

restart-db:
	docker restart $(DB_CONTAINER_NAME)
	@echo "✅ Database restarted!"

createdb:
	docker exec -it $(DB_CONTAINER_NAME) createdb --username=$(DB_USER) --owner=$(DB_USER) $(DB_NAME)
	@echo "✅ Database created!"
dropdb:
	docker exec -it $(DB_CONTAINER_NAME) dropdb $(DB_NAME)
	@echo "✅ Database dropped!"

migrateup:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" -verbose down

migratecreate:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

run:
	go run cmd/server/main.go

# Hot reload with Air
dev:
	@echo "Starting Order Service with hot reload..."
	@which air > /dev/null || go install github.com/air-verse/air@latest
	air -c .air.toml

# Debug with Delve
debug:
	@echo "Starting Order Service in debug mode..."
	@which dlv > /dev/null || go install github.com/go-delve/delve/cmd/dlv@latest
	dlv debug ./cmd/server/main.go --headless --listen=:2346 --api-version=2 --accept-multiclient

# Build for debug (with debug symbols)
build-debug:
	go build -gcflags="all=-N -l" -o bin/order-service-debug cmd/server/main.go
	@echo "✅ Debug build complete!"