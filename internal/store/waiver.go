package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rr173/task159-licensecompat/internal/model"
)

func (s *Store) SaveWaiver(ctx context.Context, value model.Waiver) error {
	var findingPayload string
	if err := s.db.QueryRowContext(ctx, `SELECT payload FROM findings WHERE id=?`, value.FindingID).Scan(&findingPayload); err != nil {
		return fmt.Errorf("%w: finding", model.ErrNotFound)
	}
	var finding model.Finding
	if err := json.Unmarshal([]byte(findingPayload), &finding); err != nil {
		return err
	}
	if !model.CanWaiveFinding(finding.Kind) {
		return fmt.Errorf("%w: only blocker findings can be waived", model.ErrInvalidState)
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO waivers(id,analysis_id,finding_id,status,payload,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, value.ID, value.AnalysisID, value.FindingID, value.Status, string(payload), timestamp(value.CreatedAt), timestamp(value.UpdatedAt))
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, "waiver", value.ID, "requested", string(payload))
	})
}
func (s *Store) Waiver(ctx context.Context, id string) (model.Waiver, error) {
	var payload string
	err := s.db.QueryRowContext(ctx, `SELECT payload FROM waivers WHERE id=?`, id).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Waiver{}, fmt.Errorf("%w: waiver", model.ErrNotFound)
	}
	if err != nil {
		return model.Waiver{}, err
	}
	var item model.Waiver
	err = json.Unmarshal([]byte(payload), &item)
	return item, err
}
func (s *Store) Waivers(ctx context.Context, analysisID string) ([]model.Waiver, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT payload FROM waivers WHERE analysis_id=? ORDER BY created_at`, analysisID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Waiver{}
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var item model.Waiver
		if err := json.Unmarshal([]byte(payload), &item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
func (s *Store) UpdateWaiver(ctx context.Context, value model.Waiver) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.transaction(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE waivers SET status=?,payload=?,updated_at=? WHERE id=?`, value.Status, string(payload), timestamp(value.UpdatedAt), value.ID)
		if err != nil {
			return err
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			return fmt.Errorf("%w: waiver", model.ErrNotFound)
		}
		return s.audit(ctx, tx, "waiver", value.ID, "status:"+string(value.Status), string(payload))
	})
}
