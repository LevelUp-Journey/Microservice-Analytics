package kafka

import (
	"analytics/analytics/domain/model/aggregates"
	"analytics/analytics/domain/model/valueobjects"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

// NewUserRegistrationConsumer crea una nueva instancia del consumidor
func NewUserRegistrationConsumer(brokers []string, groupID, topic string, handler UserRegistrationEventHandler) (*UserRegistrationConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V3_0_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("error creating consumer group: %w", err)
	}

	return &UserRegistrationConsumer{
		consumerGroup: consumerGroup,
		topic:         topic,
		handler:       handler,
	}, nil
}

// Start inicia el consumo de mensajes
func (c *UserRegistrationConsumer) Start(ctx context.Context) error {
	handler := &userRegistrationGroupHandler{
		consumer: c,
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping User Registration Kafka consumer...")
			return c.consumerGroup.Close()
		default:
			if err := c.consumerGroup.Consume(ctx, []string{c.topic}, handler); err != nil {
				log.Printf("Error consuming messages: %v", err)
			}
		}
	}
}

// Close cierra el consumidor
func (c *UserRegistrationConsumer) Close() error {
	return c.consumerGroup.Close()
}

// userRegistrationGroupHandler implementa sarama.ConsumerGroupHandler
type userRegistrationGroupHandler struct {
	consumer *UserRegistrationConsumer
}

func (h *userRegistrationGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *userRegistrationGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *userRegistrationGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if err := h.processMessage(session.Context(), message); err != nil {
			log.Printf("Error processing message: %v", err)
			// Continuar procesando otros mensajes incluso si uno falla
			continue
		}
		session.MarkMessage(message, "")
	}
	return nil
}

func (h *userRegistrationGroupHandler) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	log.Printf("Received user registration message from topic %s, partition %d, offset %d",
		message.Topic, message.Partition, message.Offset)

	// Deserializar el evento
	var event UserRegisteredEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("error unmarshaling event: %w", err)
	}

	// Convertir evento a dominio
	userReg, err := h.eventToDomain(&event)
	if err != nil {
		return fmt.Errorf("error converting event to domain: %w", err)
	}

	// Procesar el evento usando el handler
	if err := h.consumer.handler.HandleUserRegistrationEvent(ctx, userReg); err != nil {
		return fmt.Errorf("error handling event: %w", err)
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
