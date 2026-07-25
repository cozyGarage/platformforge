package progress

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const GhostHintEvery = 2

type Score struct {
	Stars             int `json:"stars"`
	Correctness       int `json:"correctness"`
	Speed             int `json:"speed"`
	Cleanliness       int `json:"cleanliness"`
	DurationSeconds   int `json:"durationSeconds"`
	FailedValidations int `json:"failedValidations"`
	HintsRevealed     int `json:"hintsRevealed"`
}

type TaskProgress struct {
	TaskID            string `json:"taskId"`
	FailedValidations int    `json:"failedValidations"`
	GhostHints        int    `json:"ghostHints"`
}

type Progress struct {
	LabID       string     `json:"labId"`
	Status      string     `json:"status"`
	Attempts    int        `json:"attempts"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Score       *Score     `json:"score,omitempty"`
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
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS progress (
			lab_id TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			attempts INTEGER NOT NULL DEFAULT 0,
			completed_at DATETIME,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS task_progress (
			lab_id TEXT NOT NULL,
			task_id TEXT NOT NULL,
			failed_validations INTEGER NOT NULL DEFAULT 0,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY(lab_id, task_id)
		)`,
		`CREATE TABLE IF NOT EXISTS lab_scores (
			lab_id TEXT PRIMARY KEY,
			stars INTEGER NOT NULL,
			correctness INTEGER NOT NULL,
			speed INTEGER NOT NULL,
			cleanliness INTEGER NOT NULL,
			duration_seconds INTEGER NOT NULL,
			failed_validations INTEGER NOT NULL,
			hints_revealed INTEGER NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
	}
	for _, stmt := range stmts {
		if _, err = db.Exec(stmt); err != nil {
			db.Close()
			return nil, err
		}
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

func (s *Store) IsCompleted(labID string) (bool, error) {
	var status string
	err := s.db.QueryRow(`SELECT status FROM progress WHERE lab_id = ?`, labID).Scan(&status)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return status == "completed", nil
}

func (s *Store) CountCompleted(labIDs []string) (int, error) {
	count := 0
	for _, id := range labIDs {
		ok, err := s.IsCompleted(id)
		if err != nil {
			return 0, err
		}
		if ok {
			count++
		}
	}
	return count, nil
}

func (s *Store) RecordFailedTasks(labID string, taskIDs []string) error {
	for _, taskID := range taskIDs {
		if taskID == "" {
			continue
		}
		_, err := s.db.Exec(`INSERT INTO task_progress(lab_id,task_id,failed_validations,updated_at)
			VALUES(?, ?, 1, CURRENT_TIMESTAMP)
			ON CONFLICT(lab_id, task_id) DO UPDATE SET
				failed_validations = failed_validations + 1,
				updated_at = CURRENT_TIMESTAMP`, labID, taskID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) TaskProgress(labID string, hintCounts map[string]int) ([]TaskProgress, error) {
	rows, err := s.db.Query(`SELECT task_id, failed_validations FROM task_progress WHERE lab_id = ? ORDER BY task_id`, labID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TaskProgress
	for rows.Next() {
		var tp TaskProgress
		if err := rows.Scan(&tp.TaskID, &tp.FailedValidations); err != nil {
			return nil, err
		}
		tp.GhostHints = GhostHintsFor(tp.FailedValidations, hintCounts[tp.TaskID])
		out = append(out, tp)
	}
	return out, rows.Err()
}

func (s *Store) TotalFailedValidations(labID string) (int, error) {
	var total sql.NullInt64
	err := s.db.QueryRow(`SELECT SUM(failed_validations) FROM task_progress WHERE lab_id = ?`, labID).Scan(&total)
	if err != nil {
		return 0, err
	}
	if !total.Valid {
		return 0, nil
	}
	return int(total.Int64), nil
}

func (s *Store) SaveScore(labID string, score Score) error {
	_, err := s.db.Exec(`INSERT INTO lab_scores(
		lab_id, stars, correctness, speed, cleanliness, duration_seconds, failed_validations, hints_revealed, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(lab_id) DO UPDATE SET
			stars=excluded.stars,
			correctness=excluded.correctness,
			speed=excluded.speed,
			cleanliness=excluded.cleanliness,
			duration_seconds=excluded.duration_seconds,
			failed_validations=excluded.failed_validations,
			hints_revealed=excluded.hints_revealed,
			updated_at=CURRENT_TIMESTAMP`,
		labID, score.Stars, score.Correctness, score.Speed, score.Cleanliness,
		score.DurationSeconds, score.FailedValidations, score.HintsRevealed)
	return err
}

func (s *Store) scoreFor(labID string) (*Score, error) {
	var sc Score
	err := s.db.QueryRow(`SELECT stars, correctness, speed, cleanliness, duration_seconds, failed_validations, hints_revealed
		FROM lab_scores WHERE lab_id = ?`, labID).Scan(
		&sc.Stars, &sc.Correctness, &sc.Speed, &sc.Cleanliness, &sc.DurationSeconds, &sc.FailedValidations, &sc.HintsRevealed)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sc, nil
}

func (s *Store) Get(labID string) (*Progress, error) {
	var p Progress
	var completed sql.NullTime
	err := s.db.QueryRow(`SELECT lab_id,status,attempts,completed_at,updated_at FROM progress WHERE lab_id = ?`, labID).
		Scan(&p.LabID, &p.Status, &p.Attempts, &completed, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if completed.Valid {
		p.CompletedAt = &completed.Time
	}
	score, err := s.scoreFor(labID)
	if err != nil {
		return nil, err
	}
	p.Score = score
	return &p, nil
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
		score, err := s.scoreFor(p.LabID)
		if err != nil {
			return nil, err
		}
		p.Score = score
		out = append(out, p)
	}
	return out, rows.Err()
}

func GhostHintsFor(failed, hintCount int) int {
	if hintCount <= 0 || failed < GhostHintEvery {
		return 0
	}
	n := failed / GhostHintEvery
	if n > hintCount {
		return hintCount
	}
	return n
}

func ComputeScore(durationSec, estimatedMinutes, failedValidations, hintsRevealed int) Score {
	correctness := 3
	speed := 1
	if estimatedMinutes <= 0 {
		speed = 2
	} else {
		par := estimatedMinutes * 60
		switch {
		case durationSec <= par:
			speed = 3
		case durationSec <= par*3/2:
			speed = 2
		}
	}
	cleanliness := 3
	if failedValidations >= 2 {
		cleanliness--
	}
	if failedValidations >= 4 {
		cleanliness--
	}
	if hintsRevealed >= 1 {
		cleanliness--
	}
	if cleanliness < 1 {
		cleanliness = 1
	}
	stars := (correctness + speed + cleanliness) / 3
	if stars < 1 {
		stars = 1
	}
	if (correctness+speed+cleanliness)%3 != 0 && stars < 3 {
		// bias upward on .67 averages so a strong pass still feels rewarding
		if correctness+speed+cleanliness >= 7 {
			stars++
		}
	}
	if stars > 3 {
		stars = 3
	}
	return Score{
		Stars:             stars,
		Correctness:       correctness,
		Speed:             speed,
		Cleanliness:       cleanliness,
		DurationSeconds:   durationSec,
		FailedValidations: failedValidations,
		HintsRevealed:     hintsRevealed,
	}
}
