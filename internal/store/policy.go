package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rr173/task159-licensecompat/internal/model"
)

func (s *Store) SavePolicy(ctx context.Context, policy model.Policy) error {
	payload, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	return s.transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO policies(id,name,version,status,payload,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, policy.ID, policy.Name, policy.Version, policy.Status, string(payload), timestamp(policy.CreatedAt), timestamp(policy.UpdatedAt))
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, "policy", policy.ID, "created", string(payload))
	})
}
func (s *Store) Policy(ctx context.Context, id string) (model.Policy, error) {
	var payload string
	err := s.db.QueryRowContext(ctx, `SELECT payload FROM policies WHERE id=?`, id).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Policy{}, fmt.Errorf("%w: policy", model.ErrNotFound)
	}
	if err != nil {
		return model.Policy{}, err
	}
	var policy model.Policy
	err = json.Unmarshal([]byte(payload), &policy)
	return policy, err
}
func (s *Store) ActivePolicy(ctx context.Context) (model.Policy, error) {
	var payload string
	err := s.db.QueryRowContext(ctx, `SELECT payload FROM policies WHERE status='active' ORDER BY updated_at DESC LIMIT 1`).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Policy{}, fmt.Errorf("%w: active policy", model.ErrNotFound)
	}
	if err != nil {
		return model.Policy{}, err
	}
	var policy model.Policy
	err = json.Unmarshal([]byte(payload), &policy)
	return policy, err
}
func (s *Store) Policies(ctx context.Context) ([]model.Policy, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT payload FROM policies ORDER BY name,version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Policy{}
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var item model.Policy
		if err := json.Unmarshal([]byte(payload), &item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
func (s *Store) ActivatePolicy(ctx context.Context, id string, updated model.Policy) error {
	payload, err := json.Marshal(updated)
	if err != nil {
		return err
	}
	return s.transaction(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `UPDATE policies SET status='retired',updated_at=? WHERE status='active'`, timestamp(updated.UpdatedAt)); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `UPDATE policies SET status=?,payload=?,updated_at=? WHERE id=?`, updated.Status, string(payload), timestamp(updated.UpdatedAt), id)
		if err != nil {
			return err
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			return fmt.Errorf("%w: policy", model.ErrNotFound)
		}
		return s.audit(ctx, tx, "policy", id, "activated", string(payload))
	})
}
