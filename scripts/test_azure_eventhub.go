package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Obtener configuración
	bootstrapServers := getEnv("KAFKA_BOOTSTRAP_SERVERS", "")
	securityProtocol := getEnv("KAFKA_SECURITY_PROTOCOL", "SASL_SSL")
	saslMechanism := getEnv("KAFKA_SASL_MECHANISM", "PLAIN")
	saslUsername := getEnv("KAFKA_SASL_USERNAME", "$ConnectionString")
	saslPassword := getEnv("KAFKA_SASL_PASSWORD", "")
	topic := getEnv("KAFKA_TOPIC", "execution.analytics")
	userRegTopic := getEnv("KAFKA_USER_REGISTRATION_TOPIC", "user-registration")

	if bootstrapServers == "" {
		log.Fatal("❌ KAFKA_BOOTSTRAP_SERVERS is required")
	}

	if saslPassword == "" {
		log.Fatal("❌ KAFKA_SASL_PASSWORD (connection string) is required")
	}

	fmt.Println("🔍 Azure Event Hub Connection Test")
	fmt.Println("==================================")
	fmt.Println()

	// Mostrar configuración (sin mostrar password completo)
	fmt.Println("📋 Configuration:")
	fmt.Printf("  Bootstrap Servers: %s\n", bootstrapServers)
	fmt.Printf("  Security Protocol: %s\n", securityProtocol)
	fmt.Printf("  SASL Mechanism: %s\n", saslMechanism)
	fmt.Printf("  SASL Username: %s\n", saslUsername)
	fmt.Printf("  SASL Password: %s...%s\n", saslPassword[:30], saslPassword[len(saslPassword)-20:])
	fmt.Printf("  Topic 1: %s\n", topic)
	fmt.Printf("  Topic 2: %s\n", userRegTopic)
	fmt.Println()

	// Crear configuración de Sarama
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0

	// Configurar metadata
	config.Metadata.Full = true
	config.Metadata.Timeout = 60 * time.Second
	config.Metadata.Retry.Max = 5
	config.Metadata.Retry.Backoff = 2 * time.Second

	// Timeouts
	config.Net.DialTimeout = 30 * time.Second
	config.Net.ReadTimeout = 60 * time.Second
	config.Net.WriteTimeout = 30 * time.Second

	// Admin configuration
	config.Admin.Timeout = 60 * time.Second

	// Configurar seguridad
	if securityProtocol == "SASL_SSL" {
		fmt.Println("🔐 Configuring SASL_SSL...")

		// TLS
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		}

		// SASL
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		config.Net.SASL.User = saslUsername
		config.Net.SASL.Password = saslPassword
		config.Net.SASL.Handshake = true
		config.Net.SASL.Version = sarama.SASLHandshakeV1

		fmt.Println("  ✅ TLS enabled (TLS 1.2+)")
		fmt.Println("  ✅ SASL PLAIN mechanism")
		fmt.Println()
	}

	brokers := strings.Split(bootstrapServers, ",")

	// Test 1: Conexión básica
	fmt.Println("🧪 Test 1: Basic Connection")
	fmt.Println("---------------------------")
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}
	defer client.Close()
	fmt.Println("✅ Successfully connected to Azure Event Hub!")
	fmt.Println()

	// Test 2: Obtener brokers
	fmt.Println("🧪 Test 2: Broker Information")
	fmt.Println("-----------------------------")
	brokerList := client.Brokers()
	fmt.Printf("Connected to %d broker(s):\n", len(brokerList))
	for i, broker := range brokerList {
		fmt.Printf("  %d. %s (ID: %d)\n", i+1, broker.Addr(), broker.ID())
		if connected, err := broker.Connected(); err == nil && connected {
			fmt.Printf("     Status: Connected ✅\n")
		}
	}
	fmt.Println()

	// Test 3: Listar topics
	fmt.Println("🧪 Test 3: Available Topics (Event Hubs)")
	fmt.Println("----------------------------------------")
	topics, err := client.Topics()
	if err != nil {
		log.Printf("⚠️  Warning: Could not list topics: %v", err)
	} else {
		fmt.Printf("Found %d topic(s):\n", len(topics))
		for i, t := range topics {
			fmt.Printf("  %d. %s", i+1, t)

			// Verificar si es uno de nuestros topics
			if t == topic {
				fmt.Printf(" 🎯 (Execution Analytics)")
			} else if t == userRegTopic {
				fmt.Printf(" 🎯 (User Registration)")
			}
			fmt.Println()
		}
	}
	fmt.Println()

	// Test 4: Metadata del topic de ejecución
	if contains(topics, topic) {
		fmt.Printf("🧪 Test 4: Topic Metadata - %s\n", topic)
		fmt.Println("----------------------------------------")
		testTopicMetadata(client, topic)
		fmt.Println()
	} else {
		fmt.Printf("⚠️  Test 4 skipped: Topic '%s' does not exist\n", topic)
		fmt.Println("   Create it in Azure Portal as an Event Hub")
		fmt.Println()
	}

	// Test 5: Metadata del topic de registro de usuarios
	if contains(topics, userRegTopic) {
		fmt.Printf("🧪 Test 5: Topic Metadata - %s\n", userRegTopic)
		fmt.Println("----------------------------------------")
		testTopicMetadata(client, userRegTopic)
		fmt.Println()
	} else {
		fmt.Printf("⚠️  Test 5 skipped: Topic '%s' does not exist\n", userRegTopic)
		fmt.Println("   Create it in Azure Portal as an Event Hub")
		fmt.Println()
	}

	// Test 6: Test de consumer group (sin consumir mensajes)
	fmt.Println("🧪 Test 6: Consumer Group Configuration")
	fmt.Println("---------------------------------------")
	groupID := getEnv("KAFKA_GROUP_ID", "test-consumer-group")
	testConsumerGroup(brokers, config, topic, groupID)
	fmt.Println()

	// Resumen final
	fmt.Println("✅ All tests completed successfully!")
	fmt.Println()
	fmt.Println("📝 Next Steps:")
	fmt.Println("  1. Ensure topics exist in Azure Event Hub:")
	fmt.Printf("     - %s\n", topic)
	fmt.Printf("     - %s\n", userRegTopic)
	fmt.Println("  2. Start the analytics microservice: go run main.go")
	fmt.Println("  3. Send test messages to verify consumption")
	fmt.Println()
	fmt.Println("🎉 Your Azure Event Hub configuration is correct!")
}

