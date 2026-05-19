package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/DSerejo/lichess-puzzle-mixer/internal/browser"
	"github.com/DSerejo/lichess-puzzle-mixer/internal/config"
	"github.com/DSerejo/lichess-puzzle-mixer/internal/db"
	"github.com/DSerejo/lichess-puzzle-mixer/internal/paths"
	"github.com/DSerejo/lichess-puzzle-mixer/internal/server"
	"github.com/DSerejo/lichess-puzzle-mixer/internal/tray"
)

func main() {
	dev := flag.Bool("dev", false, "serve web assets from ./web on disk (no rebuild for frontend changes)")
	dataDir := flag.String("data-dir", "", "app data directory (default: ./.devdata in dev mode, OS app dir otherwise)")
	port := flag.Int("port", 7777, "HTTP listen port")
	noBrowser := flag.Bool("no-browser", false, "do not open a browser on startup")
	importCSV := flag.String("import-csv", "", "import puzzles from a Lichess CSV file (.csv or .csv.bz2)")
	importOnly := flag.Bool("import-only", false, "exit after --import-csv completes")
	flag.Parse()

	resolvedDir, err := paths.DataDir(resolveDataDir(*dataDir, *dev))
	if err != nil {
		slog.Error("data dir", "err", err)
		os.Exit(1)
	}

	if _, err := config.Load(resolvedDir); err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	database, err := db.Open(paths.DBPath(resolvedDir))
	if err != nil {
		slog.Error("open database", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	if *importCSV != "" {
		result, err := database.ImportCSV(context.Background(), *importCSV)
		if err != nil {
			slog.Error("import csv", "path", *importCSV, "err", err)
			os.Exit(1)
		}
		slog.Info("import complete", "rows", result.RowsImported, "db", paths.DBPath(resolvedDir))
		if *importOnly {
			return
		}
	}

	if *importOnly {
		slog.Error("--import-only requires --import-csv")
		os.Exit(1)
	}

	cfg, err := config.Load(resolvedDir)
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}
	if *port == 7777 && cfg.PreferredPort != 0 {
		*port = cfg.PreferredPort
	}

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("listen", "addr", addr, "err", err)
		os.Exit(1)
	}

	srv := server.New(server.Config{
		Dev:       *dev,
		DataDir:   resolvedDir,
		WebFS:     webFS,
		DB:        database,
		SampleCSV: sampleCSV,
	})

	httpServer := &http.Server{
		Handler:      srv,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	url := "http://" + ln.Addr().String()
	slog.Info("starting", "url", url, "dev", *dev, "data_dir", resolvedDir)

	go func() {
		if err := httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			slog.Error("serve", "err", err)
			os.Exit(1)
		}
	}()

	if !*noBrowser {
		if err := browser.Open(url); err != nil {
			slog.Warn("open browser", "err", err, "url", url)
		}
	}

	var shutdownOnce sync.Once
	shutdownCh := make(chan struct{})
	requestShutdown := func() {
		shutdownOnce.Do(func() { close(shutdownCh) })
	}

	if tray.Available() {
		slog.Info("system tray active — use Quit in the tray menu to exit")
		go tray.Run(url, requestShutdown)
	} else {
		slog.Info("running without system tray; press Ctrl+C in the terminal to quit", "url", url)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
	case <-shutdownCh:
	}

	slog.Info("shutting down")
	tray.Quit()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

func resolveDataDir(flagVal string, dev bool) string {
	if flagVal != "" {
		return flagVal
	}
	if dev {
		return "./.devdata"
	}
	return paths.DefaultDataDir()
}
