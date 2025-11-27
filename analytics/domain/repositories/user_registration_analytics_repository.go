package repositories

import (
	"github.com/nanab/analytics-service/analytics/domain/model/aggregates"
	"github.com/nanab/analytics-service/analytics/domain/model/valueobjects"
	"context"
	"time"
)

// UserRegistrationAnalyticsRepository define el contrato para el repositorio
type UserRegistrationAnalyticsRepository interface {
	// Save guarda o actualiza un UserRegistrationAnalytics
	Save(ctx context.Context, userReg *aggregates.UserRegistrationAnalytics) error

	// FindByUserID busca por ID de usuario
	FindByUserID(ctx context.Context, userID valueobjects.UserID) (*aggregates.UserRegistrationAnalytics, error)

	// FindByEmail busca por email
	FindByEmail(ctx context.Context, email valueobjects.Email) (*aggregates.UserRegistrationAnalytics, error)

	// FindByProvider busca todos los usuarios registrados con un proveedor
	FindByProvider(ctx context.Context, provider valueobjects.Provider, limit, offset int) ([]*aggregates.UserRegistrationAnalytics, error)

	// FindByDateRange busca registros en un rango de fechas
	FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*aggregates.UserRegistrationAnalytics, error)

	// FindAll busca todos los registros con paginación
	FindAll(ctx context.Context, limit, offset int) ([]*aggregates.UserRegistrationAnalytics, error)

	// CountByProvider cuenta registros por proveedor
	CountByProvider(ctx context.Context, provider valueobjects.Provider) (int64, error)

	// CountTotal cuenta el total de registros
	CountTotal(ctx context.Context) (int64, error)

	// GetProviderStats obtiene estadísticas por proveedor
	GetProviderStats(ctx context.Context) ([]ProviderStats, error)

	// GetDailyRegistrationStats obtiene estadísticas diarias de registros
	GetDailyRegistrationStats(ctx context.Context, startDate, endDate time.Time) ([]DailyRegistrationStats, error)

	// GetTopEmailDomains obtiene los dominios de email más usados
	GetTopEmailDomains(ctx context.Context, limit int) ([]EmailDomainStats, error)
}

// ProviderStats representa estadísticas por proveedor
type ProviderStats struct {
	Provider       string
	TotalUsers     int64
	PercentageUsed float64
}

// DailyRegistrationStats representa estadísticas diarias de registros
type DailyRegistrationStats struct {
	Date           time.Time
	TotalRegistrations int64
	OAuthRegistrations int64
	LocalRegistrations int64
}

// EmailDomainStats representa estadísticas por dominio de email
type EmailDomainStats struct {
	Domain     string
	TotalUsers int64
	Percentage float64
}
