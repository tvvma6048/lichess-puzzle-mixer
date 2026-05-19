CREATE TABLE IF NOT EXISTS puzzles (
    puzzle_id          TEXT PRIMARY KEY,
    fen                TEXT NOT NULL,
    moves              TEXT NOT NULL,
    rating             INTEGER NOT NULL,
    rating_deviation   INTEGER NOT NULL,
    popularity         INTEGER NOT NULL,
    nb_plays           INTEGER NOT NULL,
    game_url           TEXT,
    opening_family     TEXT,
    opening_variation  TEXT,
    side_to_move       TEXT NOT NULL,
    solution_length    INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_puzzles_rating ON puzzles(rating);
CREATE INDEX IF NOT EXISTS idx_puzzles_popularity ON puzzles(popularity);
CREATE INDEX IF NOT EXISTS idx_puzzles_side ON puzzles(side_to_move);
CREATE INDEX IF NOT EXISTS idx_puzzles_length ON puzzles(solution_length);
CREATE INDEX IF NOT EXISTS idx_puzzles_opening_family ON puzzles(opening_family);

CREATE TABLE IF NOT EXISTS puzzle_themes (
    puzzle_id  TEXT NOT NULL,
    theme      TEXT NOT NULL,
    PRIMARY KEY (puzzle_id, theme)
) WITHOUT ROWID;

CREATE INDEX IF NOT EXISTS idx_puzzle_themes_theme ON puzzle_themes(theme);

CREATE TABLE IF NOT EXISTS meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
