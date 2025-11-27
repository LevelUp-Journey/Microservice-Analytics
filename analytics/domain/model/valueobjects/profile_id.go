package valueobjects

import (
	"errors"
	"github.com/google/uuid"
)

// ProfileID representa el identificador único de un perfil en la comunidad
type ProfileID struct {
	value string
}

// NewProfileID crea un nuevo ProfileID validado
func NewProfileID(value string) (ProfileID, error) {
	if value == "" {
		return ProfileID{}, errors.New("profile ID cannot be empty")
	}

	// Validar que sea un UUID válido
	if _, err := uuid.Parse(value); err != nil {
		return ProfileID{}, errors.New("profile ID must be a valid UUID")
	}

	return ProfileID{value: value}, nil
}

// Value retorna el valor del ProfileID
func (p ProfileID) Value() string {
	return p.value
}

// String implementa fmt.Stringer
func (p ProfileID) String() string {
	return p.value
}

// Equals compara dos ProfileIDs
func (p ProfileID) Equals(other ProfileID) bool {
	return p.value == other.value
}
