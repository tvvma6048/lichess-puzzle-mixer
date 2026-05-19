package lichess

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var gameIDRe = regexp.MustCompile(`lichess\.org/([a-zA-Z0-9]{8,12})`)

// GameRef is parsed from a Lichess game URL.
type GameRef struct {
	ID  string
	Ply int // half-move index from URL hash (#53), 0 if absent
}

// ParseGameURL extracts game id and optional ply from a Lichess game link.
func ParseGameURL(raw string) (GameRef, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return GameRef{}, fmt.Errorf("empty game url")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return GameRef{}, err
	}
	m := gameIDRe.FindStringSubmatch(u.String())
	if len(m) < 2 {
		return GameRef{}, fmt.Errorf("not a lichess game url")
	}
	ref := GameRef{ID: m[1]}
	if u.Fragment != "" {
		if ply, err := strconv.Atoi(strings.TrimPrefix(u.Fragment, "#")); err == nil && ply > 0 {
			ref.Ply = ply
		}
	}
	return ref, nil
}

// FetchPGN downloads the game PGN from Lichess.
func FetchPGN(ctx context.Context, gameID string) (string, error) {
	exportURL := fmt.Sprintf("https://lichess.org/game/export/%s.pgn", gameID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, exportURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "github.com/DSerejo/lichess-puzzle-mixer/1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("lichess export: HTTP %d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 512*1024))
	if err != nil {
		return "", err
	}
	return string(body), nil
}
