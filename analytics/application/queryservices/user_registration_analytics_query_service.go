package queryservices

import (
	"analytics/analytics/domain/model/aggregates"
	"analytics/analytics/domain/model/valueobjects"
	"analytics/analytics/domain/repositories"
	"context"
	"time"
)

// UserRegistrationAnalyticsQueryService maneja las consultas de registro de usuarios
type UserRegistrationAnalyticsQueryService struct {
	repository repositories.UserRegistrationAnalyticsRepository
}

// NewUserRegistrationAnalyticsQueryService crea una nueva instancia del servicio
func NewUserRegistrationAnalyticsQueryService(repository repositories.UserRegistrationAnalyticsRepository) *UserRegistrationAnalyticsQueryService {
	return &UserRegistrationAnalyticsQueryService{
		repository: repository,
	}
}

// GetByUserID obtiene un registro por ID de usuario
func (s *UserRegistrationAnalyticsQueryService) GetByUserID(ctx context.Context, userID valueobjects.UserID) (*aggregates.UserRegistrationAnalytics, error) {
	return s.repository.FindByUserID(ctx, userID)
}

// GetByEmail obtiene un registro por email
func (s *UserRegistrationAnalyticsQueryService) GetByEmail(ctx context.Context, email valueobjects.Email) (*aggregates.UserRegistrationAnalytics, error) {
	return s.repository.FindByEmail(ctx, email)
}

// GetByProvider obtiene todos los registros de un proveedor
func (s *UserRegistrationAnalyticsQueryService) GetByProvider(ctx context.Context, provider valueobjects.Provider, limit, offset int) ([]*aggregates.UserRegistrationAnalytics, error) {
	return s.repository.FindByProvider(ctx, provider, limit, offset)
}

// GetByDateRange obtiene registros en un rango de fechas
func (s *UserRegistrationAnalyticsQueryService) GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*aggregates.UserRegistrationAnalytics, error) {
	return s.repository.FindByDateRange(ctx, startDate, endDate, limit, offset)
}

// GetAll obtiene todos los registros con paginación
func (s *UserRegistrationAnalyticsQueryService) GetAll(ctx context.Context, limit, offset int) ([]*aggregates.UserRegistrationAnalytics, error) {
	return s.repository.FindAll(ctx, limit, offset)
}

// GetTotalUsers obtiene el total de usuarios registrados
func (s *UserRegistrationAnalyticsQueryService) GetTotalUsers(ctx context.Context) (int64, error) {
	return s.repository.CountTotal(ctx)
}

// GetProviderStats obtiene estadísticas por proveedor
func (s *UserRegistrationAnalyticsQueryService) GetProviderStats(ctx context.Context) ([]repositories.ProviderStats, error) {
	return s.repository.GetProviderStats(ctx)
}

// GetDailyRegistrationStats obtiene estadísticas diarias
func (s *UserRegistrationAnalyticsQueryService) GetDailyRegistrationStats(ctx context.Context, startDate, endDate time.Time) ([]repositories.DailyRegistrationStats, error) {
	return s.repository.GetDailyRegistrationStats(ctx, startDate, endDate)
}

// GetTopEmailDomains obtiene los dominios de email más usados
func (s *UserRegistrationAnalyticsQueryService) GetTopEmailDomains(ctx context.Context, limit int) ([]repositories.EmailDomainStats, error) {
	return s.repository.GetTopEmailDomains(ctx, limit)
}
