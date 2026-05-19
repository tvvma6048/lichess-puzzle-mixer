package db

import (
	"context"
	"path/filepath"
	"testing"
)

func TestCountForkAndPin(t *testing.T) {
	csvPath := filepath.Join("..", "..", "testdata", "lichess_sample.csv")
	database := openWithSample(t, csvPath)
	defer database.Close()

	ctx := context.Background()

	orCount, err := database.CountPuzzles(ctx, Filter{
		Themes:        []string{"fork", "pin"},
		ThemesMode:    "or",
		PopularityMin: -100,
		LengthMax:     99,
		SideToMove:    "any",
		RatingMax:     4000,
	})
	if err != nil {
		t.Fatalf("or count: %v", err)
	}

	andCount, err := database.CountPuzzles(ctx, Filter{
		Themes:        []string{"fork", "pin"},
		ThemesMode:    "and",
		PopularityMin: -100,
		LengthMax:     99,
		SideToMove:    "any",
		RatingMax:     4000,
	})
	if err != nil {
		t.Fatalf("and count: %v", err)
	}

	if andCount != 1 {
		t.Fatalf("AND fork+pin count = %d, want 1", andCount)
	}
	if orCount <= andCount {
		t.Fatalf("OR count %d should exceed AND count %d", orCount, andCount)
	}
}

func TestThemeGroupsForkOrPinAndShort(t *testing.T) {
	csvPath := filepath.Join("..", "..", "testdata", "lichess_sample.csv")
	database := openWithSample(t, csvPath)
	defer database.Close()

	ctx := context.Background()
	base := Filter{
		PopularityMin: -100,
		LengthMax:     99,
		SideToMove:    "any",
		RatingMax:     4000,
	}

	grouped, err := database.CountPuzzles(ctx, Filter{
		ThemeGroups:   [][]string{{"fork", "pin"}, {"short"}},
		PopularityMin: base.PopularityMin,
		LengthMax:     base.LengthMax,
		SideToMove:    base.SideToMove,
		RatingMax:     base.RatingMax,
	})
	if err != nil {
		t.Fatalf("grouped count: %v", err)
	}

	forkOnly, err := database.CountPuzzles(ctx, Filter{
		ThemeGroups:   [][]string{{"fork"}, {"short"}},
		PopularityMin: base.PopularityMin,
		LengthMax:     base.LengthMax,
		SideToMove:    base.SideToMove,
		RatingMax:     base.RatingMax,
	})
	if err != nil {
		t.Fatalf("fork+short count: %v", err)
	}

	orForkPin, err := database.CountPuzzles(ctx, Filter{
		Themes:        []string{"fork", "pin"},
		ThemesMode:    "or",
		PopularityMin: base.PopularityMin,
		LengthMax:     base.LengthMax,
		SideToMove:    base.SideToMove,
		RatingMax:     base.RatingMax,
	})
	if err != nil {
		t.Fatalf("or count: %v", err)
	}

	if grouped == 0 {
		t.Fatal("(fork or pin) and short count = 0, want > 0")
	}
	if grouped >= orForkPin {
		t.Fatalf("grouped %d should be less than fork OR pin %d", grouped, orForkPin)
	}
	if grouped < forkOnly {
		t.Fatalf("grouped %d should be >= fork+short %d", grouped, forkOnly)
	}
}
