package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"strings"
)

type Filter struct {
	// ThemeGroups: each inner slice is OR; groups are ANDed (e.g. [["fork","pin"],["short"]]).
	ThemeGroups [][]string `json:"theme_groups,omitempty"`
	Themes      []string `json:"themes,omitempty"`      // legacy: single group
	ThemesMode  string   `json:"themes_mode,omitempty"` // legacy: "or" | "and"
	RatingMin   int      `json:"rating_min"`
	RatingMax     int      `json:"rating_max"`
	PopularityMin int      `json:"popularity_min"`
	SideToMove    string   `json:"side_to_move"`
	LengthMin     int      `json:"length_min"`
	LengthMax     int      `json:"length_max"`
	OpeningFamily *string  `json:"opening_family"`
}

func (f *Filter) Normalize() {
	f.ThemeGroups = f.resolvedThemeGroups()
	if f.RatingMax == 0 {
		f.RatingMax = 4000
	}
	if f.LengthMax == 0 {
		f.LengthMax = 99
	}
	if f.SideToMove == "" {
		f.SideToMove = "any"
	}
}

// resolvedThemeGroups merges legacy themes/themes_mode into theme_groups.
func (f *Filter) resolvedThemeGroups() [][]string {
	if len(f.ThemeGroups) > 0 {
		return dedupeThemeGroups(f.ThemeGroups)
	}
	if len(f.Themes) == 0 {
		return nil
	}
	mode := f.ThemesMode
	if mode == "" {
		mode = "or"
	}
	if mode == "and" {
		groups := make([][]string, 0, len(f.Themes))
		for _, t := range f.Themes {
			groups = append(groups, []string{t})
		}
		return dedupeThemeGroups(groups)
	}
	return dedupeThemeGroups([][]string{f.Themes})
}

func dedupeThemeGroups(groups [][]string) [][]string {
	out := make([][]string, 0, len(groups))
	for _, g := range groups {
		seen := make(map[string]struct{})
		var clean []string
		for _, t := range g {
			if t == "" {
				continue
			}
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			clean = append(clean, t)
		}
		if len(clean) > 0 {
			out = append(out, clean)
		}
	}
	return out
}

type Puzzle struct {
	PuzzleID       string   `json:"puzzle_id"`
	FEN            string   `json:"fen"`
	Moves          []string `json:"moves"`
	Rating         int      `json:"rating"`
	Themes         []string `json:"themes"`
	SideToMove     string   `json:"side_to_move"`
	SolutionLength int      `json:"solution_length"`
	GameURL        string   `json:"game_url,omitempty"`
	Opening        string   `json:"opening,omitempty"`
}

func DecodeFilter(r io.Reader) (Filter, error) {
	var f Filter
	if err := json.NewDecoder(r).Decode(&f); err != nil {
		return Filter{}, err
	}
	f.Normalize()
	return f, nil
}

func (db *DB) NextPuzzle(ctx context.Context, f Filter) (*Puzzle, error) {
	f.Normalize()
	query, args := buildFilterQuery(f, false)
	row := db.sql.QueryRowContext(ctx, query, args...)

	var p Puzzle
	var movesRaw string
	var gameURL sql.NullString
	var openingVar sql.NullString
	err := row.Scan(
		&p.PuzzleID, &p.FEN, &movesRaw, &p.Rating,
		&p.SideToMove, &p.SolutionLength, &gameURL, &openingVar,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.Moves = splitMoves(movesRaw)
	if gameURL.Valid {
		p.GameURL = gameURL.String
	}
	if openingVar.Valid {
		p.Opening = openingVar.String
	}
	p.Themes, err = db.ThemesForPuzzle(p.PuzzleID)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (db *DB) CountPuzzles(ctx context.Context, f Filter) (int64, error) {
	f.Normalize()
	query, args := buildFilterQuery(f, true)
	var count int64
	err := db.sql.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func buildFilterQuery(f Filter, countOnly bool) (string, []any) {
	var b strings.Builder
	var args []any

	if countOnly {
		b.WriteString("SELECT COUNT(*) FROM puzzles p WHERE 1=1")
	} else {
		b.WriteString(`SELECT p.puzzle_id, p.fen, p.moves, p.rating, p.side_to_move, p.solution_length, p.game_url, p.opening_variation
			FROM puzzles p WHERE 1=1`)
	}

	b.WriteString(" AND p.rating BETWEEN ? AND ?")
	args = append(args, f.RatingMin, f.RatingMax)

	b.WriteString(" AND p.popularity >= ?")
	args = append(args, f.PopularityMin)

	b.WriteString(" AND p.solution_length BETWEEN ? AND ?")
	args = append(args, f.LengthMin, f.LengthMax)

	b.WriteString(" AND (? = 'any' OR p.side_to_move = ?)")
	args = append(args, f.SideToMove, f.SideToMove)

	if f.OpeningFamily != nil && *f.OpeningFamily != "" {
		b.WriteString(" AND p.opening_family = ?")
		args = append(args, *f.OpeningFamily)
	}

	for _, group := range f.ThemeGroups {
		if len(group) == 0 {
			continue
		}
		placeholders := strings.Repeat("?,", len(group))
		placeholders = placeholders[:len(placeholders)-1]
		b.WriteString(" AND EXISTS (SELECT 1 FROM puzzle_themes pt WHERE pt.puzzle_id = p.puzzle_id AND pt.theme IN (")
		b.WriteString(placeholders)
		b.WriteString("))")
		for _, t := range group {
			args = append(args, t)
		}
	}

	if !countOnly {
		b.WriteString(" ORDER BY RANDOM() LIMIT 1")
	}

	return b.String(), args
}
