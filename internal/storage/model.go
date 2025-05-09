package storage

import "database/sql"

type Analysis struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	DaysBetweenWater int    `json:"days_between_water"`
	DaysToMaturity   int    `json:"days_to_maturity"`
	Status           string `json:"status"` // "pending", "completed", "failed"
	Error            string `json:"error,omitempty"`
}

type Storage interface {
	GetConnection() (*sql.DB, error)
	CreateAnalysis() (int, error)
	UpdateAnalysis(id int, name string, daysBetweenWater, daysToMaturity int) error
	UpdateAnalysisStatus(id int, status string, errMsg string) error
	GetAnalysis(id int) (*Analysis, error)
	setup() error
}
