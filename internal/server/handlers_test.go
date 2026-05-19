package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/DSerejo/lichess-puzzle-mixer/internal/db"
)

func TestHandleStatusEmptyDB(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "puzzles.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	s := New(Config{Dev: true, DataDir: t.TempDir(), DB: database})
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", rec.Code)
	}
	var body statusResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.DBReady {
		t.Fatal("expected db_ready=false")
	}
}

func TestHandlePuzzleNext(t *testing.T) {
	csvPath := filepath.Join("..", "..", "testdata", "lichess_sample.csv")
	database, err := db.Open(filepath.Join(t.TempDir(), "puzzles.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer database.Close()
	if _, err := database.ImportCSV(t.Context(), csvPath); err != nil {
		t.Fatalf("import: %v", err)
	}

	s := New(Config{Dev: true, DataDir: t.TempDir(), DB: database})
	body := `{"themes":["fork"],"themes_mode":"or","rating_min":0,"rating_max":4000,"popularity_min":-100,"length_min":0,"length_max":99,"side_to_move":"any"}`
	req := httptest.NewRequest(http.MethodPost, "/api/puzzle/next", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var puzzle db.Puzzle
	if err := json.NewDecoder(rec.Body).Decode(&puzzle); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if puzzle.PuzzleID == "" {
		t.Fatal("expected puzzle_id")
	}
}

func TestHandlePuzzleCountThemeGroups(t *testing.T) {
	csvPath := filepath.Join("..", "..", "testdata", "lichess_sample.csv")
	database, err := db.Open(filepath.Join(t.TempDir(), "puzzles.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer database.Close()
	if _, err := database.ImportCSV(t.Context(), csvPath); err != nil {
		t.Fatalf("import: %v", err)
	}

	s := New(Config{Dev: true, DataDir: t.TempDir(), DB: database})
	baseFilter := `{"rating_min":0,"rating_max":4000,"popularity_min":-100,"length_min":0,"length_max":99,"side_to_move":"any"}`

	allReq := httptest.NewRequest(http.MethodPost, "/api/puzzle/count", bytes.NewBufferString(baseFilter))
	allRec := httptest.NewRecorder()
	s.ServeHTTP(allRec, allReq)
	if allRec.Code != http.StatusOK {
		t.Fatalf("all count status %d body %s", allRec.Code, allRec.Body.String())
	}
	var allBody map[string]int64
	if err := json.NewDecoder(allRec.Body).Decode(&allBody); err != nil {
		t.Fatalf("decode all: %v", err)
	}
	if allBody["count"] != 500 {
		t.Fatalf("unfiltered count = %d, want 500", allBody["count"])
	}

	forkBody := baseFilter[:len(baseFilter)-1] + `,"theme_groups":[["fork"]]}`
	forkReq := httptest.NewRequest(http.MethodPost, "/api/puzzle/count", bytes.NewBufferString(forkBody))
	forkRec := httptest.NewRecorder()
	s.ServeHTTP(forkRec, forkReq)
	if forkRec.Code != http.StatusOK {
		t.Fatalf("fork count status %d body %s", forkRec.Code, forkRec.Body.String())
	}
	var forkCount map[string]int64
	if err := json.NewDecoder(forkRec.Body).Decode(&forkCount); err != nil {
		t.Fatalf("decode fork: %v", err)
	}
	if forkCount["count"] != 54 {
		t.Fatalf("fork count = %d, want 54", forkCount["count"])
	}
	if forkCount["count"] >= allBody["count"] {
		t.Fatalf("fork count should be less than unfiltered")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/puzzle/count?theme_groups=fork,pin%7Cshort&rating_min=0&rating_max=4000&popularity_min=-100&length_min=0&length_max=99&side_to_move=any", nil)
	getRec := httptest.NewRecorder()
	s.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get count status %d body %s", getRec.Code, getRec.Body.String())
	}
	var groupedCount map[string]int64
	if err := json.NewDecoder(getRec.Body).Decode(&groupedCount); err != nil {
		t.Fatalf("decode grouped: %v", err)
	}
	if groupedCount["count"] != 50 {
		t.Fatalf("(fork|pin)+short count = %d, want 50", groupedCount["count"])
	}
}

func TestDatabaseImportSampleAndDelete(t *testing.T) {
	csvPath := filepath.Join("..", "..", "testdata", "lichess_sample.csv")
	sample, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}

	dataDir := t.TempDir()
	database, err := db.Open(filepath.Join(dataDir, "puzzles.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer database.Close()

	s := New(Config{Dev: true, DataDir: dataDir, DB: database, SampleCSV: sample})

	req := httptest.NewRequest(http.MethodPost, "/api/database/import-sample", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("import sample: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/database", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	var info databaseResponse
	if err := json.NewDecoder(rec.Body).Decode(&info); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !info.DBReady || info.PuzzleCount == 0 {
		t.Fatalf("expected ready database, got %+v", info)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/database", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: %d %s", rec.Code, rec.Body.String())
	}

	st, err := database.Status()
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if st.Ready || st.PuzzleCount != 0 {
		t.Fatalf("expected empty db, got %+v", st)
	}
}

func TestPuzzlePreambleInvalidURL(t *testing.T) {
	s := New(Config{Dev: true, DataDir: t.TempDir()})
	req := httptest.NewRequest(http.MethodGet, "/api/puzzle/preamble?game_url=not-a-url&fen=start", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var body preambleResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.OK {
		t.Fatal("expected ok=false for bad url")
	}
}

func TestAPIUnknownRoute(t *testing.T) {
	s := New(Config{Dev: true, DataDir: t.TempDir()})

	req := httptest.NewRequest(http.MethodGet, "/api/unknown", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want 404", rec.Code)
	}
}
