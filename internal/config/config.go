package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// gRPC Server
	GRPCPort int

	// Kafka
	KafkaBrokers           string
	KafkaTopicOrderCreated string

	// Database
	DatabaseURL string
	DBHost      string
	DBPort      int
	DBUser      string
	DBPassword  string
	DBName      string
}

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	config := &Config{}

	// gRPC Server
	config.GRPCPort = getEnvAsInt("GRPC_PORT", 5001)

	// Kafka
	config.KafkaBrokers = getEnv("KAFKA_BROKERS", "localhost:9092")
	config.KafkaTopicOrderCreated = getEnv("KAFKA_TOPIC_ORDER_CREATED", "order.created")

	// Database
	config.DatabaseURL = getEnv("DATABASE_URL", "")
	if config.DatabaseURL == "" {
		// Build from individual components
		config.DBHost = getEnv("DB_HOST", "localhost")
		config.DBPort = getEnvAsInt("DB_PORT", 5432)
		config.DBUser = getEnv("DB_USER", "postgres")
		config.DBPassword = getEnv("DB_PASSWORD", "postgres")
		config.DBName = getEnv("DB_NAME", "orderdb")
		config.DatabaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName,
		)
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
