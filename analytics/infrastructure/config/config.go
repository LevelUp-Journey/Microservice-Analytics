package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config almacena toda la configuración de la aplicación
type Config struct {
	Server struct {
		Port string
		IP   string
	}
	Database struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
		SSLMode  string
	}
	Kafka struct {
		BootstrapServers []string
		GroupID          string
		Topic            string
		SecurityProtocol string
	}
	ServiceDiscovery struct {
		URL         string
		ServiceName string
		Enabled     bool
	}
}

// Load carga la configuración desde variables de entorno
func Load() (*Config, error) {
	// Intentar cargar .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{}

	// Server configuration
	config.Server.Port = getEnv("SERVER_PORT", "8080")
	config.Server.IP = getEnv("SERVER_IP", "127.0.0.1")

	// Database configuration
	config.Database.Host = getEnv("DB_HOST", "localhost")
	config.Database.Port = getEnv("DB_PORT", "5432")
	config.Database.User = getEnv("DB_USER", "postgres")
	config.Database.Password = getEnv("DB_PASSWORD", "postgres")
	config.Database.Name = getEnv("DB_NAME", "analytics_db")
	config.Database.SSLMode = getEnv("DB_SSLMODE", "disable")

	// Kafka configuration
	bootstrapServers := getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
	config.Kafka.BootstrapServers = strings.Split(bootstrapServers, ",")
	config.Kafka.GroupID = getEnv("KAFKA_GROUP_ID", "analytics-consumer-group")
	config.Kafka.Topic = getEnv("KAFKA_TOPIC", "execution.analytics")
	config.Kafka.SecurityProtocol = getEnv("KAFKA_SECURITY_PROTOCOL", "PLAINTEXT")

	// Service Discovery configuration
	config.ServiceDiscovery.URL = getEnv("SERVICE_DISCOVERY_URL", "http://127.0.0.1:8761/eureka/")
	config.ServiceDiscovery.ServiceName = getEnv("SERVICE_NAME", "analytics-service")
	config.ServiceDiscovery.Enabled = getEnv("SERVICE_DISCOVERY_ENABLED", "false") == "true"

	return config, nil
}

// GetDatabaseDSN retorna el DSN para PostgreSQL
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetServerAddress retorna la dirección del servidor
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.IP, c.Server.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
