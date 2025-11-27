package repositories

import (
	"github.com/nanab/analytics-service/analytics/domain/model/aggregates"
	"github.com/nanab/analytics-service/analytics/domain/model/entities"
	"github.com/nanab/analytics-service/analytics/domain/model/valueobjects"
	"github.com/nanab/analytics-service/analytics/domain/repositories"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// PostgresExecutionAnalyticsRepository implementa el repositorio usando PostgreSQL
type PostgresExecutionAnalyticsRepository struct {
	db *gorm.DB
}

// NewPostgresExecutionAnalyticsRepository crea una nueva instancia del repositorio
func NewPostgresExecutionAnalyticsRepository(db *gorm.DB) repositories.ExecutionAnalyticsRepository {
	return &PostgresExecutionAnalyticsRepository{db: db}
}

// Save guarda o actualiza un ExecutionAnalytics
func (r *PostgresExecutionAnalyticsRepository) Save(ctx context.Context, execution *aggregates.ExecutionAnalytics) error {
	model := r.toModel(execution)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model).Error; err != nil {
			return err
		}

		// Actualizar el ID del aggregate
		execution.SetID(model.ID)
		return nil
	})
}

// FindByExecutionID busca por ID de ejecución
func (r *PostgresExecutionAnalyticsRepository) FindByExecutionID(ctx context.Context, executionID valueobjects.ExecutionID) (*aggregates.ExecutionAnalytics, error) {
	var model ExecutionAnalyticsModel

	if err := r.db.WithContext(ctx).
		Preload("TestResults").
		Where("execution_id = ?", executionID.Value()).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomain(&model)
}

// FindByStudentID busca todas las ejecuciones de un estudiante
func (r *PostgresExecutionAnalyticsRepository) FindByStudentID(ctx context.Context, studentID valueobjects.StudentID, limit, offset int) ([]*aggregates.ExecutionAnalytics, error) {
	var models []ExecutionAnalyticsModel

	query := r.db.WithContext(ctx).
		Preload("TestResults").
		Where("student_id = ?", studentID.Value()).
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return r.toDomainList(models)
}

// FindByChallengeID busca todas las ejecuciones de un challenge
func (r *PostgresExecutionAnalyticsRepository) FindByChallengeID(ctx context.Context, challengeID valueobjects.ChallengeID, limit, offset int) ([]*aggregates.ExecutionAnalytics, error) {
	var models []ExecutionAnalyticsModel

	query := r.db.WithContext(ctx).
		Preload("TestResults").
		Where("challenge_id = ?", challengeID.Value()).
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return r.toDomainList(models)
}

// FindByDateRange busca ejecuciones en un rango de fechas
func (r *PostgresExecutionAnalyticsRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*aggregates.ExecutionAnalytics, error) {
	var models []ExecutionAnalyticsModel

	query := r.db.WithContext(ctx).
		Preload("TestResults").
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return r.toDomainList(models)
}

// CountByStudentID cuenta ejecuciones de un estudiante
func (r *PostgresExecutionAnalyticsRepository) CountByStudentID(ctx context.Context, studentID valueobjects.StudentID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ExecutionAnalyticsModel{}).
		Where("student_id = ?", studentID.Value()).
		Count(&count).Error
	return count, err
}

// CountByChallengeID cuenta ejecuciones de un challenge
func (r *PostgresExecutionAnalyticsRepository) CountByChallengeID(ctx context.Context, challengeID valueobjects.ChallengeID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ExecutionAnalyticsModel{}).
		Where("challenge_id = ?", challengeID.Value()).
		Count(&count).Error
	return count, err
}

// GetSuccessRateByStudent obtiene tasa de éxito por estudiante
func (r *PostgresExecutionAnalyticsRepository) GetSuccessRateByStudent(ctx context.Context, studentID valueobjects.StudentID) (float64, error) {
	var result struct {
		SuccessRate float64
	}

	err := r.db.WithContext(ctx).
		Model(&ExecutionAnalyticsModel{}).
		Select("AVG(CASE WHEN success = true THEN 100.0 ELSE 0.0 END) as success_rate").
		Where("student_id = ?", studentID.Value()).
		Scan(&result).Error

	return result.SuccessRate, err
}

// GetSuccessRateByChallenge obtiene tasa de éxito por challenge
func (r *PostgresExecutionAnalyticsRepository) GetSuccessRateByChallenge(ctx context.Context, challengeID valueobjects.ChallengeID) (float64, error) {
	var result struct {
		SuccessRate float64
	}

	err := r.db.WithContext(ctx).
		Model(&ExecutionAnalyticsModel{}).
		Select("AVG(CASE WHEN success = true THEN 100.0 ELSE 0.0 END) as success_rate").
		Where("challenge_id = ?", challengeID.Value()).
		Scan(&result).Error

	return result.SuccessRate, err
}

// GetAverageExecutionTimeByChallenge obtiene tiempo promedio de ejecución por challenge
func (r *PostgresExecutionAnalyticsRepository) GetAverageExecutionTimeByChallenge(ctx context.Context, challengeID valueobjects.ChallengeID) (float64, error) {
	var result struct {
		AvgTime float64
	}

	err := r.db.WithContext(ctx).
		Model(&ExecutionAnalyticsModel{}).
		Select("AVG(execution_time_ms) as avg_time").
		Where("challenge_id = ?", challengeID.Value()).
		Scan(&result).Error

	return result.AvgTime, err
}

