package entities

import (
	"github.com/nanab/analytics-service/analytics/domain/model/valueobjects"
)

// TestResult representa el resultado de un test individual
type TestResult struct {
	testID       valueobjects.TestID
	testName     string
	passed       bool
	errorMessage string
}

// NewTestResult crea una nueva instancia de TestResult
func NewTestResult(
	testID valueobjects.TestID,
	testName string,
	passed bool,
	errorMessage string,
) *TestResult {
	return &TestResult{
		testID:       testID,
		testName:     testName,
		passed:       passed,
		errorMessage: errorMessage,
	}
}

// TestID retorna el ID del test
func (t *TestResult) TestID() valueobjects.TestID {
	return t.testID
}

// TestName retorna el nombre del test
func (t *TestResult) TestName() string {
	return t.testName
}

// Passed indica si el test pasó
func (t *TestResult) Passed() bool {
	return t.passed
}

// ErrorMessage retorna el mensaje de error si el test falló
func (t *TestResult) ErrorMessage() string {
	return t.errorMessage
}

// HasError indica si el test tiene un mensaje de error
func (t *TestResult) HasError() bool {
	return t.errorMessage != ""
}
