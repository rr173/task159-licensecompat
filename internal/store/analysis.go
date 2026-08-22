package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"licensecompat.local/internal/model"
	"time"
)

func (s *Store) SaveAnalysis(ctx context.Context, value model.Analysis) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO analyses(id,submission_id,policy_id,status,input_snapshot,policy_snapshot,published_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, value.ID, value.SubmissionID, value.PolicyID, value.Status, value.InputSnapshot, value.PolicySnapshot, nullableTime(value.PublishedAt), timestamp(value.CreatedAt), timestamp(value.UpdatedAt))
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, "analysis", value.ID, "created", value.InputSnapshot)
	})
}
func (s *Store) Analysis(ctx context.Context, id string) (model.Analysis, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,submission_id,policy_id,status,input_snapshot,policy_snapshot,published_at,created_at,updated_at FROM analyses WHERE id=?`, id)
	return scanAnalysis(row)
}
func (s *Store) AnalysesForSubmission(ctx context.Context, submissionID string) ([]model.Analysis, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,submission_id,policy_id,status,input_snapshot,policy_snapshot,published_at,created_at,updated_at FROM analyses WHERE submission_id=? ORDER BY created_at`, submissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Analysis{}
	for rows.Next() {
		item, err := scanAnalysis(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
func (s *Store) RecoverableAnalyses(ctx context.Context) ([]model.Analysis, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,submission_id,policy_id,status,input_snapshot,policy_snapshot,published_at,created_at,updated_at FROM analyses WHERE status IN ('queued','running') ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Analysis{}
	for rows.Next() {
		item, err := scanAnalysis(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
func (s *Store) UpdateAnalysis(ctx context.Context, value model.Analysis) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		var current string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM analyses WHERE id=?`, value.ID).Scan(&current); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w: analysis", model.ErrNotFound)
			}
			return err
		}
		if !model.CanTransitionAnalysis(model.AnalysisStatus(current), value.Status) && current != string(value.Status) {
			return fmt.Errorf("%w: analysis %s -> %s", model.ErrInvalidState, current, value.Status)
		}
		result, err := tx.ExecContext(ctx, `UPDATE analyses SET status=?,published_at=?,updated_at=? WHERE id=?`, value.Status, nullableTime(value.PublishedAt), timestamp(value.UpdatedAt), value.ID)
		if err != nil {
			return err
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			return fmt.Errorf("%w: analysis", model.ErrNotFound)
		}
		return s.audit(ctx, tx, "analysis", value.ID, "status:"+string(value.Status), "{}")
	})
}
func (s *Store) ReplaceFindings(ctx context.Context, analysisID string, findings []model.Finding) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM analyses WHERE id=?`, analysisID).Scan(&status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w: analysis", model.ErrNotFound)
			}
			return err
		}
		if !model.CanReplaceFindings(model.AnalysisStatus(status)) {
			return fmt.Errorf("%w: findings are frozen", model.ErrImmutable)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM findings WHERE analysis_id=?`, analysisID); err != nil {
			return err
		}
		for _, finding := range findings {
			payload, err := json.Marshal(finding)
			if err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO findings(id,analysis_id,kind,code,payload,created_at) VALUES(?,?,?,?,?,?)`, finding.ID, analysisID, finding.Kind, finding.Code, string(payload), timestamp(finding.CreatedAt)); err != nil {
				return err
			}
		}
		return s.audit(ctx, tx, "analysis", analysisID, "findings-replaced", fmt.Sprintf(`{"count":%d}`, len(findings)))
	})
}
func (s *Store) Findings(ctx context.Context, analysisID string) ([]model.Finding, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT payload FROM findings WHERE analysis_id=? ORDER BY id LIMIT 1`, analysisID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Finding{}
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var finding model.Finding
		if err := json.Unmarshal([]byte(payload), &finding); err != nil {
			return nil, err
		}
		out = append(out, finding)
	}
	return out, rows.Err()
}
func scanAnalysis(row interface{ Scan(...any) error }) (model.Analysis, error) {
	var item model.Analysis
	var published sql.NullString
	var created, updated string
	err := row.Scan(&item.ID, &item.SubmissionID, &item.PolicyID, &item.Status, &item.InputSnapshot, &item.PolicySnapshot, &published, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Analysis{}, fmt.Errorf("%w: analysis", model.ErrNotFound)
	}
	if err != nil {
		return model.Analysis{}, err
	}
	item.CreatedAt, err = parseTime(created)
	if err != nil {
		return model.Analysis{}, err
	}
	item.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return model.Analysis{}, err
	}
	if published.Valid {
		value, err := parseTime(published.String)
		if err != nil {
			return model.Analysis{}, err
		}
		item.PublishedAt = &value
	}
	return item, nil
}
func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return timestamp(*value)
}
