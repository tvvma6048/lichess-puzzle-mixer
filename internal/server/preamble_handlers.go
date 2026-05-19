package server

import (
	"encoding/json"
	"net/http"

	"github.com/DSerejo/lichess-puzzle-mixer/internal/lichess"
)

type preambleResponse struct {
	OK  bool   `json:"ok"`
	PGN string `json:"pgn,omitempty"`
	Ply int    `json:"ply,omitempty"`
}

func (s *Server) handlePuzzlePreamble(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameURL := r.URL.Query().Get("game_url")
	fen := r.URL.Query().Get("fen")
	if gameURL == "" || fen == "" {
		writePreamble(w, preambleResponse{OK: false})
		return
	}

	ref, err := lichess.ParseGameURL(gameURL)
	if err != nil {
		writePreamble(w, preambleResponse{OK: false})
		return
	}

	pgn, err := lichess.FetchPGN(r.Context(), ref.ID)
	if err != nil {
		writePreamble(w, preambleResponse{OK: false})
		return
	}

	writePreamble(w, preambleResponse{
		OK:  true,
		PGN: pgn,
		Ply: ref.Ply,
	})
}

func writePreamble(w http.ResponseWriter, resp preambleResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
