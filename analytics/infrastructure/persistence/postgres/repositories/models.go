package repositories

import (
	"time"
)

// ExecutionAnalyticsModel es el modelo GORM para persistencia
type ExecutionAnalyticsModel struct {
	ID              uint              `gorm:"primaryKey"`
	ExecutionID     string            `gorm:"uniqueIndex;not null;type:uuid"`
	ChallengeID     string            `gorm:"index;not null;type:uuid"`
	CodeVersionID   string            `gorm:"type:uuid"`
	StudentID       string            `gorm:"index;not null;type:uuid"`
	Language        string            `gorm:"index;not null"`
	Status          string            `gorm:"not null"`
	Timestamp       time.Time         `gorm:"index;not null"`
	ExecutionTimeMs int64             `gorm:"not null"`
	ExitCode        int               `gorm:"not null"`
	TotalTests      int               `gorm:"not null"`
	PassedTests     int               `gorm:"not null"`
	FailedTests     int               `gorm:"not null"`
	Success         bool              `gorm:"index;not null"`
	ServerInstance  string            `gorm:"not null"`
	CreatedAt       time.Time         `gorm:"autoCreateTime"`
	UpdatedAt       time.Time         `gorm:"autoUpdateTime"`
	TestResults     []TestResultModel `gorm:"foreignKey:ExecutionAnalyticsID;constraint:OnDelete:CASCADE"`
}

// TableName especifica el nombre de la tabla
func (ExecutionAnalyticsModel) TableName() string {
	return "execution_analytics"
}

// TestResultModel es el modelo GORM para resultados de tests
type TestResultModel struct {
	ID                   uint      `gorm:"primaryKey"`
	ExecutionAnalyticsID uint      `gorm:"index;not null"`
	TestID               string    `gorm:"not null;type:uuid"`
	TestName             string    `gorm:"not null"`
	Passed               bool      `gorm:"not null"`
	ErrorMessage         string    `gorm:"type:text"`
	CreatedAt            time.Time `gorm:"autoCreateTime"`
}

// TableName especifica el nombre de la tabla
func (TestResultModel) TableName() string {
	return "test_results"
}
