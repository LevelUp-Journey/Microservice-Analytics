package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config almacena toda la configuración de la aplicación
type Config struct {
	Server struct {
		Port     string
		IP       string
		Hostname string // Hostname o IP pública para registro en Eureka
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
		SaslMechanism    string
		SaslUsername     string
		SaslPassword     string
		// Azure Event Hub specific settings
		RequestTimeoutMs int
		SessionTimeoutMs int
		EnableAutoCommit bool
	}
	KafkaUserRegistration struct {
		Topic   string
		GroupID string
	}
	ServiceDiscovery struct {
		URL         string
		ServiceName string
		Enabled     bool
		InstanceIP  string // IP específica para registro en Eureka (opcional)
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
	config.Server.Hostname = getEnv("SERVER_HOSTNAME", "") // Hostname público para Eureka

	// Database configuration
	config.Database.Host = getEnv("DB_HOST", "localhost")
	config.Database.Port = getEnv("DB_PORT", "5432")
	config.Database.User = getEnv("DB_USER", "postgres")
	config.Database.Password = getEnv("DB_PASSWORD", "postgres")
	config.Database.Name = getEnv("DB_NAME", "analytics_db")
	config.Database.SSLMode = getEnv("DB_SSLMODE", "disable")

	// Kafka configuration - Compatible con Azure Event Hub
	bootstrapServers := getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
	config.Kafka.BootstrapServers = strings.Split(bootstrapServers, ",")
	config.Kafka.GroupID = getEnv("KAFKA_GROUP_ID", "analytics-consumer-group")
	config.Kafka.Topic = getEnv("KAFKA_TOPIC", "execution.analytics")

	// Security configuration for Azure Event Hub
	config.Kafka.SecurityProtocol = getEnv("KAFKA_SECURITY_PROTOCOL", "PLAINTEXT")
	config.Kafka.SaslMechanism = getEnv("KAFKA_SASL_MECHANISM", "PLAIN")

	// Azure Event Hub connection configuration
	// For Azure Event Hub, username is always "$ConnectionString"
	config.Kafka.SaslUsername = getEnv("KAFKA_SASL_USERNAME", "$ConnectionString")

	// Password is the full connection string for Azure Event Hub
	config.Kafka.SaslPassword = getEnv("KAFKA_SASL_PASSWORD", "")

	// Azure Event Hub specific timeouts
	config.Kafka.RequestTimeoutMs = getEnvAsInt("KAFKA_REQUEST_TIMEOUT_MS", 60000)
	config.Kafka.SessionTimeoutMs = getEnvAsInt("KAFKA_SESSION_TIMEOUT_MS", 60000)
	config.Kafka.EnableAutoCommit = getEnvAsBool("KAFKA_ENABLE_AUTO_COMMIT", true)

	// Kafka User Registration configuration
	config.KafkaUserRegistration.Topic = getEnv("KAFKA_USER_REGISTRATION_TOPIC", "iam.user.registered")
	config.KafkaUserRegistration.GroupID = getEnv("KAFKA_USER_REGISTRATION_GROUP_ID", "user-registration-analytics-group")

	// Service Discovery configuration
	config.ServiceDiscovery.URL = getEnv("SERVICE_DISCOVERY_URL", "http://127.0.0.1:8761/eureka/")
	config.ServiceDiscovery.ServiceName = getEnv("SERVICE_NAME", "analytics-service")
	config.ServiceDiscovery.Enabled = getEnvAsBool("SERVICE_DISCOVERY_ENABLED", false)
	config.ServiceDiscovery.InstanceIP = getEnv("EUREKA_INSTANCE_IP", "") // IP para registro en Eureka

	// Log configuration (sin mostrar credenciales sensibles)
	log.Printf("Kafka Configuration:")
	log.Printf("  Bootstrap Servers: %v", config.Kafka.BootstrapServers)
	log.Printf("  Security Protocol: %s", config.Kafka.SecurityProtocol)
	log.Printf("  SASL Mechanism: %s", config.Kafka.SaslMechanism)
	log.Printf("  Group ID: %s", config.Kafka.GroupID)
	log.Printf("  Topic: %s", config.Kafka.Topic)
	log.Printf("  User Registration Topic: %s", config.KafkaUserRegistration.Topic)

	if config.Kafka.SecurityProtocol == "SASL_SSL" && config.Kafka.SaslPassword != "" {
		log.Printf("  Azure Event Hub: Configured ✓")
	}

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

// IsSaslEnabled verifica si SASL está habilitado
func (c *Config) IsSaslEnabled() bool {
	return c.Kafka.SecurityProtocol == "SASL_SSL" || c.Kafka.SecurityProtocol == "SASL_PLAINTEXT"
}

// getEnv obtiene una variable de entorno o retorna un valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt obtiene una variable de entorno como entero o retorna un valor por defecto
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid integer value for %s, using default: %d", key, defaultValue)
		return defaultValue
	}
	return value
}

// getEnvAsBool obtiene una variable de entorno como booleano o retorna un valor por defecto
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid boolean value for %s, using default: %t", key, defaultValue)
		return defaultValue
	}
	return value
}
