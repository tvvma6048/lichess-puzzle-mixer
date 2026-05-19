package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/DSerejo/lichess-puzzle-mixer/internal/db"
)

type statusResponse struct {
	DBReady       bool   `json:"db_ready"`
	PuzzleCount   int64  `json:"puzzle_count"`
	ImportedAt    string `json:"imported_at,omitempty"`
	SchemaVersion int    `json:"schema_version"`
	Message       string `json:"message,omitempty"`
}

func (s *Server) api(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/status":
		s.handleStatus(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/api/puzzle/next":
		s.handlePuzzleNext(w, r)
	case r.URL.Path == "/api/puzzle/count":
		s.handlePuzzleCount(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/api/puzzle/preamble":
		s.handlePuzzlePreamble(w, r)
	case r.URL.Path == "/api/database":
		s.handleDatabase(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/api/database/import-sample":
		s.handleDatabaseImportSample(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/api/database/import":
		s.handleDatabaseImportUpload(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/api/database/download-lichess":
		s.handleDatabaseDownloadLichess(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	resp := statusResponse{
		SchemaVersion: 1,
		Message:       "Lichess Puzzle Mixer",
	}
	if s.db != nil {
		st, err := s.db.Status()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		resp.DBReady = st.Ready
		resp.PuzzleCount = st.PuzzleCount
		resp.ImportedAt = st.ImportedAt
		resp.SchemaVersion = st.SchemaVersion
		if !st.Ready {
			resp.Message = "Import a puzzle database to begin (make import-sample)"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handlePuzzleNext(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("database not configured"))
		return
	}
	st, err := s.db.Status()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !st.Ready {
		writeError(w, http.StatusServiceUnavailable, errors.New("puzzle database not imported"))
		return
	}

	filter, err := db.DecodeFilter(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	puzzle, err := s.db.NextPuzzle(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if puzzle == nil {
		writeError(w, http.StatusNotFound, errors.New("no puzzle matches filter"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(puzzle)
}

func (s *Server) handlePuzzleCount(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("database not configured"))
		return
	}

	var filter db.Filter
	var err error
	switch r.Method {
	case http.MethodPost:
		filter, err = db.DecodeFilter(r.Body)
	case http.MethodGet:
		filter, err = filterFromQuery(r)
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	count, err := s.db.CountPuzzles(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int64{"count": count})
}

func filterFromQuery(r *http.Request) (db.Filter, error) {
	q := r.URL.Query()
	f := db.Filter{
		ThemesMode: q.Get("themes_mode"),
		SideToMove: q.Get("side_to_move"),
	}
	if v := q.Get("opening_family"); v != "" {
		f.OpeningFamily = &v
	}
	if v := q.Get("themes"); v != "" {
		f.Themes = strings.Split(v, ",")
	}
	if v := q.Get("theme_groups"); v != "" {
		for _, part := range strings.Split(v, "|") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			var group []string
			for _, t := range strings.Split(part, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					group = append(group, t)
				}
			}
			if len(group) > 0 {
				f.ThemeGroups = append(f.ThemeGroups, group)
			}
		}
	}
	f.RatingMin = atoiDefault(q.Get("rating_min"), 0)
	f.RatingMax = atoiDefault(q.Get("rating_max"), 0)
	f.PopularityMin = atoiDefault(q.Get("popularity_min"), 0)
	f.LengthMin = atoiDefault(q.Get("length_min"), 0)
	f.LengthMax = atoiDefault(q.Get("length_max"), 0)
	f.Normalize()
	return f, nil
}

func writeError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
