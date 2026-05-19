package config

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/DSerejo/lichess-puzzle-mixer/internal/paths"
)

const currentVersion = 1

type Config struct {
	Version           int    `json:"version"`
	SoundEnabled      bool   `json:"sound_enabled"`
	AnimationsEnabled bool   `json:"animations_enabled"`
	RevealThemesAfter bool   `json:"reveal_themes_after_solve"`
	PieceSet          string `json:"piece_set"`
	BoardTheme        string `json:"board_theme"`
	AutoNextPuzzle    bool   `json:"auto_next_puzzle"`
	AutoNextDelayMs   int    `json:"auto_next_delay_ms"`
	PreferredPort     int    `json:"preferred_port"`
}

func Defaults() Config {
	return Config{
		Version:           currentVersion,
		SoundEnabled:      true,
		AnimationsEnabled: true,
		RevealThemesAfter: true,
		PieceSet:          "cburnett",
		BoardTheme:        "brown",
		AutoNextPuzzle:    true,
		AutoNextDelayMs:   800,
		PreferredPort:     7777,
	}
}

func Load(dataDir string) (Config, error) {
	path := paths.ConfigPath(dataDir)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		cfg := Defaults()
		if saveErr := Save(dataDir, cfg); saveErr != nil {
			return cfg, saveErr
		}
		return cfg, nil
	}
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Version == 0 {
		cfg.Version = currentVersion
	}
	if cfg.PreferredPort == 0 {
		cfg.PreferredPort = 7777
	}
	return cfg, nil
}

func Save(dataDir string, cfg Config) error {
	path := paths.ConfigPath(dataDir)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
