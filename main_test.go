package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DSerejo/lichess-puzzle-mixer/internal/server"
)

func TestEmbeddedWebAssets(t *testing.T) {
	s := server.New(server.Config{
		Dev:     false,
		DataDir: t.TempDir(),
		WebFS:   webFS,
	})

	tests := []struct {
		path       string
		wantCode   int
		wantSubstr string
	}{
		{"/", http.StatusOK, "Lichess Puzzle Mixer"},
		{"/app.js", http.StatusOK, "initFilters"},
		{"/filters.js", http.StatusOK, "fetchPuzzleCount"},
		{"/board.js", http.StatusOK, "PuzzleBoard"},
		{"/vendor/chessground.bundle.js", http.StatusOK, "Chessground"},
		{"/vendor/chessground.cburnett.css", http.StatusOK, "piece.pawn.white"},
		{"/style.css", http.StatusOK, "--bg"},
		{"/api/status", http.StatusOK, `"db_ready"`},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, req)

			if rec.Code != tc.wantCode {
				t.Fatalf("GET %s: status %d, want %d", tc.path, rec.Code, tc.wantCode)
			}
			if !strings.Contains(rec.Body.String(), tc.wantSubstr) {
				t.Fatalf("GET %s: body missing %q", tc.path, tc.wantSubstr)
			}
		})
	}
}
