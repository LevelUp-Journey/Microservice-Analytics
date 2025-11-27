package repositories

import (
	"github.com/nanab/analytics-service/analytics/domain/model/aggregates"
	"github.com/nanab/analytics-service/analytics/domain/model/valueobjects"
	"github.com/nanab/analytics-service/analytics/domain/repositories"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// PostgresUserRegistrationAnalyticsRepository implementa el repositorio usando PostgreSQL
type PostgresUserRegistrationAnalyticsRepository struct {
	db *gorm.DB
}

// NewPostgresUserRegistrationAnalyticsRepository crea una nueva instancia del repositorio
func NewPostgresUserRegistrationAnalyticsRepository(db *gorm.DB) repositories.UserRegistrationAnalyticsRepository {
	return &PostgresUserRegistrationAnalyticsRepository{db: db}
}

// Save guarda o actualiza un UserRegistrationAnalytics
func (r *PostgresUserRegistrationAnalyticsRepository) Save(ctx context.Context, userReg *aggregates.UserRegistrationAnalytics) error {
	model := r.toModel(userReg)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model).Error; err != nil {
			return err
		}

		// Actualizar el ID del aggregate
		userReg.SetID(model.ID)
		return nil
	})
}

// FindByUserID busca por ID de usuario
func (r *PostgresUserRegistrationAnalyticsRepository) FindByUserID(ctx context.Context, userID valueobjects.UserID) (*aggregates.UserRegistrationAnalytics, error) {
	var model UserRegistrationAnalyticsModel

	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID.Value()).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomain(&model)
}

// FindByEmail busca por email - Ya no aplica, devuelve nil
func (r *PostgresUserRegistrationAnalyticsRepository) FindByEmail(ctx context.Context, email valueobjects.Email) (*aggregates.UserRegistrationAnalytics, error) {
	return nil, nil
}

// FindByProvider busca todos los usuarios registrados con un proveedor - Ya no aplica
func (r *PostgresUserRegistrationAnalyticsRepository) FindByProvider(ctx context.Context, provider valueobjects.Provider, limit, offset int) ([]*aggregates.UserRegistrationAnalytics, error) {
	return []*aggregates.UserRegistrationAnalytics{}, nil
}

// FindByDateRange busca registros en un rango de fechas
func (r *PostgresUserRegistrationAnalyticsRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*aggregates.UserRegistrationAnalytics, error) {
	var models []UserRegistrationAnalyticsModel

	query := r.db.WithContext(ctx).
		Where("registered_at BETWEEN ? AND ?", startDate, endDate).
		Order("registered_at DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return r.toDomainList(models)
}

// FindAll busca todos los registros con paginación
func (r *PostgresUserRegistrationAnalyticsRepository) FindAll(ctx context.Context, limit, offset int) ([]*aggregates.UserRegistrationAnalytics, error) {
	var models []UserRegistrationAnalyticsModel

	query := r.db.WithContext(ctx).
		Order("registered_at DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return r.toDomainList(models)
}

// CountByProvider cuenta registros por proveedor - Ya no aplica
func (r *PostgresUserRegistrationAnalyticsRepository) CountByProvider(ctx context.Context, provider valueobjects.Provider) (int64, error) {
	return 0, nil
}

// CountTotal cuenta el total de registros
func (r *PostgresUserRegistrationAnalyticsRepository) CountTotal(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&UserRegistrationAnalyticsModel{}).
		Count(&count).Error
	return count, err
}

// GetProviderStats obtiene estadísticas por proveedor - Ya no aplica
func (r *PostgresUserRegistrationAnalyticsRepository) GetProviderStats(ctx context.Context) ([]repositories.ProviderStats, error) {
	return []repositories.ProviderStats{}, nil
}

// GetDailyRegistrationStats obtiene estadísticas diarias de registros
func (r *PostgresUserRegistrationAnalyticsRepository) GetDailyRegistrationStats(ctx context.Context, startDate, endDate time.Time) ([]repositories.DailyRegistrationStats, error) {
	var results []struct {
		Date               time.Time
		TotalRegistrations int64
	}

	err := r.db.WithContext(ctx).
		Model(&UserRegistrationAnalyticsModel{}).
		Select("DATE(registered_at) as date, COUNT(*) as total_registrations").
		Where("registered_at BETWEEN ? AND ?", startDate, endDate).
		Group("DATE(registered_at)").
		Order("date DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make([]repositories.DailyRegistrationStats, len(results))
	for i, result := range results {
		stats[i] = repositories.DailyRegistrationStats{
			Date:               result.Date,
			TotalRegistrations: result.TotalRegistrations,
			OAuthRegistrations: 0,
			LocalRegistrations: result.TotalRegistrations,
		}
	}

	return stats, nil
}

// GetTopEmailDomains obtiene los dominios de email más usados - Ya no aplica
func (r *PostgresUserRegistrationAnalyticsRepository) GetTopEmailDomains(ctx context.Context, limit int) ([]repositories.EmailDomainStats, error) {
	return []repositories.EmailDomainStats{}, nil
}

// toModel convierte un aggregate a modelo GORM
func (r *PostgresUserRegistrationAnalyticsRepository) toModel(userReg *aggregates.UserRegistrationAnalytics) *UserRegistrationAnalyticsModel {
	return &UserRegistrationAnalyticsModel{
		ID:           userReg.ID(),
		UserID:       userReg.UserID().Value(),
		ProfileID:    userReg.ProfileID().Value(),
		Username:     userReg.Username(),
		ProfileURL:   userReg.ProfileURL(),
		RegisteredAt: userReg.RegisteredAt(),
		CreatedAt:    userReg.CreatedAt(),
		UpdatedAt:    userReg.UpdatedAt(),
	}
}

// toDomain convierte un modelo GORM a aggregate
func (r *PostgresUserRegistrationAnalyticsRepository) toDomain(model *UserRegistrationAnalyticsModel) (*aggregates.UserRegistrationAnalytics, error) {
	userID, err := valueobjects.NewUserID(model.UserID)
	if err != nil {
		return nil, err
	}

	profileID, err := valueobjects.NewProfileID(model.ProfileID)
	if err != nil {
		return nil, err
	}

	userReg, err := aggregates.NewUserRegistrationAnalytics(
		userID,
		profileID,
		model.Username,
		model.ProfileURL,
		model.RegisteredAt,
	)
	if err != nil {
		return nil, err
	}

	userReg.SetID(model.ID)
	userReg.SetCreatedAt(model.CreatedAt)
	userReg.SetUpdatedAt(model.UpdatedAt)

	return userReg, nil
}

// toDomainList convierte una lista de modelos GORM a aggregates
func (r *PostgresUserRegistrationAnalyticsRepository) toDomainList(models []UserRegistrationAnalyticsModel) ([]*aggregates.UserRegistrationAnalytics, error) {
	result := make([]*aggregates.UserRegistrationAnalytics, 0, len(models))

	for _, model := range models {
		domain, err := r.toDomain(&model)
		if err != nil {
			continue
		}
		result = append(result, domain)
	}

	return result, nil
}
