package progress

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Progress struct {
	LabID       string     `json:"labId"`
	Status      string     `json:"status"`
	Attempts    int        `json:"attempts"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS progress (
		lab_id TEXT PRIMARY KEY,
		status TEXT NOT NULL,
		attempts INTEGER NOT NULL DEFAULT 0,
		completed_at DATETIME,
		updated_at DATETIME NOT NULL
	)`); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) MarkStarted(labID string) error {
	_, err := s.db.Exec(`INSERT INTO progress(lab_id,status,attempts,updated_at)
		VALUES(?, 'in_progress', 1, CURRENT_TIMESTAMP)
		ON CONFLICT(lab_id) DO UPDATE SET status='in_progress', attempts=attempts+1, updated_at=CURRENT_TIMESTAMP`, labID)
	return err
}

func (s *Store) MarkCompleted(labID string) error {
	_, err := s.db.Exec(`INSERT INTO progress(lab_id,status,attempts,completed_at,updated_at)
		VALUES(?, 'completed', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(lab_id) DO UPDATE SET status='completed', completed_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP`, labID)
	return err
}

func (s *Store) List() ([]Progress, error) {
	rows, err := s.db.Query(`SELECT lab_id,status,attempts,completed_at,updated_at FROM progress ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Progress
	for rows.Next() {
		var p Progress
		var completed sql.NullTime
		if err := rows.Scan(&p.LabID, &p.Status, &p.Attempts, &completed, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if completed.Valid {
			p.CompletedAt = &completed.Time
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
