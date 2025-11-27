package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nanab/analytics-service/analytics/domain/model/aggregates"
	"github.com/nanab/analytics-service/analytics/domain/model/entities"
	"github.com/nanab/analytics-service/analytics/domain/model/valueobjects"

	"github.com/IBM/sarama"
)

// EventHandler define el contrato para procesar eventos
type EventHandler interface {
	HandleExecutionAnalyticsEvent(ctx context.Context, execution *aggregates.ExecutionAnalytics) error
}

// ConsumerConfig contiene la configuración para el consumidor de Kafka
type ConsumerConfig struct {
	Brokers          []string
	GroupID          string
	Topic            string
	SecurityProtocol string
	SaslMechanism    string
	SaslUsername     string
	SaslPassword     string
	RequestTimeoutMs int
	SessionTimeoutMs int
	EnableAutoCommit bool
}

// Consumer representa el consumidor de Kafka
type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	handler       EventHandler
}

// NewConsumer crea una nueva instancia del consumidor compatible con Azure Event Hub
func NewConsumer(brokers []string, groupID, topic string, handler EventHandler) (*Consumer, error) {
	config := &ConsumerConfig{
		Brokers:          brokers,
		GroupID:          groupID,
		Topic:            topic,
		SecurityProtocol: "PLAINTEXT",
		EnableAutoCommit: true,
		RequestTimeoutMs: 60000,
		SessionTimeoutMs: 60000,
	}
	return NewConsumerWithConfig(config, handler)
}

// NewConsumerWithConfig crea una nueva instancia del consumidor con configuración personalizada
func NewConsumerWithConfig(cfg *ConsumerConfig, handler EventHandler) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0 // Azure Event Hub es compatible con Kafka 1.0+

	// Consumer group settings
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	// Auto commit settings
	if cfg.EnableAutoCommit {
		config.Consumer.Offsets.AutoCommit.Enable = true
		config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	}

	// Timeouts - Azure Event Hub requiere timeouts más altos
	config.Net.DialTimeout = 30 * time.Second
	config.Net.ReadTimeout = time.Duration(cfg.RequestTimeoutMs) * time.Millisecond
	config.Net.WriteTimeout = 30 * time.Second

	config.Consumer.MaxProcessingTime = time.Duration(cfg.SessionTimeoutMs) * time.Millisecond
	config.Consumer.Group.Session.Timeout = time.Duration(cfg.SessionTimeoutMs) * time.Millisecond
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second

	// Configuración de metadata para Azure Event Hub
	config.Metadata.Retry.Max = 5
	config.Metadata.Retry.Backoff = 2 * time.Second
	config.Metadata.Timeout = 60 * time.Second
	config.Metadata.Full = false

	// Configuración de seguridad para Azure Event Hub
	if cfg.SecurityProtocol == "SASL_SSL" {
		log.Println("Configuring SASL_SSL for Azure Event Hub...")

		// Habilitar TLS
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		}

		// Configurar SASL
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext // Azure Event Hub usa PLAIN
		config.Net.SASL.User = cfg.SaslUsername              // Debe ser "$ConnectionString"
		config.Net.SASL.Password = cfg.SaslPassword          // Connection string completo
		config.Net.SASL.Handshake = true
		config.Net.SASL.Version = sarama.SASLHandshakeV1

		log.Printf("SASL configured with mechanism: %s, username: %s", cfg.SaslMechanism, cfg.SaslUsername)
	} else if cfg.SecurityProtocol == "SASL_PLAINTEXT" {
		log.Println("Configuring SASL_PLAINTEXT...")

		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		config.Net.SASL.User = cfg.SaslUsername
		config.Net.SASL.Password = cfg.SaslPassword
		config.Net.SASL.Handshake = true
	}

	// Crear consumer group
	log.Printf("Creating consumer group with brokers: %v, groupID: %s", cfg.Brokers, cfg.GroupID)
	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("error creating consumer group: %w", err)
	}

	log.Printf("Consumer group created successfully for topic: %s", cfg.Topic)

	return &Consumer{
		consumerGroup: consumerGroup,
		topic:         cfg.Topic,
		handler:       handler,
	}, nil
}

