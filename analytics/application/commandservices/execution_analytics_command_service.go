package commandservices

import (
	"github.com/nanab/analytics-service/analytics/domain/model/aggregates"
	"github.com/nanab/analytics-service/analytics/domain/repositories"
	"context"
	"fmt"
	"log"
)

// ExecutionAnalyticsCommandService maneja comandos para ExecutionAnalytics
type ExecutionAnalyticsCommandService struct {
	repository repositories.ExecutionAnalyticsRepository
}

// NewExecutionAnalyticsCommandService crea una nueva instancia del servicio
func NewExecutionAnalyticsCommandService(repository repositories.ExecutionAnalyticsRepository) *ExecutionAnalyticsCommandService {
	return &ExecutionAnalyticsCommandService{
		repository: repository,
	}
}

// SaveExecutionAnalytics guarda un nuevo registro de analytics
func (s *ExecutionAnalyticsCommandService) SaveExecutionAnalytics(ctx context.Context, execution *aggregates.ExecutionAnalytics) error {
	// Verificar si ya existe
	existing, err := s.repository.FindByExecutionID(ctx, execution.ExecutionID())
	if err != nil {
		return fmt.Errorf("error checking existing execution: %w", err)
	}

	if existing != nil {
		log.Printf("Execution analytics already exists for execution ID: %s", execution.ExecutionID().Value())
		return nil // Idempotencia: no es error si ya existe
	}

	// Guardar nuevo registro
	if err := s.repository.Save(ctx, execution); err != nil {
		return fmt.Errorf("error saving execution analytics: %w", err)
	}

	log.Printf("Saved execution analytics: %s (Student: %s, Challenge: %s, Success: %v)",
		execution.ExecutionID().Value(),
		execution.StudentID().Value(),
		execution.ChallengeID().Value(),
		execution.Success(),
	)

	return nil
}

// HandleExecutionAnalyticsEvent implementa EventHandler de Kafka
func (s *ExecutionAnalyticsCommandService) HandleExecutionAnalyticsEvent(ctx context.Context, execution *aggregates.ExecutionAnalytics) error {
	return s.SaveExecutionAnalytics(ctx, execution)
}
