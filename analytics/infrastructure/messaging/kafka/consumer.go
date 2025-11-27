package kafka

import (
	"analytics/analytics/domain/model/aggregates"
	"analytics/analytics/domain/model/entities"
	"analytics/analytics/domain/model/valueobjects"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// EventHandler define el contrato para procesar eventos
type EventHandler interface {
	HandleExecutionAnalyticsEvent(ctx context.Context, execution *aggregates.ExecutionAnalytics) error
}

// Consumer representa el consumidor de Kafka
type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	handler       EventHandler
}

// NewConsumer crea una nueva instancia del consumidor
func NewConsumer(brokers []string, groupID, topic string, handler EventHandler) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V3_0_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("error creating consumer group: %w", err)
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		topic:         topic,
		handler:       handler,
	}, nil
}

// Start inicia el consumo de mensajes
func (c *Consumer) Start(ctx context.Context) error {
	handler := &consumerGroupHandler{
		consumer: c,
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer...")
			return c.consumerGroup.Close()
		default:
			if err := c.consumerGroup.Consume(ctx, []string{c.topic}, handler); err != nil {
				log.Printf("Error consuming messages: %v", err)
			}
		}
	}
}

// Close cierra el consumidor
func (c *Consumer) Close() error {
	return c.consumerGroup.Close()
}

// consumerGroupHandler implementa sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	consumer *Consumer
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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

func (h *consumerGroupHandler) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	log.Printf("Received message from topic %s, partition %d, offset %d",
		message.Topic, message.Partition, message.Offset)

	// Deserializar el evento
	var event ExecutionAnalyticsEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("error unmarshaling event: %w", err)
	}

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
