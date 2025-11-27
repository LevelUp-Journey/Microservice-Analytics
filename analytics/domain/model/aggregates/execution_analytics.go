package aggregates

import (
	"analytics/analytics/domain/model/entities"
	"analytics/analytics/domain/model/valueobjects"
	"errors"
	"time"
)

// ExecutionAnalytics es el aggregate root que representa el análisis de una ejecución
type ExecutionAnalytics struct {
	id              uint
	executionID     valueobjects.ExecutionID
	challengeID     valueobjects.ChallengeID
	codeVersionID   string
	studentID       valueobjects.StudentID
	language        valueobjects.ProgrammingLanguage
	status          valueobjects.ExecutionStatus
	timestamp       time.Time
	executionTimeMs int64
	exitCode        int
	totalTests      int
	passedTests     int
	failedTests     int
	success         bool
	serverInstance  string
	testResults     []*entities.TestResult
	createdAt       time.Time
	updatedAt       time.Time
}

// NewExecutionAnalytics crea un nuevo aggregate de ExecutionAnalytics
func NewExecutionAnalytics(
	executionID valueobjects.ExecutionID,
	challengeID valueobjects.ChallengeID,
	codeVersionID string,
	studentID valueobjects.StudentID,
	language valueobjects.ProgrammingLanguage,
	status valueobjects.ExecutionStatus,
	timestamp time.Time,
	executionTimeMs int64,
	exitCode int,
	totalTests int,
	passedTests int,
	failedTests int,
	success bool,
	serverInstance string,
) (*ExecutionAnalytics, error) {
	// Validaciones de negocio
	if executionTimeMs < 0 {
		return nil, errors.New("execution time cannot be negative")
	}

	if totalTests < 0 || passedTests < 0 || failedTests < 0 {
		return nil, errors.New("test counts cannot be negative")
	}

	if passedTests+failedTests != totalTests {
		return nil, errors.New("passed + failed tests must equal total tests")
	}

	now := time.Now()
	return &ExecutionAnalytics{
		executionID:     executionID,
		challengeID:     challengeID,
		codeVersionID:   codeVersionID,
		studentID:       studentID,
		language:        language,
		status:          status,
		timestamp:       timestamp,
		executionTimeMs: executionTimeMs,
		exitCode:        exitCode,
		totalTests:      totalTests,
		passedTests:     passedTests,
		failedTests:     failedTests,
		success:         success,
		serverInstance:  serverInstance,
		testResults:     make([]*entities.TestResult, 0),
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

// AddTestResult agrega un resultado de test al aggregate
func (e *ExecutionAnalytics) AddTestResult(testResult *entities.TestResult) {
	e.testResults = append(e.testResults, testResult)
	e.updatedAt = time.Now()
}

// SetTestResults establece todos los resultados de tests
func (e *ExecutionAnalytics) SetTestResults(testResults []*entities.TestResult) {
	e.testResults = testResults
	e.updatedAt = time.Now()
}

// CalculateSuccessRate calcula el porcentaje de éxito
func (e *ExecutionAnalytics) CalculateSuccessRate() float64 {
	if e.totalTests == 0 {
		return 0.0
	}
	return (float64(e.passedTests) / float64(e.totalTests)) * 100.0
}

// IsSlowExecution indica si la ejecución fue lenta (>5000ms)
func (e *ExecutionAnalytics) IsSlowExecution() bool {
	return e.executionTimeMs > 5000
}

// HasTestFailures indica si hubo tests fallidos
func (e *ExecutionAnalytics) HasTestFailures() bool {
	return e.failedTests > 0
}

// Getters
func (e *ExecutionAnalytics) ID() uint {
	return e.id
}

func (e *ExecutionAnalytics) ExecutionID() valueobjects.ExecutionID {
	return e.executionID
}

func (e *ExecutionAnalytics) ChallengeID() valueobjects.ChallengeID {
	return e.challengeID
}

func (e *ExecutionAnalytics) CodeVersionID() string {
	return e.codeVersionID
}

func (e *ExecutionAnalytics) StudentID() valueobjects.StudentID {
	return e.studentID
}

func (e *ExecutionAnalytics) Language() valueobjects.ProgrammingLanguage {
	return e.language
}

func (e *ExecutionAnalytics) Status() valueobjects.ExecutionStatus {
	return e.status
}

func (e *ExecutionAnalytics) Timestamp() time.Time {
	return e.timestamp
}

func (e *ExecutionAnalytics) ExecutionTimeMs() int64 {
	return e.executionTimeMs
}

func (e *ExecutionAnalytics) ExitCode() int {
	return e.exitCode
}

func (e *ExecutionAnalytics) TotalTests() int {
	return e.totalTests
}

func (e *ExecutionAnalytics) PassedTests() int {
	return e.passedTests
}

func (e *ExecutionAnalytics) FailedTests() int {
	return e.failedTests
}

func (e *ExecutionAnalytics) Success() bool {
	return e.success
}

func (e *ExecutionAnalytics) ServerInstance() string {
	return e.serverInstance
}

func (e *ExecutionAnalytics) TestResults() []*entities.TestResult {
	return e.testResults
}

func (e *ExecutionAnalytics) CreatedAt() time.Time {
	return e.createdAt
}

func (e *ExecutionAnalytics) UpdatedAt() time.Time {
	return e.updatedAt
}

// Setters (solo para reconstrucción desde DB)
func (e *ExecutionAnalytics) SetID(id uint) {
	e.id = id
}

func (e *ExecutionAnalytics) SetCreatedAt(t time.Time) {
	e.createdAt = t
}

func (e *ExecutionAnalytics) SetUpdatedAt(t time.Time) {
	e.updatedAt = t
}
