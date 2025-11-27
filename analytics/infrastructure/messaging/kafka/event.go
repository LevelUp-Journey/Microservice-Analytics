package kafka

import (
	"time"
)

// ExecutionAnalyticsEvent representa el evento recibido de Kafka
type ExecutionAnalyticsEvent struct {
	ExecutionID     string            `json:"execution_id"`
	ChallengeID     string            `json:"challenge_id"`
	CodeVersionID   string            `json:"code_version_id"`
	StudentID       string            `json:"student_id"`
	Language        string            `json:"language"`
	Status          string            `json:"status"`
	Timestamp       time.Time         `json:"timestamp"`
	ExecutionTimeMs int64             `json:"execution_time_ms"`
	ExitCode        int               `json:"exit_code"`
	TotalTests      int               `json:"total_tests"`
	PassedTests     int               `json:"passed_tests"`
	FailedTests     int               `json:"failed_tests"`
	Success         bool              `json:"success"`
	TestResults     []TestResultEvent `json:"test_results"`
	ServerInstance  string            `json:"server_instance"`
}

// TestResultEvent representa un resultado de test en el evento
type TestResultEvent struct {
	TestID       string `json:"test_id"`
	TestName     string `json:"test_name"`
	Passed       bool   `json:"passed"`
	ErrorMessage string `json:"error_message"`
}
