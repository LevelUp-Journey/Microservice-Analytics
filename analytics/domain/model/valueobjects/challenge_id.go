package valueobjects

import (
	"errors"

	"github.com/google/uuid"
)

// ChallengeID representa el identificador único de un desafío
type ChallengeID struct {
	value string
}

// NewChallengeID crea un nuevo ChallengeID validando que sea un UUID válido
func NewChallengeID(value string) (ChallengeID, error) {
	if value == "" {
		return ChallengeID{}, errors.New("challenge ID cannot be empty")
	}

	if _, err := uuid.Parse(value); err != nil {
		return ChallengeID{}, errors.New("invalid challenge ID format: must be a valid UUID")
	}

	return ChallengeID{value: value}, nil
}

// Value retorna el valor del ChallengeID
func (c ChallengeID) Value() string {
	return c.value
}

// Equals compara dos ChallengeIDs
func (c ChallengeID) Equals(other ChallengeID) bool {
	return c.value == other.value
}

// String implementa Stringer
func (c ChallengeID) String() string {
	return c.value
}
