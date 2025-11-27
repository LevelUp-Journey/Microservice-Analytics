package valueobjects

import (
	"errors"
	"github.com/google/uuid"
)

// UserID representa el identificador único de un usuario en el sistema IAM
type UserID struct {
	value string
}

// NewUserID crea un nuevo UserID validado
func NewUserID(value string) (UserID, error) {
	if value == "" {
		return UserID{}, errors.New("user ID cannot be empty")
	}

	// Validar que sea un UUID válido
	if _, err := uuid.Parse(value); err != nil {
		return UserID{}, errors.New("user ID must be a valid UUID")
	}

	return UserID{value: value}, nil
}

// Value retorna el valor del UserID
func (u UserID) Value() string {
	return u.value
}

// String implementa fmt.Stringer
func (u UserID) String() string {
	return u.value
}

// Equals compara dos UserIDs
func (u UserID) Equals(other UserID) bool {
	return u.value == other.value
}
