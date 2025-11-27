package commandservices

import (
	"analytics/analytics/domain/model/aggregates"
	"analytics/analytics/domain/model/entities"
	"analytics/analytics/domain/model/valueobjects"
	"analytics/analytics/domain/repositories"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

// SyncService maneja la sincronización de eventos de Kafka
type SyncService struct {
	kafkaBrokers []string
	topic        string
	repository   repositories.ExecutionAnalyticsRepository
}

// NewSyncService crea una nueva instancia del servicio de sincronización
func NewSyncService(kafkaBrokers []string, topic string, repository repositories.ExecutionAnalyticsRepository) *SyncService {
	return &SyncService{
		kafkaBrokers: kafkaBrokers,
		topic:        topic,
		repository:   repository,
	}
}

// KafkaEvent representa la estructura del evento en Kafka
type KafkaEvent struct {
	ExecutionID    string            `json:"execution_id"`
	ChallengeID    string            `json:"challenge_id"`
	CodeVersionID  string            `json:"code_version_id"`
	StudentID      string            `json:"student_id"`
	Language       string            `json:"language"`
	Status         string            `json:"status"`
	Timestamp      time.Time         `json:"timestamp"`
	ExecutionTime  int64             `json:"execution_time_ms"`
	ExitCode       int               `json:"exit_code"`
	TotalTests     int               `json:"total_tests"`
	PassedTests    int               `json:"passed_tests"`
	FailedTests    int               `json:"failed_tests"`
	Success        bool              `json:"success"`
	ServerInstance string            `json:"server_instance,omitempty"`
	TestResults    []KafkaTestResult `json:"test_results"`
}

// KafkaTestResult representa un resultado de test en el evento
type KafkaTestResult struct {
	TestID       string `json:"test_id"`
	TestName     string `json:"test_name"`
	Passed       bool   `json:"passed"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// SyncAllEvents sincroniza todos los eventos disponibles del tópico de Kafka
func (s *SyncService) SyncAllEvents(ctx context.Context) (int, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_6_0_0

	// Crear consumer
	consumer, err := sarama.NewConsumer(s.kafkaBrokers, config)
	if err != nil {
		return 0, fmt.Errorf("error creating Kafka consumer: %w", err)
	}
	defer consumer.Close()

	// Obtener particiones del tópico
	partitions, err := consumer.Partitions(s.topic)
	if err != nil {
		return 0, fmt.Errorf("error getting partitions: %w", err)
	}

	totalSynced := 0

	// Consumir de cada partición
	for _, partition := range partitions {
		// Obtener el offset más antiguo y el más reciente
		oldestOffset, err := consumer.ConsumePartition(s.topic, partition, sarama.OffsetOldest)
		if err != nil {
			log.Printf("Error consuming partition %d: %v", partition, err)
			continue
		}

		log.Printf("Syncing partition %d of topic %s", partition, s.topic)

		// Timeout para evitar esperar indefinidamente
		timeout := time.After(10 * time.Second)
		messageCount := 0

	messageLoop:
		for {
			select {
			case msg := <-oldestOffset.Messages():
				if msg == nil {
					break messageLoop
				}

				// Procesar mensaje
				var event KafkaEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					log.Printf("Error unmarshaling message: %v", err)
					continue
				}

				// Convertir a domain model
				execution, err := s.eventToDomain(event)
				if err != nil {
					log.Printf("Error converting event to domain: %v", err)
					continue
				}

				// Verificar si ya existe
				existing, err := s.repository.FindByExecutionID(ctx, execution.ExecutionID())
				if err != nil {
					log.Printf("Error checking existing execution: %v", err)
					continue
				}

				if existing == nil {
					// Guardar en la base de datos
					if err := s.repository.Save(ctx, execution); err != nil {
						log.Printf("Error saving execution: %v", err)
						continue
					}
					totalSynced++
					messageCount++
					log.Printf("Synced execution: %s (Partition %d, Offset %d)",
						execution.ExecutionID().Value(), partition, msg.Offset)
				}

			case err := <-oldestOffset.Errors():
				log.Printf("Partition %d error: %v", partition, err)
				break messageLoop

			case <-timeout:
				log.Printf("Timeout reached for partition %d after syncing %d messages", partition, messageCount)
				break messageLoop
			}
		}

		oldestOffset.Close()
		log.Printf("Completed partition %d: synced %d new messages", partition, messageCount)
	}

	log.Printf("Sync completed: %d total new executions synced", totalSynced)
	return totalSynced, nil
}

// eventToDomain convierte un evento de Kafka al modelo de dominio
func (s *SyncService) eventToDomain(event KafkaEvent) (*aggregates.ExecutionAnalytics, error) {
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
		return nil, fmt.Errorf("invalid language: %w", err)
	}

	status, err := valueobjects.NewExecutionStatus(event.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}

	// Crear el aggregate
	execution, err := aggregates.NewExecutionAnalytics(
		executionID,
		challengeID,
		event.CodeVersionID,
		studentID,
		language,
		status,
		event.Timestamp,
		event.ExecutionTime,
		event.ExitCode,
		event.TotalTests,
		event.PassedTests,
		event.FailedTests,
		event.Success,
		event.ServerInstance,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating execution analytics: %w", err)
	}

	// Agregar test results
	for _, tr := range event.TestResults {
		testID, err := valueobjects.NewTestID(tr.TestID)
		if err != nil {
			log.Printf("Skipping invalid test ID %s: %v", tr.TestID, err)
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
