package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nanab/analytics-service/analytics/domain/model/aggregates"
	"github.com/nanab/analytics-service/analytics/domain/model/valueobjects"

	"github.com/IBM/sarama"
)

// UserRegistrationEventHandler define el contrato para procesar eventos de registro de usuarios
type UserRegistrationEventHandler interface {
	HandleUserRegistrationEvent(ctx context.Context, userReg *aggregates.UserRegistrationAnalytics) error
}

// UserRegistrationConsumer representa el consumidor de Kafka para eventos de registro de usuarios
type UserRegistrationConsumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	handler       UserRegistrationEventHandler
}

// NewUserRegistrationConsumer crea una nueva instancia del consumidor compatible con Azure Event Hub
func NewUserRegistrationConsumer(brokers []string, groupID, topic string, handler UserRegistrationEventHandler) (*UserRegistrationConsumer, error) {
	config := &ConsumerConfig{
		Brokers:          brokers,
		GroupID:          groupID,
		Topic:            topic,
		SecurityProtocol: "PLAINTEXT",
		EnableAutoCommit: true,
		RequestTimeoutMs: 60000,
		SessionTimeoutMs: 60000,
	}
	return NewUserRegistrationConsumerWithConfig(config, handler)
}

// NewUserRegistrationConsumerWithConfig crea una nueva instancia del consumidor con configuración personalizada
func NewUserRegistrationConsumerWithConfig(cfg *ConsumerConfig, handler UserRegistrationEventHandler) (*UserRegistrationConsumer, error) {
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
		log.Println("Configuring SASL_SSL for User Registration Consumer (Azure Event Hub)...")

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

		log.Printf("SASL configured for User Registration Consumer with mechanism: %s, username: %s", cfg.SaslMechanism, cfg.SaslUsername)
	} else if cfg.SecurityProtocol == "SASL_PLAINTEXT" {
		log.Println("Configuring SASL_PLAINTEXT for User Registration Consumer...")

		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		config.Net.SASL.User = cfg.SaslUsername
		config.Net.SASL.Password = cfg.SaslPassword
		config.Net.SASL.Handshake = true
	}

	// Crear consumer group
	log.Printf("Creating user registration consumer group with brokers: %v, groupID: %s", cfg.Brokers, cfg.GroupID)
	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("error creating user registration consumer group: %w", err)
	}

	log.Printf("User registration consumer group created successfully for topic: %s", cfg.Topic)

	return &UserRegistrationConsumer{
		consumerGroup: consumerGroup,
		topic:         cfg.Topic,
		handler:       handler,
	}, nil
}

// Start inicia el consumo de mensajes
func (c *UserRegistrationConsumer) Start(ctx context.Context) error {
	handler := &userRegistrationGroupHandler{
		consumer: c,
	}

	// Goroutine para manejar errores del consumer group
	go func() {
		for err := range c.consumerGroup.Errors() {
			log.Printf("User registration consumer group error: %v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping User Registration Kafka consumer...")
			return c.consumerGroup.Close()
		default:
			log.Printf("Starting user registration consumer session for topic: %s", c.topic)
			if err := c.consumerGroup.Consume(ctx, []string{c.topic}, handler); err != nil {
				log.Printf("Error consuming user registration messages: %v", err)
				// Esperar un poco antes de reintentar
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// Close cierra el consumidor
func (c *UserRegistrationConsumer) Close() error {
	log.Println("Closing User Registration Kafka consumer...")
	return c.consumerGroup.Close()
}

// userRegistrationGroupHandler implementa sarama.ConsumerGroupHandler
type userRegistrationGroupHandler struct {
	consumer *UserRegistrationConsumer
}

func (h *userRegistrationGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("User registration consumer group session setup - MemberID: %s, GenerationID: %d",
		session.MemberID(), session.GenerationID())
	return nil
}

func (h *userRegistrationGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("User registration consumer group session cleanup - MemberID: %s", session.MemberID())
	return nil
}

func (h *userRegistrationGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("Starting to consume user registration partition %d from offset %d", claim.Partition(), claim.InitialOffset())

	for {
		select {
		case <-session.Context().Done():
			log.Println("User registration session context done, stopping consumption")
			return nil
		case message, ok := <-claim.Messages():
			if !ok {
				log.Println("User registration message channel closed")
				return nil
			}

			if err := h.processMessage(session.Context(), message); err != nil {
				log.Printf("Error processing user registration message (offset %d): %v", message.Offset, err)
				// Continuar procesando otros mensajes incluso si uno falla
				continue
			}

			// Marcar el mensaje como procesado
			session.MarkMessage(message, "")
		}
	}
}

func (h *userRegistrationGroupHandler) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	log.Printf("Received user registration message from topic %s, partition %d, offset %d, timestamp: %v",
		message.Topic, message.Partition, message.Offset, message.Timestamp)

	// Deserializar el evento
	var event UserRegisteredEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("error unmarshaling user registration event: %w", err)
	}

	log.Printf("Processing user registration event: UserID=%s, Username=%s, ProfileID=%s",
		event.UserID, event.Username, event.ProfileID)

	// Convertir evento a dominio
	userReg, err := h.eventToDomain(&event)
	if err != nil {
		return fmt.Errorf("error converting user registration event to domain: %w", err)
	}

	// Procesar el evento usando el handler
	if err := h.consumer.handler.HandleUserRegistrationEvent(ctx, userReg); err != nil {
		return fmt.Errorf("error handling user registration event: %w", err)
	}

	log.Printf("Successfully processed user registration: %s (username: %s)", event.UserID, event.Username)
	return nil
}

// eventToDomain convierte el evento de Kafka a un aggregate de dominio
func (h *userRegistrationGroupHandler) eventToDomain(event *UserRegisteredEvent) (*aggregates.UserRegistrationAnalytics, error) {
	// Crear value objects
	userID, err := valueobjects.NewUserID(event.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	profileID, err := valueobjects.NewProfileID(event.ProfileID)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	// Convertir el array de OccurredOn a time.Time
	registeredAt, err := parseOccurredOn(event.OccurredOn)
	if err != nil {
		return nil, fmt.Errorf("invalid occurred date: %w", err)
	}

	// Crear aggregate
	userReg, err := aggregates.NewUserRegistrationAnalytics(
		userID,
		profileID,
		event.Username,
		event.ProfileURL,
		registeredAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating user registration analytics aggregate: %w", err)
	}

	return userReg, nil
}

// parseOccurredOn convierte el array [year, month, day, hour, minute, second, nano] a time.Time
func parseOccurredOn(arr []int) (time.Time, error) {
	if len(arr) < 6 {
		return time.Time{}, fmt.Errorf("invalid occurredOn array, expected at least 6 elements, got %d", len(arr))
	}

	// Extraer componentes
	year := arr[0]
	month := time.Month(arr[1])
	day := arr[2]
	hour := arr[3]
	minute := arr[4]
	second := arr[5]

	// Nano es opcional (puede ser el 7mo elemento)
	nano := 0
	if len(arr) >= 7 {
		nano = arr[6]
	}

	// Crear el time.Time en UTC
	return time.Date(year, month, day, hour, minute, second, nano, time.UTC), nil
}
