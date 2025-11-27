package repositories

import (
	"analytics/analytics/domain/model/aggregates"
	"analytics/analytics/domain/model/valueobjects"
	"context"
	"time"
)

// ExecutionAnalyticsRepository define el contrato para el repositorio
type ExecutionAnalyticsRepository interface {
	// Save guarda o actualiza un ExecutionAnalytics
	Save(ctx context.Context, execution *aggregates.ExecutionAnalytics) error

	// FindByExecutionID busca por ID de ejecución
	FindByExecutionID(ctx context.Context, executionID valueobjects.ExecutionID) (*aggregates.ExecutionAnalytics, error)

	// FindByStudentID busca todas las ejecuciones de un estudiante
	FindByStudentID(ctx context.Context, studentID valueobjects.StudentID, limit, offset int) ([]*aggregates.ExecutionAnalytics, error)

	// FindByChallengeID busca todas las ejecuciones de un challenge
	FindByChallengeID(ctx context.Context, challengeID valueobjects.ChallengeID, limit, offset int) ([]*aggregates.ExecutionAnalytics, error)

	// FindByDateRange busca ejecuciones en un rango de fechas
	FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*aggregates.ExecutionAnalytics, error)

	// CountByStudentID cuenta ejecuciones de un estudiante
	CountByStudentID(ctx context.Context, studentID valueobjects.StudentID) (int64, error)

	// CountByChallengeID cuenta ejecuciones de un challenge
	CountByChallengeID(ctx context.Context, challengeID valueobjects.ChallengeID) (int64, error)

	// GetSuccessRateByStudent obtiene tasa de éxito por estudiante
	GetSuccessRateByStudent(ctx context.Context, studentID valueobjects.StudentID) (float64, error)

	// GetSuccessRateByChallenge obtiene tasa de éxito por challenge
	GetSuccessRateByChallenge(ctx context.Context, challengeID valueobjects.ChallengeID) (float64, error)

	// GetAverageExecutionTimeByChallenge obtiene tiempo promedio de ejecución por challenge
	GetAverageExecutionTimeByChallenge(ctx context.Context, challengeID valueobjects.ChallengeID) (float64, error)

	// GetDailyExecutionStats obtiene estadísticas diarias de ejecuciones
	GetDailyExecutionStats(ctx context.Context, startDate, endDate time.Time) ([]DailyStats, error)

	// GetLanguageUsageStats obtiene estadísticas de uso de lenguajes
	GetLanguageUsageStats(ctx context.Context, startDate, endDate time.Time) ([]LanguageStats, error)

	// GetTopFailedChallenges obtiene los challenges con más fallos
	GetTopFailedChallenges(ctx context.Context, limit int) ([]ChallengeStats, error)
}

// DailyStats representa estadísticas diarias
type DailyStats struct {
	Date            time.Time
	TotalExecutions int64
	SuccessfulExecs int64
	FailedExecs     int64
	AvgExecTime     float64
}

// LanguageStats representa estadísticas por lenguaje
type LanguageStats struct {
	Language        string
	TotalExecutions int64
	SuccessRate     float64
}

// ChallengeStats representa estadísticas por challenge
type ChallengeStats struct {
	ChallengeID     string
	TotalExecutions int64
	SuccessRate     float64
	AvgExecTime     float64
}
