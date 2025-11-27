package valueobjects

import (
	"errors"

	"github.com/google/uuid"
)

// StudentID representa el identificador único de un estudiante
type StudentID struct {
	value string
}

// NewStudentID crea un nuevo StudentID validando que sea un UUID válido
func NewStudentID(value string) (StudentID, error) {
	if value == "" {
		return StudentID{}, errors.New("student ID cannot be empty")
	}

	if _, err := uuid.Parse(value); err != nil {
		return StudentID{}, errors.New("invalid student ID format: must be a valid UUID")
	}

	return StudentID{value: value}, nil
}

// Value retorna el valor del StudentID
func (s StudentID) Value() string {
	return s.value
}

// Equals compara dos StudentIDs
func (s StudentID) Equals(other StudentID) bool {
	return s.value == other.value
}

// String implementa Stringer
func (s StudentID) String() string {
	return s.value
}