// GetDailyExecutionStats obtiene estadísticas diarias de ejecuciones
func (r *PostgresExecutionAnalyticsRepository) GetDailyExecutionStats(ctx context.Context, startDate, endDate time.Time) ([]repositories.DailyStats, error) {
	var results []repositories.DailyStats

	err := r.db.WithContext(ctx).
		Model(&ExecutionAnalyticsModel{}).
		Select(`
			DATE(timestamp) as date,
			COUNT(*) as total_executions,
			SUM(CASE WHEN success = true THEN 1 ELSE 0 END) as successful_execs,
			SUM(CASE WHEN success = false THEN 1 ELSE 0 END) as failed_execs,
			AVG(execution_time_ms) as avg_exec_time
		`).
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Group("DATE(timestamp)").
		Order("date DESC").
		Scan(&results).Error

	return results, err
}

// GetLanguageUsageStats obtiene estadísticas de uso de lenguajes
func (r *PostgresExecutionAnalyticsRepository) GetLanguageUsageStats(ctx context.Context, startDate, endDate time.Time) ([]repositories.LanguageStats, error) {
	var results []repositories.LanguageStats

	err := r.db.WithContext(ctx).
		Model(&ExecutionAnalyticsModel{}).
		Select(`
			language,
			COUNT(*) as total_executions,
			AVG(CASE WHEN success = true THEN 100.0 ELSE 0.0 END) as success_rate
		`).
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Group("language").
		Order("total_executions DESC").
		Scan(&results).Error

	return results, err
}

// GetTopFailedChallenges obtiene los challenges con más fallos
func (r *PostgresExecutionAnalyticsRepository) GetTopFailedChallenges(ctx context.Context, limit int) ([]repositories.ChallengeStats, error) {
	var results []repositories.ChallengeStats

	query := r.db.WithContext(ctx).
		Model(&ExecutionAnalyticsModel{}).
		Select(`
			challenge_id,
			COUNT(*) as total_executions,
			AVG(CASE WHEN success = true THEN 100.0 ELSE 0.0 END) as success_rate,
			AVG(execution_time_ms) as avg_exec_time
		`).
		Group("challenge_id").
		Order("success_rate ASC, total_executions DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Scan(&results).Error
	return results, err
}

// toModel convierte del dominio a modelo de persistencia
func (r *PostgresExecutionAnalyticsRepository) toModel(execution *aggregates.ExecutionAnalytics) ExecutionAnalyticsModel {
	model := ExecutionAnalyticsModel{
		ID:              execution.ID(),
		ExecutionID:     execution.ExecutionID().Value(),
		ChallengeID:     execution.ChallengeID().Value(),
		CodeVersionID:   execution.CodeVersionID(),
		StudentID:       execution.StudentID().Value(),
		Language:        execution.Language().Value(),
		Status:          execution.Status().Value(),
		Timestamp:       execution.Timestamp(),
		ExecutionTimeMs: execution.ExecutionTimeMs(),
		ExitCode:        execution.ExitCode(),
		TotalTests:      execution.TotalTests(),
		PassedTests:     execution.PassedTests(),
		FailedTests:     execution.FailedTests(),
		Success:         execution.Success(),
		ServerInstance:  execution.ServerInstance(),
		CreatedAt:       execution.CreatedAt(),
		UpdatedAt:       execution.UpdatedAt(),
		TestResults:     make([]TestResultModel, 0),
	}

	for _, testResult := range execution.TestResults() {
		model.TestResults = append(model.TestResults, TestResultModel{
			TestID:       testResult.TestID().Value(),
			TestName:     testResult.TestName(),
			Passed:       testResult.Passed(),
			ErrorMessage: testResult.ErrorMessage(),
		})
	}

	return model
}

// toDomain convierte del modelo de persistencia al dominio
func (r *PostgresExecutionAnalyticsRepository) toDomain(model *ExecutionAnalyticsModel) (*aggregates.ExecutionAnalytics, error) {
	executionID, err := valueobjects.NewExecutionID(model.ExecutionID)
	if err != nil {
		return nil, err
	}

	challengeID, err := valueobjects.NewChallengeID(model.ChallengeID)
	if err != nil {
		return nil, err
	}

	studentID, err := valueobjects.NewStudentID(model.StudentID)
	if err != nil {
		return nil, err
	}

	language, err := valueobjects.NewProgrammingLanguage(model.Language)
	if err != nil {
		return nil, err
	}

	status, err := valueobjects.NewExecutionStatus(model.Status)
	if err != nil {
		return nil, err
	}

	execution, err := aggregates.NewExecutionAnalytics(
		executionID,
		challengeID,
		model.CodeVersionID,
		studentID,
		language,
		status,
		model.Timestamp,
		model.ExecutionTimeMs,
		model.ExitCode,
		model.TotalTests,
		model.PassedTests,
		model.FailedTests,
		model.Success,
		model.ServerInstance,
	)
	if err != nil {
		return nil, err
	}

	execution.SetID(model.ID)
	execution.SetCreatedAt(model.CreatedAt)
	execution.SetUpdatedAt(model.UpdatedAt)

	// Convertir test results
	testResults := make([]*entities.TestResult, 0, len(model.TestResults))
	for _, tr := range model.TestResults {
		testID, err := valueobjects.NewTestID(tr.TestID)
		if err != nil {
			return nil, err
		}

		testResults = append(testResults, entities.NewTestResult(
			testID,
			tr.TestName,
			tr.Passed,
			tr.ErrorMessage,
		))
	}
	execution.SetTestResults(testResults)

	return execution, nil
}

// toDomainList convierte una lista de modelos a dominio
func (r *PostgresExecutionAnalyticsRepository) toDomainList(models []ExecutionAnalyticsModel) ([]*aggregates.ExecutionAnalytics, error) {
	results := make([]*aggregates.ExecutionAnalytics, 0, len(models))
	for i := range models {
		domain, err := r.toDomain(&models[i])
		if err != nil {
			return nil, err
		}
		results = append(results, domain)
	}
	return results, nil
}
