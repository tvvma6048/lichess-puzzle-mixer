package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const schemaVersion = "1"

type DB struct {
	sql *sql.DB
}

type Status struct {
	Ready         bool
	PuzzleCount   int64
	ImportedAt    string
	SchemaVersion int
}

func Open(path string) (*DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", path)
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	db := &DB{sql: sqlDB}
	if err := db.migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	_, _ = db.sql.Exec(`PRAGMA cache_size = -64000`)
	return db, nil
}

func (db *DB) Close() error {
	return db.sql.Close()
}

func (db *DB) migrate() error {
	schema, err := loadSchemaSQL()
	if err != nil {
		return err
	}
	if _, err := db.sql.Exec(schema); err != nil {
		return err
	}
	return db.setMeta("schema_version", schemaVersion)
}

func (db *DB) setMeta(key, value string) error {
	_, err := db.sql.Exec(
		`INSERT INTO meta(key, value) VALUES(?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}

func (db *DB) getMeta(key string) (string, bool, error) {
	var value string
	err := db.sql.QueryRow(`SELECT value FROM meta WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (db *DB) Status() (Status, error) {
	count, err := db.PuzzleCount()
	if err != nil {
		return Status{}, err
	}
	st := Status{
		Ready:         count > 0,
		PuzzleCount:   count,
		SchemaVersion: 1,
	}
	if imported, ok, err := db.getMeta("imported_at"); err != nil {
		return Status{}, err
	} else if ok {
		st.ImportedAt = imported
	}
	return st, nil
}

func (db *DB) PuzzleCount() (int64, error) {
	var count int64
	err := db.sql.QueryRow(`SELECT COUNT(*) FROM puzzles`).Scan(&count)
	return count, err
}

func (db *DB) ThemesForPuzzle(puzzleID string) ([]string, error) {
	rows, err := db.sql.Query(
		`SELECT theme FROM puzzle_themes WHERE puzzle_id = ? ORDER BY theme`,
		puzzleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var themes []string
	for rows.Next() {
		var theme string
		if err := rows.Scan(&theme); err != nil {
			return nil, err
		}
		themes = append(themes, theme)
	}
	return themes, rows.Err()
}

// Wipe removes all puzzles and import metadata.
func (db *DB) Wipe(ctx context.Context) error {
	if err := db.resetForImport(); err != nil {
		return err
	}
	_, _ = db.sql.ExecContext(ctx, `PRAGMA journal_mode = WAL`)
	_, _ = db.sql.ExecContext(ctx, `PRAGMA synchronous = NORMAL`)
	return nil
}

func (db *DB) resetForImport() error {
	_, err := db.sql.Exec(`
		DELETE FROM puzzle_themes;
		DELETE FROM puzzles;
		DELETE FROM meta WHERE key IN ('imported_at', 'puzzle_count', 'lichess_csv_etag');
	`)
	return err
}

func (db *DB) finishImport(count int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if err := db.setMeta("imported_at", now); err != nil {
		return err
	}
	return db.setMeta("puzzle_count", fmt.Sprintf("%d", count))
}
