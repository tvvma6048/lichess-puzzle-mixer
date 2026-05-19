package db

import (
	"compress/bzip2"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

)

type ImportResult struct {
	RowsImported int64
}

// ImportCSV loads puzzles from a Lichess puzzle export (plain CSV).
func (db *DB) ImportCSV(ctx context.Context, path string) (ImportResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return ImportResult{}, err
	}
	defer f.Close()
	return db.ImportReader(ctx, f, path)
}

// ImportReader loads puzzles from a Lichess CSV stream (optionally compressed).
func (db *DB) ImportReader(ctx context.Context, r io.Reader, filename string) (ImportResult, error) {
	r = decompressByName(r, filename)
	return db.Import(ctx, r)
}

func decompressByName(r io.Reader, filename string) io.Reader {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".bz2"):
		return bzip2.NewReader(r)
	default:
		return r
	}
}

func (db *DB) Import(ctx context.Context, r io.Reader) (ImportResult, error) {
	if err := db.resetForImport(); err != nil {
		return ImportResult{}, err
	}

	_, _ = db.sql.Exec(`PRAGMA journal_mode = OFF`)
	_, _ = db.sql.Exec(`PRAGMA synchronous = OFF`)
	_, _ = db.sql.Exec(`PRAGMA temp_store = MEMORY`)
	_, _ = db.sql.Exec(`PRAGMA cache_size = -262144`)

	cr := csv.NewReader(r)
	cr.FieldsPerRecord = 10
	cr.ReuseRecord = true

	header, err := cr.Read()
	if err != nil {
		return ImportResult{}, fmt.Errorf("read header: %w", err)
	}
	if len(header) < 8 || header[0] != "PuzzleId" {
		return ImportResult{}, fmt.Errorf("unexpected CSV header: %v", header)
	}

	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return ImportResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	puzzleStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO puzzles (
			puzzle_id, fen, moves, rating, rating_deviation, popularity, nb_plays,
			game_url, opening_family, opening_variation, side_to_move, solution_length
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return ImportResult{}, err
	}
	defer puzzleStmt.Close()

	themeStmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO puzzle_themes (puzzle_id, theme) VALUES (?, ?)`)
	if err != nil {
		return ImportResult{}, err
	}
	defer themeStmt.Close()

	var imported int64
	for {
		if err := ctx.Err(); err != nil {
			return ImportResult{}, err
		}
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ImportResult{}, fmt.Errorf("read row %d: %w", imported+1, err)
		}
		if len(record) < 8 {
			continue
		}

		puzzleID := record[0]
		fen := record[1]
		moves := record[2]
		rating, _ := strconv.Atoi(record[3])
		ratingDev, _ := strconv.Atoi(record[4])
		popularity, _ := strconv.Atoi(record[5])
		nbPlays, _ := strconv.Atoi(record[6])
		themesRaw := record[7]
		gameURL := ""
		if len(record) > 8 {
			gameURL = record[8]
		}
		openingRaw := ""
		if len(record) > 9 {
			openingRaw = record[9]
		}

		family, variation := parseOpeningTags(openingRaw)
		side := solverSide(fen)
		length := solutionLength(moves)

		_, err = puzzleStmt.ExecContext(ctx,
			puzzleID, fen, moves, rating, ratingDev, popularity, nbPlays,
			nullIfEmpty(gameURL), family, variation, side, length,
		)
		if err != nil {
			return ImportResult{}, fmt.Errorf("insert puzzle %s: %w", puzzleID, err)
		}
		for _, theme := range splitThemes(themesRaw) {
			if _, err := themeStmt.ExecContext(ctx, puzzleID, theme); err != nil {
				return ImportResult{}, fmt.Errorf("insert theme %s/%s: %w", puzzleID, theme, err)
			}
		}
		imported++
	}

	if err := tx.Commit(); err != nil {
		return ImportResult{}, err
	}

	_, _ = db.sql.Exec(`PRAGMA journal_mode = WAL`)
	_, _ = db.sql.Exec(`PRAGMA synchronous = NORMAL`)
	_, _ = db.sql.Exec(`ANALYZE`)

	if err := db.finishImport(imported); err != nil {
		return ImportResult{}, err
	}

	return ImportResult{RowsImported: imported}, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
