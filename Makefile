.PHONY: proto generate run build

# Generate Go code from proto files
proto:
	@echo "Generating Go code from proto files..."
	@mkdir -p go-proto
	protoc --go_out=./go-proto --go_opt=paths=source_relative \
		--go-grpc_out=./go-proto --go-grpc_opt=paths=source_relative \
		--proto_path=proto \
		proto/modules/*.proto proto/services/*.proto
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