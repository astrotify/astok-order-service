package main

import (
	"log"
	"order-service/internal/config"
	"order-service/internal/database"
	"order-service/internal/grpc"
	"order-service/internal/kafka"
	"order-service/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Database connected successfully")

	// Initialize Kafka producer
	producer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("❌ Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()
	log.Println("✅ Kafka producer connected")

	// Initialize service
	orderService := service.NewOrderService(db, producer, cfg.KafkaTopicOrderCreated)

	// Initialize gRPC handler
	orderHandler := grpc.NewOrderGrpcHandler(orderService)

	// Start gRPC server
	log.Println("🚀 Starting Order Service...")
	if err := grpc.StartGRPCServer(cfg.GRPCPort, orderHandler); err != nil {
		log.Fatalf("❌ Failed to start gRPC server: %v", err)
	}
}
