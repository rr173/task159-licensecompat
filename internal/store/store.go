package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db  *sql.DB
	now func() time.Time
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db, now: time.Now}
	if err := s.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error { return s.db.Close() }
func (s *Store) DB() *sql.DB  { return s.db }

func (s *Store) migrate(ctx context.Context) error {
	stmts := []string{
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE IF NOT EXISTS policies (id TEXT PRIMARY KEY, name TEXT NOT NULL, version INTEGER NOT NULL, status TEXT NOT NULL, payload TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS policies_name_version ON policies(name, version)`,
		`CREATE TABLE IF NOT EXISTS submissions (id TEXT PRIMARY KEY, name TEXT NOT NULL, release TEXT NOT NULL, fingerprint TEXT NOT NULL UNIQUE, payload TEXT NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS components (submission_id TEXT NOT NULL, component_id TEXT NOT NULL, license TEXT NOT NULL, status TEXT NOT NULL, payload TEXT NOT NULL, PRIMARY KEY(submission_id,component_id), FOREIGN KEY(submission_id) REFERENCES submissions(id))`,
		`CREATE TABLE IF NOT EXISTS edges (submission_id TEXT NOT NULL, source_id TEXT NOT NULL, target_id TEXT NOT NULL, scope TEXT NOT NULL, PRIMARY KEY(submission_id,source_id,target_id,scope), FOREIGN KEY(submission_id) REFERENCES submissions(id))`,
		`CREATE TABLE IF NOT EXISTS analyses (id TEXT PRIMARY KEY, submission_id TEXT NOT NULL, policy_id TEXT NOT NULL, status TEXT NOT NULL, input_snapshot TEXT NOT NULL, policy_snapshot TEXT NOT NULL, published_at TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL, FOREIGN KEY(submission_id) REFERENCES submissions(id), FOREIGN KEY(policy_id) REFERENCES policies(id))`,
		`CREATE INDEX IF NOT EXISTS analyses_submission ON analyses(submission_id, created_at)`,
		`CREATE TABLE IF NOT EXISTS findings (id TEXT PRIMARY KEY, analysis_id TEXT NOT NULL, kind TEXT NOT NULL, code TEXT NOT NULL, payload TEXT NOT NULL, created_at TEXT NOT NULL, FOREIGN KEY(analysis_id) REFERENCES analyses(id))`,
		`CREATE TABLE IF NOT EXISTS waivers (id TEXT PRIMARY KEY, analysis_id TEXT NOT NULL, finding_id TEXT NOT NULL, status TEXT NOT NULL, payload TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL, FOREIGN KEY(analysis_id) REFERENCES analyses(id), FOREIGN KEY(finding_id) REFERENCES findings(id))`,
		`CREATE TABLE IF NOT EXISTS audit_events (id INTEGER PRIMARY KEY AUTOINCREMENT, subject_type TEXT NOT NULL, subject_id TEXT NOT NULL, action TEXT NOT NULL, payload TEXT NOT NULL, created_at TEXT NOT NULL)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate %q: %w", stmt, err)
		}
	}
	return nil
}

func (s *Store) transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
func timestamp(value time.Time) string          { return value.UTC().Format(time.RFC3339Nano) }
func parseTime(value string) (time.Time, error) { return time.Parse(time.RFC3339Nano, value) }
