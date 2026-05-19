# Chess Puzzle Trainer — Design Spec

A standalone desktop app that serves a local web UI for solving Lichess puzzles with **multi-theme filtering** — the missing "custom Healthy Mix" feature.

Distributed as a single Go binary per platform. UI runs in the user's browser at `http://localhost:<port>`.

---

## 1. Goals & Non-Goals

### Goals

- Single-binary install, no runtime dependencies, ~20MB per platform.
- Pick **multiple puzzle themes** with OR / AND combination logic.
- Full Lichess filter parity: themes, rating, popularity, opening, side to move, puzzle length.
- Save and recall **named filter presets** ("Tactics warmup", "Endgame grind", etc.).
- Lichess-quality board: drag + click moves, animations, sounds.
- On first launch, prompt the user to download the Lichess puzzle DB (~330MB bz2).
- Auto-open the user's browser to the app on launch.

### Non-Goals (v1)

- User accounts / profiles.
- Solve history, stats, ratings, streaks.
- Engine analysis after solve.
- Mobile-optimized UI (desktop-first; should still work on tablet).
- Spaced repetition / Woodpecker mode.
- Cloud sync.

---

## 2. Tech Stack


| Layer             | Choice                                                       | Why                                                                                                                    |
| ----------------- | ------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------- |
| Backend language  | **Go 1.22+**                                                 | Single static binary, easy cross-compile, fast HTTP, `embed.FS` for assets.                                            |
| HTTP server       | `net/http` stdlib + `chi` router (or stdlib mux on Go 1.22+) | Tiny, no heavy deps.                                                                                                   |
| Database          | **SQLite** via `modernc.org/sqlite` (pure-Go driver)         | No CGO → trivial cross-compilation. Slightly slower than `mattn/go-sqlite3` but speed is irrelevant for this workload. |
| Decompression     | `compress/bzip2` (stdlib)                                    | bz2 is in the standard library — zero extra deps.                                                                      |
| Frontend bundling | None                                                         | Vanilla JS + ES modules. All assets embedded via `embed.FS`.                                                           |
| Board UI          | [chessground](https://github.com/lichess-org/chessground)    | Same board as Lichess. Vendored copy in `web/vendor/`.                                                                 |
| Chess rules       | [chess.js](https://github.com/jhlywa/chess.js)               | Move generation, validation, FEN/PGN. Vendored.                                                                        |
| Styling           | Hand-rolled CSS, Lichess-inspired                            | No framework needed for this scope.                                                                                    |


---

## 3. Project Layout

```
puzzle-trainer/
├── go.mod
├── go.sum
├── main.go                       # Entry point: parse flags, start server, open browser
├── internal/
│   ├── server/
│   │   ├── server.go             # HTTP setup, route registration
│   │   ├── handlers.go           # Route handlers
│   │   └── middleware.go         # Logging, CORS for localhost dev
│   ├── db/
│   │   ├── db.go                 # SQLite open, migrations, prepared statements
│   │   ├── schema.sql            # Embedded via go:embed
│   │   ├── query.go              # Filter → SQL builder
│   │   └── import.go             # CSV → SQLite ingestion
│   ├── lichess/
│   │   └── download.go           # Stream-download + bz2 decompress
│   ├── paths/
│   │   └── paths.go              # OS-specific app data dir resolution
│   ├── presets/
│   │   └── presets.go            # Load/save presets.json
│   └── browser/
│       └── open.go               # Platform-specific browser launch
├── web/                          # Embedded via go:embed
│   ├── index.html
│   ├── app.js                    # Main SPA logic
│   ├── board.js                  # Chessground wiring + puzzle play loop
│   ├── filters.js                # Filter panel UI
│   ├── presets.js                # Preset management UI
│   ├── api.js                    # Fetch wrappers
│   ├── style.css
│   ├── sounds/                   # move.mp3, capture.mp3, success.mp3, error.mp3
│   ├── pieces/                   # SVG piece sets (cburnett, etc.)
│   └── vendor/
│       ├── chessground.min.js
│       ├── chessground.css
│       └── chess.min.js
└── scripts/
    └── build-all.sh              # Cross-compile for windows/linux/macos (amd64+arm64)

```

---

## 4. Architecture

```
┌────────────────────────────────────────────────────────────┐
│ puzzle-trainer binary                                       │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ Go process                                            │   │
│ │  • Listens on 127.0.0.1:<port> (default 7777,         │   │
│ │    auto-increments if taken)                          │   │
│ │  • Serves embedded /web assets at /                   │   │
│ │  • Exposes JSON API at /api/*                         │   │
│ │  • Opens browser to app URL on launch                 │   │
│ └──────────────────────────────────────────────────────┘   │
│                       │                                     │
│                       │ SQL                                 │
│                       ▼                                     │
│              ┌─────────────────┐                            │
│              │  puzzles.db     │  ← SQLite, ~1.5GB           │
│              │  presets.json   │  ← saved filter presets     │
│              │  config.json    │  ← settings (sound, etc.)   │
│              └─────────────────┘                            │
│              in OS app-data dir                             │
└────────────────────────────────────────────────────────────┘

First-run flow (no puzzles.db):
  user opens binary
    → Go starts, opens browser
    → frontend hits /api/status → sees db_ready: false
    → shows setup screen with "Download puzzle database" button
    → user clicks → POST /api/db/import
    → Go streams Lichess .bz2 → decompress → CSV parse → INSERT
    → Server-Sent Events stream progress to frontend
    → on completion, frontend reloads → ready

Steady-state flow:
  user picks themes + filters in UI
    → frontend builds filter object, POSTs to /api/puzzle/next
    → Go queries SQLite, returns one random matching puzzle
    → frontend renders position on chessground
    → user plays moves; frontend validates against solution
    → on solve → POST /api/puzzle/next for the next one

```

---

## 5. Data Model

### 5.1 Why SQLite (not MySQL/Postgres/Mongo)

The dataset is **~5M rows, read-only after import, single-user, localhost-only**. SQLite is the right choice on every dimension:

- **Zero ops:** it's a file. No daemon, no port, no config, no service to start.
- **Speed:** with proper indexes, all filter queries return in <50ms even on the full DB.
- **Portability:** one file, copy anywhere, works.
- **Embeddable:** ships with the binary via the Go driver. Nothing for the user to install.

MySQL/Postgres would demand a server process for zero benefit at this scale. Mongo's document model gives nothing — the data is rigidly tabular and the core operation is set intersection over themes, which SQL does cleanly.

### 5.2 Schema

The Lichess CSV stores themes as a space-separated string per row (e.g. `"crushing fork middlegame short"`). For multi-theme queries to be fast, themes **must** be denormalized into a junction table. `LIKE '%fork%'` cannot use an index and would scan every row.

```sql
-- internal/db/schema.sql

CREATE TABLE IF NOT EXISTS puzzles (
    puzzle_id          TEXT PRIMARY KEY,
    fen                TEXT NOT NULL,    -- position BEFORE opponent's first move
    moves              TEXT NOT NULL,    -- UCI moves, space-separated; first = opponent setup, rest = solution
    rating             INTEGER NOT NULL,
    rating_deviation   INTEGER NOT NULL,
    popularity         INTEGER NOT NULL, -- -100..100
    nb_plays           INTEGER NOT NULL,
    game_url           TEXT,
    opening_family     TEXT,             -- nullable; only set for puzzles before move 20
    opening_variation  TEXT,             -- nullable
    side_to_move       TEXT NOT NULL,    -- 'w' or 'b' — derived from FEN at import time
    solution_length    INTEGER NOT NULL  -- number of plies in the solution (not counting opponent setup)
);

CREATE INDEX idx_puzzles_rating          ON puzzles(rating);
CREATE INDEX idx_puzzles_popularity      ON puzzles(popularity);
CREATE INDEX idx_puzzles_side            ON puzzles(side_to_move);
CREATE INDEX idx_puzzles_length          ON puzzles(solution_length);
CREATE INDEX idx_puzzles_opening_family  ON puzzles(opening_family);

CREATE TABLE IF NOT EXISTS puzzle_themes (
    puzzle_id  TEXT NOT NULL,
    theme      TEXT NOT NULL,
    PRIMARY KEY (puzzle_id, theme)
) WITHOUT ROWID;

CREATE INDEX idx_puzzle_themes_theme ON puzzle_themes(theme);

CREATE TABLE IF NOT EXISTS meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
-- expected keys:
--   schema_version    e.g. "1"
--   imported_at       ISO-8601 timestamp
--   lichess_csv_etag  HTTP ETag from the download for update detection
--   puzzle_count      total row count, cached for fast UI display

```

**Size estimate:** ~5M puzzles × ~5 themes average = ~25M junction rows. Total `.db` file ~1.2–1.6GB. The size is worth the query speed.

`side_to_move` and `solution_length` are derived once at import time — they're cheap to compute and let us index them for fast filtering, rather than re-parsing FEN/moves on every query.

### 5.3 Files in app data dir

```
~/.local/share/puzzle-trainer/                       (Linux)
%APPDATA%/puzzle-trainer/                            (Windows)
~/Library/Application Support/puzzle-trainer/        (macOS)

├── puzzles.db        # ~1.5GB SQLite
├── presets.json      # saved filter presets
└── config.json       # UI settings (sound, animations, reveal-themes toggle, etc.)

```

Path resolution via `os.UserConfigDir()` on Windows/macOS and `$XDG_DATA_HOME` (with `~/.local/share` fallback) on Linux. App-data dir is more appropriate than cache dir for a ~1.5GB asset users would not want silently evicted.

### 5.4 presets.json format

```json
{
  "version": 1,
  "presets": [
    {
      "id": "01HXYZ...",
      "name": "Tactics warmup",
      "filter": {
        "themes": ["fork", "pin", "skewer", "discoveredAttack"],
        "themes_mode": "or",
        "rating_min": 1200,
        "rating_max": 1500,
        "popularity_min": 80,
        "side_to_move": "any",
        "length_min": 1,
        "length_max": 3,
        "opening_family": null
      },
      "created_at": "2026-05-18T10:00:00Z"
    }
  ]
}

```

### 5.5 config.json format

```json
{
  "version": 1,
  "sound_enabled": true,
  "animations_enabled": true,
  "reveal_themes_after_solve": true,
  "piece_set": "cburnett",
  "board_theme": "brown",
  "auto_next_puzzle": true,
  "auto_next_delay_ms": 800,
  "preferred_port": 7777
}

```

---

## 6. Filter → SQL

The filter shape sent from frontend to backend:

```json
{
  "themes": ["fork", "pin", "skewer"],
  "themes_mode": "or",
  "rating_min": 1200,
  "rating_max": 1800,
  "popularity_min": 80,
  "side_to_move": "any",
  "length_min": 1,
  "length_max": 5,
  "opening_family": null
}

```

### 6.1 OR mode (any selected theme)

```sql
SELECT p.*
FROM puzzles p
WHERE p.rating BETWEEN ? AND ?
  AND p.popularity >= ?
  AND p.solution_length BETWEEN ? AND ?
  AND (? = 'any' OR p.side_to_move = ?)
  AND (? IS NULL OR p.opening_family = ?)
  AND EXISTS (
    SELECT 1 FROM puzzle_themes pt
    WHERE pt.puzzle_id = p.puzzle_id
      AND pt.theme IN (?, ?, ?)  -- expanded to match selected themes
  )
ORDER BY RANDOM()
LIMIT 1;

```

### 6.2 AND mode (must have all selected themes)

```sql
SELECT p.*
FROM puzzles p
WHERE p.rating BETWEEN ? AND ?
  AND p.popularity >= ?
  AND p.solution_length BETWEEN ? AND ?
  AND (? = 'any' OR p.side_to_move = ?)
  AND (? IS NULL OR p.opening_family = ?)
  AND (
    SELECT COUNT(DISTINCT theme)
    FROM puzzle_themes
    WHERE puzzle_id = p.puzzle_id
      AND theme IN (?, ?, ?)
  ) = ?  -- count of selected themes
ORDER BY RANDOM()
LIMIT 1;

```

### 6.3 Performance notes

- `ORDER BY RANDOM() LIMIT 1` scans the matching set. For typical filters (a handful of themes + rating range) this is <50ms.
- For extremely broad filters that match millions of rows, an alternative is two queries: `SELECT COUNT(*)` then `SELECT ... LIMIT 1 OFFSET <random>`. Optimize only if a real filter is slow.
- Always run `ANALYZE` once after import — it builds query planner statistics. Significant speedup on the junction query.
- Use `PRAGMA cache_size = -64000` (64MB page cache) on connection open. Helps hot queries against the themes index.
- Open the read connection with `?_pragma=query_only(1)` to lock the DB read-only for safety after import is complete.

### 6.4 Live match counter

The UI shows the count of puzzles matching the current filter, so the user knows if their AND combination is too restrictive. Implemented as a debounced `GET /api/puzzle/count?...` — same query as above with `COUNT(*)` instead of `SELECT p.*`. Should return in <100ms.

---

## 7. HTTP API

All endpoints under `127.0.0.1:<port>/api`. Localhost-only bind. No auth (single user, local machine).

### `GET /api/status`

```json
{
  "db_ready": true,
  "puzzle_count": 4823651,
  "imported_at": "2026-05-18T10:00:00Z",
  "schema_version": 1
}

```

### `GET /api/themes`

Returns the full list of Lichess themes, grouped by category (motifs, mates, phases, lengths, origin, special). Static — baked into the binary, not queried from DB.

```json
{
  "groups": [
    {
      "name": "Tactical motifs",
      "themes": [
        { "id": "fork", "name": "Fork", "description": "..." },
        { "id": "pin", "name": "Pin", "description": "..." }
      ]
    }
  ]
}

```

### `GET /api/openings`

Returns the list of distinct `opening_family` values present in the DB, for the opening filter dropdown.

### `POST /api/puzzle/next`

Body: filter object. Returns: one puzzle.

```json
{
  "puzzle_id": "00sHx",
  "fen": "q3k1nr/1pp1nQpp/...",
  "moves": ["e8d7", "a2e6", "d7d8", "f7f8"],
  "rating": 1760,
  "themes": ["mate", "mateIn2", "middlegame", "short"],
  "side_to_move": "b",
  "solution_length": 3,
  "game_url": "https://lichess.org/...",
  "opening": "Sicilian Defense: Najdorf Variation"
}

```

Note: `themes` is returned but the frontend hides it unless settings allow revealing.

### `GET /api/puzzle/count?...`

Query string carries the filter (URL-encoded). Returns:

```json
{ "count": 12453 }

```

### `POST /api/db/import`

Triggers the download + import pipeline. Returns immediately with a job ID. Progress streamed via `/api/db/import/stream`.

### `GET /api/db/import/stream`

Server-Sent Events stream:

```
event: progress
data: {"stage": "downloading", "bytes": 12345678, "total": 346000000}

event: progress
data: {"stage": "decompressing", "bytes": 800000000}

event: progress
data: {"stage": "inserting", "rows": 1250000, "total_estimate": 5000000}

event: done
data: {"puzzle_count": 4823651, "duration_seconds": 423}

```

### `GET /api/presets` / `POST /api/presets` / `DELETE /api/presets/:id`

CRUD for filter presets.

### `GET /api/config` / `PUT /api/config`

Read/update `config.json`.

---

## 8. First-Run Database Import

### 8.1 Download

- Source: `https://database.lichess.org/lichess_db_puzzle.csv.bz2`
- Size: ~330MB compressed → ~800MB uncompressed CSV.
- Stream the response body directly into a `bzip2.NewReader`, then into a `csv.Reader`. No temp file needed.
- Capture and persist the HTTP `ETag` in `meta.lichess_csv_etag` so a future "Check for updates" can `HEAD` the URL and compare.

### 8.2 Import pipeline

```
HTTP body
  → bzip2.NewReader
  → bufio.Reader
  → csv.NewReader (configured: comma, FieldsPerRecord = 10)
  → batch buffer (10,000 rows)
  → SQLite tx: INSERT INTO puzzles + INSERT INTO puzzle_themes
  → progress callback every batch
  → final COMMIT, then ANALYZE

```

### 8.3 Performance settings during import

Apply on the import connection (not the read connection):

```sql
PRAGMA journal_mode = OFF;
PRAGMA synchronous = OFF;
PRAGMA temp_store = MEMORY;
PRAGMA cache_size = -262144;  -- 256 MB

```

Wrap the whole import in **one transaction**. Expect 5–15 minutes total depending on disk.

After import:

```sql
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
ANALYZE;

```

### 8.4 Resumability

Best-effort: if the import fails mid-way, the user can click "Retry" and it starts over (we DROP TABLE and recreate, since incremental resume is complex and the dataset is reasonably sized). Document this clearly in the UI.

### 8.5 Updates

`Settings → Check for database update`:

1. `HEAD` the Lichess URL, read `ETag`.
2. Compare with `meta.lichess_csv_etag`.
3. If different, prompt the user. On confirm, run the same import (which replaces the DB).

---

## 9. Frontend

### 9.1 Layout

```
┌─────────────────────────────────────────────────────────────┐
│ Header:  Puzzle Trainer       [Presets ▼]  [Settings] [?]   │
├──────────────────┬──────────────────────────────────────────┤
│                  │                                           │
│   Filter panel   │            Chessground board              │
│                  │                                           │
│   ▸ Themes       │                                           │
│   ▸ Rating       │                                           │
│   ▸ Length       │                                           │
│   ▸ Side         │                                           │
│   ▸ Opening      │                                           │
│   ▸ Popularity   │                                           │
│                  │                                           │
│   [or | and]     │   To move: BLACK     Rating: 1620         │
│   12,451 match   │   Themes: ••••• (hidden until solve)      │
│                  │                                           │
│   [Save preset]  │   [Hint] [Solution] [Skip] [Next ▶]       │
└──────────────────┴──────────────────────────────────────────┘

```

### 9.2 Theme picker

- Grouped checkboxes matching Lichess's grouping:
  - **Tactical motifs:** fork, pin, skewer, discoveredAttack, hangingPiece, trappedPiece, attraction, deflection, sacrifice, xRayAttack, zugzwang, interference, intermezzo, quietMove, defensiveMove, advancedPawn, attackingF2F7, capturingDefender, kingsideAttack, queensideAttack, clearance, exposedKing, doubleCheck
  - **Mates:** mate, mateIn1..5, anastasiaMate, arabianMate, backRankMate, bodenMate, doubleBishopMate, dovetailMate, hookMate, smotheredMate, killBoxMate, vukovicMate, etc.
  - **Phases:** opening, middlegame, endgame, rookEndgame, bishopEndgame, pawnEndgame, knightEndgame, queenEndgame, queenRookEndgame
  - **Length:** oneMove, short, long, veryLong
  - **Origin/quality:** master, masterVsMaster, superGM
  - **Special:** equality, advantage, crushing
- Search box at the top filters the visible themes by name.
- "Clear all" / "Select all in group" links.
- AND/OR toggle in the panel header, visually distinct.

### 9.3 Puzzle play loop

This is the part most implementations get wrong. The Lichess CSV's `FEN` is the position **before** the opponent's setup move. So the play loop is:

1. Receive puzzle JSON.
2. Set up board with the given FEN. Show "BLACK to move" / "WHITE to move" based on the **opposite** side of the FEN's active color (because the opponent moves first to set up the puzzle).
3. Wait 400ms, then auto-play the **first move** in the moves array (the opponent's setup). Animate it.
4. Now it's the **user's** turn. The active side after that setup move is the puzzle solver.
5. User makes a move. Compare to the next move in the solution array (UCI format).
  - **Correct:** play move on board, animate. Check if more moves remain. If yes, auto-play the opponent's reply after a short delay. If no, puzzle solved.
  - **Incorrect:** show red highlight on the destination square, "Wrong" indicator, allow retry. Optionally play error sound.
6. On solve: success animation, optionally reveal themes (per setting), auto-load next puzzle after delay (per setting).

**Mate-in-1 exception:** for mate-in-1 puzzles, any move that delivers checkmate should count as correct, not just the move in the solution string. Validate with chess.js: if the user's move results in `in_checkmate()`, accept it.

**Promotion handling:** chessground supports a promotion picker. When the user's move is a pawn reaching the back rank, intercept, show picker, send move with promotion piece.

**Premove:** allow it (Lichess does). If correct in sequence, plays automatically.

### 9.4 Sounds (Lichess-style)

Required audio files in `web/sounds/`:

- `move.mp3` — quiet click for a regular move
- `capture.mp3` — slightly louder for captures
- `check.mp3` — for moves that give check
- `success.mp3` — puzzle solved
- `error.mp3` — wrong move
- `start.mp3` — new puzzle loaded (optional)

Source: rip from Lichess's open-source assets (they're in the lila repo under a permissive license — check the specific files) or use any compatible CC0 set.

### 9.5 Settings panel

Modal accessible from header:

- Sound enabled (toggle)
- Animations enabled (toggle)
- Reveal themes after solve (toggle, default OFF)
- Auto-next puzzle after solve (toggle + delay slider)
- Piece set (dropdown: cburnett, merida, alpha, …)
- Board theme (dropdown: brown, blue, green, gray, …)
- Check for database update (button)
- About / version / data location

### 9.6 Presets panel

Header dropdown shows saved presets. Clicking one applies the filter immediately. Inline:

- Apply
- Rename
- Delete

"Save current filter as preset" button in the filter panel prompts for a name.

---

## 10. Server Startup Sequence

```go
func main() {
    // 1. Parse flags (--port, --data-dir, --no-browser, --headless)
    // 2. Resolve data dir; create if missing.
    // 3. Load config.json (with defaults).
    // 4. Open or create puzzles.db (no import yet).
    // 5. Pick a port: try config.preferred_port, increment if EADDRINUSE.
    //    Bind to 127.0.0.1 only.
    // 6. Start HTTP server in a goroutine.
    // 7. Unless --no-browser: open browser to http://127.0.0.1:<port>
    //    - Linux: xdg-open
    //    - macOS: open
    //    - Windows: rundll32 url.dll,FileProtocolHandler
    // 8. Wait for SIGINT / SIGTERM. Graceful shutdown.
}

```

Browser-open helper handles the common case where the browser is already open — most OSes will focus an existing tab or open a new one. No special handling needed.

---

## 11. Build & Distribution

### 11.1 Cross-compilation

Because the SQLite driver is `modernc.org/sqlite` (pure Go), cross-compilation is straightforward:

```bash
# scripts/build-all.sh
mkdir -p dist

GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o dist/puzzle-trainer-linux-amd64       ./
GOOS=linux   GOARCH=arm64 go build -ldflags="-s -w" -o dist/puzzle-trainer-linux-arm64       ./
GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o dist/puzzle-trainer-macos-amd64       ./
GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o dist/puzzle-trainer-macos-arm64       ./
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/puzzle-trainer-windows-amd64.exe ./

```

`-ldflags="-s -w"` strips debug symbols; binaries land around 18–22MB each.

### 11.2 Asset embedding

```go
// internal/server/assets.go
import "embed"

//go:embed all:web
var webFS embed.FS

//go:embed db/schema.sql
var schemaSQL string

```

The frontend is served as a static file handler against `webFS` (rooted at `web/`).

### 11.3 macOS code signing

Out of scope for v1. Users will see a Gatekeeper warning on first launch (right-click → Open). Document in README.

### 11.4 Windows SmartScreen

Same situation — users may see a SmartScreen warning. Document. Signing certs are expensive; defer.

---

## 12. Open Questions / Future Work

- **Engine analysis after solve.** Embed `stockfish.wasm` or a Go chess engine for post-solve exploration. v2.
- **Solve history / stats.** Add a `solve_history` SQLite table; opt-in toggle. v2.
- **Theme weight in OR mode.** "60% fork, 30% pin, 10% skewer" — would need rejection sampling on top of the random query. v2.
- **NOT (exclude) themes.** "Anything tactical except sacrifices." Simple SQL extension. v2.
- **Color-blind / accessibility pass** on board themes and indicators.
- **Native window via Tauri or webview.** Would lose the "any browser" simplicity but gain a less weird UX (no leftover browser tab on quit). v2.
- **Auto-update.** Self-replacing binary on new release. v2 at earliest; for now manual download from GitHub releases is fine.

---

## 13. Implementation Order (suggested)

1. **Skeleton:** Go HTTP server, embedded static assets, "Hello world" page on `localhost:7777`, browser auto-open.
2. **Paths + config:** OS-specific data dir resolution, config.json load/save.
3. **DB schema + import:** Download a tiny test slice of the Lichess CSV manually (say, 10k puzzles), get import working against a real `.csv` file on disk. Skip network for now.
4. **Filter query builder:** OR mode first; verify `EXPLAIN QUERY PLAN` shows index use.
5. **Frontend skeleton:** Chessground rendering a hard-coded FEN; play one hard-coded puzzle end-to-end.
6. **Wire frontend to backend:** `/api/puzzle/next` working with OR mode.
7. **AND mode + remaining filters** (rating, length, side, opening, popularity).
8. **Theme picker UI** with grouping, search, count badge.
9. **Presets:** CRUD endpoints + frontend dropdown.
10. **Sounds + animations** polish pass.
11. **Real download pipeline:** stream from Lichess URL, SSE progress, setup screen.
12. **Settings panel.**
13. **Cross-compile + release.**

Each step is independently testable. Don't combine steps.