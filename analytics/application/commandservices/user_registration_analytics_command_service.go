package commandservices

import (
	"analytics/analytics/domain/model/aggregates"
	"analytics/analytics/domain/repositories"
	"context"
	"log"
)

// UserRegistrationAnalyticsCommandService maneja los comandos de registro de usuarios
type UserRegistrationAnalyticsCommandService struct {
	repository repositories.UserRegistrationAnalyticsRepository
}

// NewUserRegistrationAnalyticsCommandService crea una nueva instancia del servicio
func NewUserRegistrationAnalyticsCommandService(repository repositories.UserRegistrationAnalyticsRepository) *UserRegistrationAnalyticsCommandService {
	return &UserRegistrationAnalyticsCommandService{
		repository: repository,
	}
}

// HandleUserRegistrationEvent procesa un evento de registro de usuario y lo guarda en analytics
func (s *UserRegistrationAnalyticsCommandService) HandleUserRegistrationEvent(ctx context.Context, userReg *aggregates.UserRegistrationAnalytics) error {
	// Verificar si el usuario ya existe
	existing, err := s.repository.FindByUserID(ctx, userReg.UserID())
	if err != nil {
		log.Printf("Error checking existing user: %v", err)
		return err
	}

	// Si ya existe, no lo guardamos de nuevo (idempotencia)
	if existing != nil {
		log.Printf("User registration already exists for user ID: %s, skipping", userReg.UserID().Value())
		return nil
	}

	// Guardar el nuevo registro
	if err := s.repository.Save(ctx, userReg); err != nil {
		log.Printf("Error saving user registration: %v", err)
		return err
	}

	log.Printf("Successfully saved user registration analytics for user ID: %s, username: %s",
		userReg.UserID().Value(), userReg.Username())
	return nil
}
