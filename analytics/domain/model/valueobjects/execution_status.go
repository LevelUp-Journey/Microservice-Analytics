package valueobjects

import "errors"

// ExecutionStatus representa el estado de una ejecución
type ExecutionStatus string

const (
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusTimeout   ExecutionStatus = "timeout"
	StatusError     ExecutionStatus = "error"
)

// NewExecutionStatus crea y valida un ExecutionStatus
func NewExecutionStatus(value string) (ExecutionStatus, error) {
	status := ExecutionStatus(value)

	switch status {
	case StatusCompleted, StatusFailed, StatusTimeout, StatusError:
		return status, nil
	default:
		return "", errors.New("invalid execution status")
	}
}

// String implementa Stringer
func (e ExecutionStatus) String() string {
	return string(e)
}

// Value retorna el valor del ExecutionStatus
func (e ExecutionStatus) Value() string {
	return string(e)
}

// IsSuccessful indica si el estado representa una ejecución exitosa
func (e ExecutionStatus) IsSuccessful() bool {
	return e == StatusCompleted
}
