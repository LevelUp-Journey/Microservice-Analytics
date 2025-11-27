package valueobjects

import (
	"errors"

	"github.com/google/uuid"
)

// ExecutionID representa el identificador único de una ejecución
type ExecutionID struct {
	value string
}

// NewExecutionID crea un nuevo ExecutionID validando que sea un UUID válido
func NewExecutionID(value string) (ExecutionID, error) {
	if value == "" {
		return ExecutionID{}, errors.New("execution ID cannot be empty")
	}

	if _, err := uuid.Parse(value); err != nil {
		return ExecutionID{}, errors.New("invalid execution ID format: must be a valid UUID")
	}

	return ExecutionID{value: value}, nil
}

// Value retorna el valor del ExecutionID
func (e ExecutionID) Value() string {
	return e.value
}

// Equals compara dos ExecutionIDs
func (e ExecutionID) Equals(other ExecutionID) bool {
	return e.value == other.value
}

// String implementa Stringer
func (e ExecutionID) String() string {
	return e.value
}
