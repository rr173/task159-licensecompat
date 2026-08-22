package store

import (
	"context"
	"database/sql"
	"time"
)

type AuditEvent struct {
	ID          int64     `json:"id"`
	SubjectType string    `json:"subject_type"`
	SubjectID   string    `json:"subject_id"`
	Action      string    `json:"action"`
	Payload     string    `json:"payload"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *Store) audit(ctx context.Context, tx *sql.Tx, subjectType, subjectID, action, payload string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO audit_events(subject_type,subject_id,action,payload,created_at) VALUES(?,?,?,?,?)`, subjectType, subjectID, action, payload, timestamp(s.now()))
	return err
}
func (s *Store) Audit(ctx context.Context, subjectType, subjectID string) ([]AuditEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,subject_type,subject_id,action,payload,created_at FROM audit_events WHERE subject_type=? AND subject_id=? ORDER BY id`, subjectType, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AuditEvent{}
	for rows.Next() {
		var event AuditEvent
		var created string
		if err := rows.Scan(&event.ID, &event.SubjectType, &event.SubjectID, &event.Action, &event.Payload, &created); err != nil {
			return nil, err
		}
		event.CreatedAt, err = parseTime(created)
		if err != nil {
			return nil, err
		}
		out = append(out, event)
	}
	return out, rows.Err()
}
