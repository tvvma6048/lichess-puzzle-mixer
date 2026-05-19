package server

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DSerejo/lichess-puzzle-mixer/internal/db"
)

type Config struct {
	Dev       bool
	DataDir   string
	WebFS     fs.FS // embedded web/ tree (production); ignored when Dev is true
	DB        *db.DB
	SampleCSV []byte
}

type Server struct {
	dataDir   string
	web       http.Handler
	db        *db.DB
	sampleCSV []byte
	importJob importJob
}

func New(cfg Config) *Server {
	var webHandler http.Handler
	if cfg.Dev {
		webRoot, err := filepath.Abs("web")
		if err != nil {
			slog.Error("resolve web dir", "err", err)
			os.Exit(1)
		}
		slog.Info("serving web from disk", "path", webRoot)
		webHandler = devFileServer(webRoot)
	} else {
		sub, err := fs.Sub(cfg.WebFS, "web")
		if err != nil {
			slog.Error("embed web fs", "err", err)
			os.Exit(1)
		}
		webHandler = http.FileServer(http.FS(sub))
	}

	return &Server{
		dataDir:   cfg.DataDir,
		web:       webHandler,
		db:        cfg.DB,
		sampleCSV: cfg.SampleCSV,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		slog.Debug("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	}()

	if strings.HasPrefix(r.URL.Path, "/api/") {
		s.api(w, r)
		return
	}
	s.web.ServeHTTP(w, r)
}

// devFileServer serves static files and disables caching so refreshes pick up edits.
func devFileServer(root string) http.Handler {
	fileServer := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		fileServer.ServeHTTP(w, r)
	})
}
