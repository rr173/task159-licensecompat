package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"licensecompat.local/internal/model"
)

func (s *Store) SaveSubmission(ctx context.Context, value model.Submission) (model.Submission, bool, error) {
	for i := range value.Components {
		value.Components[i].License = model.CanonicalLicense(value.Components[i].License)
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return model.Submission{}, false, err
	}
	err = s.transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO submissions(id,name,release,fingerprint,payload,created_at) VALUES(?,?,?,?,?,?)`, value.ID, value.Name, value.Release, value.Fingerprint, string(payload), timestamp(value.CreatedAt))
		if err != nil {
			return err
		}
		for _, component := range value.Components {
			encoded, _ := json.Marshal(component)
			if _, err := tx.ExecContext(ctx, `INSERT INTO components(submission_id,component_id,license,status,payload) VALUES(?,?,?,?,?)`, value.ID, component.ID, component.License, component.Status, string(encoded)); err != nil {
				return err
			}
		}
		for _, edge := range value.Edges {
			if _, err := tx.ExecContext(ctx, `INSERT INTO edges(submission_id,source_id,target_id,scope) VALUES(?,?,?,?)`, value.ID, edge.From, edge.To, edge.Scope); err != nil {
				return err
			}
		}
		return s.audit(ctx, tx, "submission", value.ID, "created", string(payload))
	})
	if err != nil {
		if isUnique(err) {
			existing, getErr := s.SubmissionByFingerprint(ctx, value.Fingerprint)
			return existing, true, getErr
		}
		return model.Submission{}, false, err
	}
	return value, false, nil
}
func (s *Store) Submission(ctx context.Context, id string) (model.Submission, error) {
	return s.readSubmission(ctx, `SELECT payload FROM submissions WHERE id=?`, id)
}
func (s *Store) SubmissionByFingerprint(ctx context.Context, fingerprint string) (model.Submission, error) {
	return s.readSubmission(ctx, `SELECT payload FROM submissions WHERE fingerprint=?`, fingerprint)
}
func (s *Store) readSubmission(ctx context.Context, query, arg string) (model.Submission, error) {
	var payload string
	err := s.db.QueryRowContext(ctx, query, arg).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Submission{}, fmt.Errorf("%w: submission", model.ErrNotFound)
	}
	if err != nil {
		return model.Submission{}, err
	}
	var value model.Submission
	err = json.Unmarshal([]byte(payload), &value)
	return value, err
}
func (s *Store) Submissions(ctx context.Context) ([]model.Submission, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT payload FROM submissions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Submission{}
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var value model.Submission
		if err := json.Unmarshal([]byte(payload), &value); err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, rows.Err()
}
func isUnique(err error) bool {
	return err != nil && (contains(err.Error(), "UNIQUE constraint failed") || contains(err.Error(), "constraint failed"))
}
func contains(value, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
