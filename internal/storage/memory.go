package storage

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type InMemoryStorage struct {
	db *sql.DB
}

func NewInMemoryStorage() (*InMemoryStorage, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	s := &InMemoryStorage{
		db: db,
	}
	if err := s.setup(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *InMemoryStorage) GetConnection() (*sql.DB, error) {
	return s.db, nil
}

func (s *InMemoryStorage) setup() error {
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS analysis (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL DEFAULT '',
		days_between_water INTEGER NOT NULL DEFAULT 0,
		days_to_maturity INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'pending',
		error TEXT NOT NULL DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`)
	return err
}

func (s *InMemoryStorage) CreateAnalysis() (int, error) {
	result, err := s.db.Exec("INSERT INTO analysis (status) VALUES ('pending')")
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s *InMemoryStorage) UpdateAnalysis(id int, name string, daysBetweenWater, daysToMaturity int) error {
	_, err := s.db.Exec(
		"UPDATE analysis SET name = ?, days_between_water = ?, days_to_maturity = ?, status = 'completed', updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		name, daysBetweenWater, daysToMaturity, id,
	)
	return err
}

func (s *InMemoryStorage) UpdateAnalysisStatus(id int, status string, errMsg string) error {
	_, err := s.db.Exec(
		"UPDATE analysis SET status = ?, error = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, errMsg, id,
	)
	return err
}

func (s *InMemoryStorage) GetAnalysis(id int) (*Analysis, error) {
	var analysis Analysis
	err := s.db.QueryRow(
		"SELECT id, name, days_between_water, days_to_maturity, status, error FROM analysis WHERE id = ?",
		id,
	).Scan(&analysis.ID, &analysis.Name, &analysis.DaysBetweenWater, &analysis.DaysToMaturity, &analysis.Status, &analysis.Error)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &analysis, nil
}
