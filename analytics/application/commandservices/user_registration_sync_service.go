package commandservices

import (
	"analytics/analytics/domain/model/aggregates"
	"analytics/analytics/domain/model/valueobjects"
	"analytics/analytics/domain/repositories"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

// UserRegistrationSyncService maneja la sincronización de eventos de registro de usuarios de Kafka
type UserRegistrationSyncService struct {
	kafkaBrokers []string
	topic        string
	repository   repositories.UserRegistrationAnalyticsRepository
}

// NewUserRegistrationSyncService crea una nueva instancia del servicio de sincronización
func NewUserRegistrationSyncService(kafkaBrokers []string, topic string, repository repositories.UserRegistrationAnalyticsRepository) *UserRegistrationSyncService {
	return &UserRegistrationSyncService{
		kafkaBrokers: kafkaBrokers,
		topic:        topic,
		repository:   repository,
	}
}

// KafkaUserRegisteredEvent representa la estructura del evento en Kafka
type KafkaUserRegisteredEvent struct {
	UserID     string  `json:"userId"`
	ProfileID  string  `json:"profileId"`
	Username   string  `json:"username"`
	ProfileURL *string `json:"profileUrl"`
	OccurredOn []int   `json:"occurredOn"` // [year, month, day, hour, minute, second, nano]
}

// SyncAllEvents sincroniza todos los eventos disponibles del tópico de Kafka
func (s *UserRegistrationSyncService) SyncAllEvents(ctx context.Context) (int, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V3_0_0_0

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
		// Obtener el offset más antiguo
		partitionConsumer, err := consumer.ConsumePartition(s.topic, partition, sarama.OffsetOldest)
		if err != nil {
			log.Printf("Error consuming partition %d: %v", partition, err)
			continue
		}

		log.Printf("Syncing partition %d of topic %s", partition, s.topic)

		messageCount := 0
		lastMessageTime := time.Now()
		idleTimeout := 3 * time.Second // Esperar 3 segundos sin mensajes antes de considerar que terminó

	messageLoop:
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				if msg == nil {
					break messageLoop
				}

				// Procesar mensaje
				if err := s.processMessage(ctx, msg); err != nil {
					log.Printf("Error processing message from partition %d, offset %d: %v",
						partition, msg.Offset, err)
					continue
				}

				messageCount++
				totalSynced++
				lastMessageTime = time.Now()

			case err := <-partitionConsumer.Errors():
				log.Printf("Error from partition %d: %v", partition, err)

			case <-time.After(idleTimeout):
				// Si no hemos recibido mensajes en idleTimeout segundos, asumir que terminamos
				if time.Since(lastMessageTime) >= idleTimeout {
					log.Printf("No more messages in partition %d after %v, processed %d messages",
						partition, idleTimeout, messageCount)
					break messageLoop
				}

			case <-ctx.Done():
				log.Printf("Context cancelled, stopping sync")
				partitionConsumer.Close()
				return totalSynced, ctx.Err()
			}
		}

		partitionConsumer.Close()
		log.Printf("Finished syncing partition %d, processed %d messages", partition, messageCount)
	}

	log.Printf("Sync completed. Total events synced: %d", totalSynced)
	return totalSynced, nil
}

// processMessage procesa un mensaje de Kafka
func (s *UserRegistrationSyncService) processMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	var event KafkaUserRegisteredEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("error unmarshaling event: %w", err)
	}

	// Verificar si ya existe
	userID, err := valueobjects.NewUserID(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	existing, err := s.repository.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error checking existing user: %w", err)
	}

	// Si ya existe, saltar (idempotencia)
	if existing != nil {
		log.Printf("User %s already exists, skipping", event.UserID)
		return nil
	}

	// Convertir a aggregate
	userReg, err := s.eventToAggregate(&event)
	if err != nil {
		return fmt.Errorf("error converting event to aggregate: %w", err)
	}

	// Guardar
	if err := s.repository.Save(ctx, userReg); err != nil {
		return fmt.Errorf("error saving user registration: %w", err)
	}

	log.Printf("Successfully synced user registration: %s (%s)", event.UserID, event.Username)
	return nil
}

// eventToAggregate convierte el evento de Kafka a aggregate de dominio
func (s *UserRegistrationSyncService) eventToAggregate(event *KafkaUserRegisteredEvent) (*aggregates.UserRegistrationAnalytics, error) {
	userID, err := valueobjects.NewUserID(event.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	profileID, err := valueobjects.NewProfileID(event.ProfileID)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	registeredAt, err := s.parseOccurredOn(event.OccurredOn)
	if err != nil {
		return nil, fmt.Errorf("invalid registered date: %w", err)
	}

	return aggregates.NewUserRegistrationAnalytics(
		userID,
		profileID,
		event.Username,
		event.ProfileURL,
		registeredAt,
	)
}

// parseOccurredOn convierte el array [year, month, day, hour, minute, second, nano] a time.Time
func (s *UserRegistrationSyncService) parseOccurredOn(arr []int) (time.Time, error) {
	if len(arr) < 6 {
		return time.Time{}, fmt.Errorf("invalid registeredAt array, expected at least 6 elements, got %d", len(arr))
	}

	year := arr[0]
	month := time.Month(arr[1])
	day := arr[2]
	hour := arr[3]
	minute := arr[4]
	second := arr[5]

	nano := 0
	if len(arr) >= 7 {
		nano = arr[6]
	}

	return time.Date(year, month, day, hour, minute, second, nano, time.UTC), nil
}
