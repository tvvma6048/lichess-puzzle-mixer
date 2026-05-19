package paths

import (
	"os"
	"path/filepath"
)

// DataDir returns the application data directory, creating it if missing.
func DataDir(override string) (string, error) {
	dir := override
	if dir == "" {
		dir = DefaultDataDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Abs(dir)
}

// DefaultDataDir is the OS-appropriate app data location.
func DefaultDataDir() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "lichess-puzzle-mixer")
	}
	return "./.devdata"
}

func DBPath(dataDir string) string {
	return filepath.Join(dataDir, "puzzles.db")
}

func ConfigPath(dataDir string) string {
	return filepath.Join(dataDir, "config.json")
}
