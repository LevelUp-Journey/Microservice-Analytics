package queryservices

import (
	"analytics/analytics/domain/model/aggregates"
	"analytics/analytics/domain/model/valueobjects"
	"analytics/analytics/domain/repositories"
	"context"
	"fmt"
	"time"
)

// ExecutionAnalyticsQueryService maneja consultas para ExecutionAnalytics
type ExecutionAnalyticsQueryService struct {
	repository repositories.ExecutionAnalyticsRepository
}

// NewExecutionAnalyticsQueryService crea una nueva instancia del servicio
func NewExecutionAnalyticsQueryService(repository repositories.ExecutionAnalyticsRepository) *ExecutionAnalyticsQueryService {
	return &ExecutionAnalyticsQueryService{
		repository: repository,
	}
}

// GetByExecutionID obtiene analytics por ID de ejecución
func (s *ExecutionAnalyticsQueryService) GetByExecutionID(ctx context.Context, executionID string) (*aggregates.ExecutionAnalytics, error) {
	id, err := valueobjects.NewExecutionID(executionID)
	if err != nil {
		return nil, fmt.Errorf("invalid execution ID: %w", err)
	}

	return s.repository.FindByExecutionID(ctx, id)
}

// GetByStudentID obtiene todas las ejecuciones de un estudiante
func (s *ExecutionAnalyticsQueryService) GetByStudentID(ctx context.Context, studentID string, page, pageSize int) ([]*aggregates.ExecutionAnalytics, error) {
	id, err := valueobjects.NewStudentID(studentID)
	if err != nil {
		return nil, fmt.Errorf("invalid student ID: %w", err)
	}

	offset := (page - 1) * pageSize
	return s.repository.FindByStudentID(ctx, id, pageSize, offset)
}

// GetByChallengeID obtiene todas las ejecuciones de un challenge
func (s *ExecutionAnalyticsQueryService) GetByChallengeID(ctx context.Context, challengeID string, page, pageSize int) ([]*aggregates.ExecutionAnalytics, error) {
	id, err := valueobjects.NewChallengeID(challengeID)
	if err != nil {
		return nil, fmt.Errorf("invalid challenge ID: %w", err)
	}

	offset := (page - 1) * pageSize
	return s.repository.FindByChallengeID(ctx, id, pageSize, offset)
}

// GetByDateRange obtiene ejecuciones en un rango de fechas
func (s *ExecutionAnalyticsQueryService) GetByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*aggregates.ExecutionAnalytics, error) {
	offset := (page - 1) * pageSize
	return s.repository.FindByDateRange(ctx, startDate, endDate, pageSize, offset)
}

// GetStudentSuccessRate obtiene la tasa de éxito de un estudiante
func (s *ExecutionAnalyticsQueryService) GetStudentSuccessRate(ctx context.Context, studentID string) (float64, error) {
	id, err := valueobjects.NewStudentID(studentID)
	if err != nil {
		return 0, fmt.Errorf("invalid student ID: %w", err)
	}

	return s.repository.GetSuccessRateByStudent(ctx, id)
}

// GetChallengeSuccessRate obtiene la tasa de éxito de un challenge
func (s *ExecutionAnalyticsQueryService) GetChallengeSuccessRate(ctx context.Context, challengeID string) (float64, error) {
	id, err := valueobjects.NewChallengeID(challengeID)
	if err != nil {
		return 0, fmt.Errorf("invalid challenge ID: %w", err)
	}

	return s.repository.GetSuccessRateByChallenge(ctx, id)
}

// GetChallengeAverageExecutionTime obtiene el tiempo promedio de ejecución de un challenge
func (s *ExecutionAnalyticsQueryService) GetChallengeAverageExecutionTime(ctx context.Context, challengeID string) (float64, error) {
	id, err := valueobjects.NewChallengeID(challengeID)
	if err != nil {
		return 0, fmt.Errorf("invalid challenge ID: %w", err)
	}

	return s.repository.GetAverageExecutionTimeByChallenge(ctx, id)
}

// GetDailyStats obtiene estadísticas diarias
func (s *ExecutionAnalyticsQueryService) GetDailyStats(ctx context.Context, startDate, endDate time.Time) ([]repositories.DailyStats, error) {
	return s.repository.GetDailyExecutionStats(ctx, startDate, endDate)
}

// GetLanguageStats obtiene estadísticas por lenguaje
func (s *ExecutionAnalyticsQueryService) GetLanguageStats(ctx context.Context, startDate, endDate time.Time) ([]repositories.LanguageStats, error) {
	return s.repository.GetLanguageUsageStats(ctx, startDate, endDate)
}

// GetTopFailedChallenges obtiene los challenges con más fallos
func (s *ExecutionAnalyticsQueryService) GetTopFailedChallenges(ctx context.Context, limit int) ([]repositories.ChallengeStats, error) {
	return s.repository.GetTopFailedChallenges(ctx, limit)
}

// GetStudentExecutionCount obtiene el conteo de ejecuciones de un estudiante
func (s *ExecutionAnalyticsQueryService) GetStudentExecutionCount(ctx context.Context, studentID string) (int64, error) {
	id, err := valueobjects.NewStudentID(studentID)
	if err != nil {
		return 0, fmt.Errorf("invalid student ID: %w", err)
	}

	return s.repository.CountByStudentID(ctx, id)
}

// GetChallengeExecutionCount obtiene el conteo de ejecuciones de un challenge
func (s *ExecutionAnalyticsQueryService) GetChallengeExecutionCount(ctx context.Context, challengeID string) (int64, error) {
	id, err := valueobjects.NewChallengeID(challengeID)
	if err != nil {
		return 0, fmt.Errorf("invalid challenge ID: %w", err)
	}

	return s.repository.CountByChallengeID(ctx, id)
}