// Start inicia el consumo de mensajes
func (c *Consumer) Start(ctx context.Context) error {
	handler := &consumerGroupHandler{
		consumer: c,
	}

	// Goroutine para manejar errores del consumer group
	go func() {
		for err := range c.consumerGroup.Errors() {
			log.Printf("Consumer group error: %v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer...")
			return c.consumerGroup.Close()
		default:
			log.Printf("Starting consumer session for topic: %s", c.topic)
			if err := c.consumerGroup.Consume(ctx, []string{c.topic}, handler); err != nil {
				log.Printf("Error consuming messages: %v", err)
				// Esperar un poco antes de reintentar
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// Close cierra el consumidor
func (c *Consumer) Close() error {
	log.Println("Closing Kafka consumer...")
	return c.consumerGroup.Close()
}

// consumerGroupHandler implementa sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	consumer *Consumer
}

func (h *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("Consumer group session setup - MemberID: %s, GenerationID: %d",
		session.MemberID(), session.GenerationID())
	return nil
}

func (h *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("Consumer group session cleanup - MemberID: %s", session.MemberID())
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("Starting to consume partition %d from offset %d", claim.Partition(), claim.InitialOffset())

	for {
		select {
		case <-session.Context().Done():
			log.Println("Session context done, stopping consumption")
			return nil
		case message, ok := <-claim.Messages():
			if !ok {
				log.Println("Message channel closed")
				return nil
			}

			if err := h.processMessage(session.Context(), message); err != nil {
				log.Printf("Error processing message (offset %d): %v", message.Offset, err)
				// Continuar procesando otros mensajes incluso si uno falla
				continue
			}

			// Marcar el mensaje como procesado
			session.MarkMessage(message, "")
		}
	}
}

func (h *consumerGroupHandler) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	log.Printf("Received message from topic %s, partition %d, offset %d, timestamp: %v",
		message.Topic, message.Partition, message.Offset, message.Timestamp)

	// Deserializar el evento
	var event ExecutionAnalyticsEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("error unmarshaling event: %w", err)
	}

	log.Printf("Processing execution analytics event: ExecutionID=%s, ChallengeID=%s, StudentID=%s",
		event.ExecutionID, event.ChallengeID, event.StudentID)

	// Convertir evento a dominio
	execution, err := h.eventToDomain(&event)
	if err != nil {
		return fmt.Errorf("error converting event to domain: %w", err)
	}

	// Procesar el evento usando el handler
	if err := h.consumer.handler.HandleExecutionAnalyticsEvent(ctx, execution); err != nil {
		return fmt.Errorf("error handling event: %w", err)
	}

	log.Printf("Successfully processed execution analytics: %s", event.ExecutionID)
	return nil
}

// eventToDomain convierte el evento de Kafka a un aggregate de dominio
func (h *consumerGroupHandler) eventToDomain(event *ExecutionAnalyticsEvent) (*aggregates.ExecutionAnalytics, error) {
	// Crear value objects
	executionID, err := valueobjects.NewExecutionID(event.ExecutionID)
	if err != nil {
		return nil, fmt.Errorf("invalid execution ID: %w", err)
	}

	challengeID, err := valueobjects.NewChallengeID(event.ChallengeID)
	if err != nil {
		return nil, fmt.Errorf("invalid challenge ID: %w", err)
	}

	studentID, err := valueobjects.NewStudentID(event.StudentID)
	if err != nil {
		return nil, fmt.Errorf("invalid student ID: %w", err)
	}

	language, err := valueobjects.NewProgrammingLanguage(event.Language)
	if err != nil {
		return nil, fmt.Errorf("invalid programming language: %w", err)
	}

	status, err := valueobjects.NewExecutionStatus(event.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid execution status: %w", err)
	}

	// Crear aggregate
	execution, err := aggregates.NewExecutionAnalytics(
		executionID,
		challengeID,
		event.CodeVersionID,
		studentID,
		language,
		status,
		event.Timestamp,
		event.ExecutionTimeMs,
		event.ExitCode,
		event.TotalTests,
		event.PassedTests,
		event.FailedTests,
		event.Success,
		event.ServerInstance,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating execution analytics aggregate: %w", err)
	}

	// Agregar test results
	for _, tr := range event.TestResults {
		testID, err := valueobjects.NewTestID(tr.TestID)
		if err != nil {
			log.Printf("Warning: invalid test ID %s, skipping test result", tr.TestID)
			continue
		}

		testResult := entities.NewTestResult(
			testID,
			tr.TestName,
			tr.Passed,
			tr.ErrorMessage,
		)
		execution.AddTestResult(testResult)
	}

	return execution, nil
}
