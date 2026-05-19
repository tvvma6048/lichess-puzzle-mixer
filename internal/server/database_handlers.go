package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DSerejo/lichess-puzzle-mixer/internal/paths"
)

const (
	lichessPuzzleCSVURL = "https://database.lichess.org/lichess_db_puzzle.csv.bz2"
	maxUploadBytes      = 512 << 20 // 512 MiB
)

type databaseResponse struct {
	DBReady       bool   `json:"db_ready"`
	PuzzleCount   int64  `json:"puzzle_count"`
	ImportedAt    string `json:"imported_at,omitempty"`
	SchemaVersion int    `json:"schema_version"`
	DataDir       string `json:"data_dir"`
	DBPath        string `json:"db_path"`
	DBSizeBytes   int64  `json:"db_size_bytes"`
	ImportRunning bool   `json:"import_running"`
	ImportStage   string `json:"import_stage,omitempty"`
	ImportMessage string `json:"import_message,omitempty"`
	ImportError   string `json:"import_error,omitempty"`
	ImportRows    int64  `json:"import_rows,omitempty"`
}

type importResultResponse struct {
	RowsImported int64  `json:"rows_imported"`
	Message      string `json:"message"`
}

func (s *Server) handleDatabase(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleDatabaseGet(w, r)
	case http.MethodDelete:
		s.handleDatabaseDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleDatabaseGet(w http.ResponseWriter, r *http.Request) {
	resp, err := s.databaseInfo()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleDatabaseDelete(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("database not configured"))
		return
	}
	if running, _, _, _, _ := s.importJob.snapshot(); running {
		writeError(w, http.StatusConflict, errors.New("import in progress"))
		return
	}
	if err := s.db.Wipe(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "database cleared"})
}

func (s *Server) handleDatabaseImportSample(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.db == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("database not configured"))
		return
	}
	if len(s.sampleCSV) == 0 {
		writeError(w, http.StatusInternalServerError, errors.New("sample database not bundled"))
		return
	}
	if !s.importJob.start("importing", "Importing sample puzzles…") {
		writeError(w, http.StatusConflict, errors.New("import in progress"))
		return
	}

	result, err := s.db.ImportReader(r.Context(), bytes.NewReader(s.sampleCSV), "lichess_sample.csv")
	s.importJob.finish(result.RowsImported, err)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(importResultResponse{
		RowsImported: result.RowsImported,
		Message:      fmt.Sprintf("Imported %d puzzles", result.RowsImported),
	})
}

func (s *Server) handleDatabaseImportUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.db == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("database not configured"))
		return
	}
	if !s.importJob.start("importing", "Importing uploaded file…") {
		writeError(w, http.StatusConflict, errors.New("import in progress"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		s.importJob.finish(0, err)
		writeError(w, http.StatusBadRequest, fmt.Errorf("upload too large or invalid: %w", err))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.importJob.finish(0, err)
		writeError(w, http.StatusBadRequest, errors.New("missing form field: file"))
		return
	}
	defer file.Close()

	name := header.Filename
	if name == "" {
		name = "upload.csv"
	}
	lower := strings.ToLower(name)
	if !strings.HasSuffix(lower, ".csv") && !strings.HasSuffix(lower, ".bz2") {
		s.importJob.finish(0, errors.New("file must be .csv or .csv.bz2"))
		writeError(w, http.StatusBadRequest, errors.New("file must be .csv or .csv.bz2"))
		return
	}

	result, err := s.db.ImportReader(r.Context(), file, name)
	s.importJob.finish(result.RowsImported, err)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(importResultResponse{
		RowsImported: result.RowsImported,
		Message:      fmt.Sprintf("Imported %d puzzles", result.RowsImported),
	})
}

func (s *Server) handleDatabaseDownloadLichess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.db == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("database not configured"))
		return
	}
	if !s.importJob.start("downloading", "Downloading Lichess puzzle database…") {
		writeError(w, http.StatusConflict, errors.New("import in progress"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Download started; poll GET /api/database for progress",
	})

	go s.runLichessDownloadImport()
}

func (s *Server) runLichessDownloadImport() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	tmpPath := filepath.Join(s.dataDir, "lichess_import.tmp.bz2")
	defer os.Remove(tmpPath)

	if err := downloadFile(ctx, lichessPuzzleCSVURL, tmpPath, func(stage, msg string) {
		s.importJob.set(stage, msg)
	}); err != nil {
		s.importJob.finish(0, err)
		return
	}

	s.importJob.set("importing", "Importing puzzles into SQLite…")
	f, err := os.Open(tmpPath)
	if err != nil {
		s.importJob.finish(0, err)
		return
	}
	defer f.Close()

	result, err := s.db.ImportReader(ctx, f, "lichess_db_puzzle.csv.bz2")
	s.importJob.finish(result.RowsImported, err)
}

func downloadFile(ctx context.Context, url, dest string, onProgress func(stage, message string)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", res.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	total := res.ContentLength
	var written int64
	buf := make([]byte, 32*1024)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, readErr := res.Body.Read(buf)
		if n > 0 {
			if _, err := out.Write(buf[:n]); err != nil {
				return err
			}
			written += int64(n)
			if onProgress != nil && total > 0 {
				pct := float64(written) / float64(total) * 100
				onProgress("downloading", fmt.Sprintf("Downloading… %.0f%%", pct))
			} else if onProgress != nil {
				onProgress("downloading", fmt.Sprintf("Downloading… %d MB", written/(1024*1024)))
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	return nil
}

func (s *Server) databaseInfo() (databaseResponse, error) {
	resp := databaseResponse{SchemaVersion: 1, DataDir: s.dataDir}
	dbPath := paths.DBPath(s.dataDir)
	resp.DBPath = dbPath

	if info, err := os.Stat(dbPath); err == nil {
		resp.DBSizeBytes = info.Size()
	}

	running, stage, message, impErr, rows := s.importJob.snapshot()
	resp.ImportRunning = running
	resp.ImportStage = stage
	resp.ImportMessage = message
	resp.ImportError = impErr
	resp.ImportRows = rows

	if s.db == nil {
		return resp, nil
	}
	st, err := s.db.Status()
	if err != nil {
		return resp, err
	}
	resp.DBReady = st.Ready
	resp.PuzzleCount = st.PuzzleCount
	resp.ImportedAt = st.ImportedAt
	resp.SchemaVersion = st.SchemaVersion
	return resp, nil
}
