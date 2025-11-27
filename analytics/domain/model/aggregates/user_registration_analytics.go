package aggregates

import (
	"github.com/nanab/analytics-service/analytics/domain/model/valueobjects"
	"errors"
	"time"
)

// UserRegistrationAnalytics es el aggregate root que representa el análisis de un registro de usuario en la comunidad
type UserRegistrationAnalytics struct {
	id           uint
	userID       valueobjects.UserID
	profileID    valueobjects.ProfileID
	username     string
	profileURL   *string
	registeredAt time.Time
	createdAt    time.Time
	updatedAt    time.Time
}

// NewUserRegistrationAnalytics crea un nuevo aggregate de UserRegistrationAnalytics
func NewUserRegistrationAnalytics(
	userID valueobjects.UserID,
	profileID valueobjects.ProfileID,
	username string,
	profileURL *string,
	registeredAt time.Time,
) (*UserRegistrationAnalytics, error) {
	// Validaciones de negocio
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	if registeredAt.After(time.Now()) {
		return nil, errors.New("registered date cannot be in the future")
	}

	now := time.Now()
	return &UserRegistrationAnalytics{
		userID:       userID,
		profileID:    profileID,
		username:     username,
		profileURL:   profileURL,
		registeredAt: registeredAt,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// HasProfileURL indica si el usuario tiene URL de perfil
func (u *UserRegistrationAnalytics) HasProfileURL() bool {
	return u.profileURL != nil && *u.profileURL != ""
}

// Getters
func (u *UserRegistrationAnalytics) ID() uint {
	return u.id
}

func (u *UserRegistrationAnalytics) UserID() valueobjects.UserID {
	return u.userID
}

func (u *UserRegistrationAnalytics) ProfileID() valueobjects.ProfileID {
	return u.profileID
}

func (u *UserRegistrationAnalytics) Username() string {
	return u.username
}

func (u *UserRegistrationAnalytics) ProfileURL() *string {
	return u.profileURL
}

func (u *UserRegistrationAnalytics) RegisteredAt() time.Time {
	return u.registeredAt
}

func (u *UserRegistrationAnalytics) CreatedAt() time.Time {
	return u.createdAt
}

func (u *UserRegistrationAnalytics) UpdatedAt() time.Time {
	return u.updatedAt
}

// Setters (solo para reconstrucción desde DB)
func (u *UserRegistrationAnalytics) SetID(id uint) {
	u.id = id
}

func (u *UserRegistrationAnalytics) SetCreatedAt(t time.Time) {
	u.createdAt = t
}

func (u *UserRegistrationAnalytics) SetUpdatedAt(t time.Time) {
	u.updatedAt = t
}
