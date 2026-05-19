package db

import (
	"context"
	"path/filepath"
	"testing"
)

func TestImportSampleCSV(t *testing.T) {
	csvPath := filepath.Join("..", "..", "testdata", "lichess_sample.csv")
	dbPath := filepath.Join(t.TempDir(), "puzzles.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer database.Close()

	result, err := database.ImportCSV(context.Background(), csvPath)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if result.RowsImported < 100 {
		t.Fatalf("rows imported = %d, want >= 100", result.RowsImported)
	}

	st, err := database.Status()
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !st.Ready {
		t.Fatal("expected db ready")
	}
	if st.PuzzleCount != result.RowsImported {
		t.Fatalf("puzzle count %d != imported %d", st.PuzzleCount, result.RowsImported)
	}
}

func TestNextPuzzleORFilter(t *testing.T) {
	csvPath := filepath.Join("..", "..", "testdata", "lichess_sample.csv")
	database := openWithSample(t, csvPath)
	defer database.Close()

	p, err := database.NextPuzzle(context.Background(), Filter{
		Themes:        []string{"fork"},
		ThemesMode:    "or",
		RatingMin:     0,
		RatingMax:     4000,
		PopularityMin: -100,
		LengthMin:     0,
		LengthMax:     99,
		SideToMove:    "any",
	})
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if p == nil {
		t.Fatal("expected a puzzle matching fork")
	}
	hasFork := false
	for _, th := range p.Themes {
		if th == "fork" {
			hasFork = true
		}
	}
	if !hasFork {
		t.Fatalf("puzzle themes %v missing fork", p.Themes)
	}
}

func openWithSample(t *testing.T, csvPath string) *DB {
	t.Helper()
	database, err := Open(filepath.Join(t.TempDir(), "puzzles.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, err := database.ImportCSV(context.Background(), csvPath); err != nil {
		t.Fatalf("import: %v", err)
	}
	return database
}
