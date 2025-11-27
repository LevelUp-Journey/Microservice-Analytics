package valueobjects

import (
	"errors"

	"github.com/google/uuid"
)

// TestID representa el identificador único de un test
type TestID struct {
	value string
}

// NewTestID crea un nuevo TestID validando que sea un UUID válido
func NewTestID(value string) (TestID, error) {
	if value == "" {
		return TestID{}, errors.New("test ID cannot be empty")
	}

	if _, err := uuid.Parse(value); err != nil {
		return TestID{}, errors.New("invalid test ID format: must be a valid UUID")
	}

	return TestID{value: value}, nil
}

// Value retorna el valor del TestID
func (t TestID) Value() string {
	return t.value
}

// Equals compara dos TestIDs
func (t TestID) Equals(other TestID) bool {
	return t.value == other.value
}

// String implementa Stringer
func (t TestID) String() string {
	return t.value
}