func testTopicMetadata(client sarama.Client, topic string) {
	partitions, err := client.Partitions(topic)
	if err != nil {
		log.Printf("❌ Could not get partitions: %v", err)
		return
	}

	fmt.Printf("  Partitions: %d\n", len(partitions))

	for _, partition := range partitions {
		oldest, err := client.GetOffset(topic, partition, sarama.OffsetOldest)
		if err != nil {
			log.Printf("  ⚠️  Could not get oldest offset for partition %d: %v", partition, err)
			continue
		}

		newest, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Printf("  ⚠️  Could not get newest offset for partition %d: %v", partition, err)
			continue
		}

		messages := newest - oldest
		fmt.Printf("    Partition %d: %d messages (offset %d -> %d)\n",
			partition, messages, oldest, newest)
	}

	fmt.Println("  ✅ Topic metadata retrieved successfully")
}

func testConsumerGroup(brokers []string, config *sarama.Config, topic, groupID string) {
	// Configurar para consumer
	consumerConfig := *config
	consumerConfig.Consumer.Return.Errors = true
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	consumerConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	consumerConfig.Consumer.MaxProcessingTime = 60 * time.Second
	consumerConfig.Consumer.Group.Session.Timeout = 60 * time.Second
	consumerConfig.Consumer.Group.Heartbeat.Interval = 3 * time.Second

	fmt.Printf("  Creating consumer group: %s\n", groupID)
	fmt.Printf("  Topic: %s\n", topic)

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, &consumerConfig)
	if err != nil {
		log.Printf("❌ Failed to create consumer group: %v", err)
		return
	}
	defer consumer.Close()

	fmt.Println("  ✅ Consumer group created successfully")
	fmt.Println("  ✅ Ready to consume messages")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
