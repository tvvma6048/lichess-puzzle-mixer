# Lichess Puzzle Mixer

[![CI](https://github.com/DSerejo/lichess-puzzle-mixer/actions/workflows/ci.yml/badge.svg)](https://github.com/DSerejo/lichess-puzzle-mixer/actions/workflows/ci.yml)
[![Release](https://github.com/DSerejo/lichess-puzzle-mixer/actions/workflows/release.yml/badge.svg)](https://github.com/DSerejo/lichess-puzzle-mixer/actions/workflows/release.yml)

**Combine multiple Lichess puzzle themes in one training set** — the main thing this app is for. On Lichess you pick a single motif per session; here you can stack themes (e.g. fork + pin, or mate + endgame) and narrow with rating, length, and popularity. Everything runs locally in your browser after you import the [Lichess puzzle database](https://database.lichess.org/#puzzles).

Inspired by [offline-chess-puzzles](https://github.com/brianch/offline-chess-puzzles), with combined-theme search as the focus and a Lichess-like board (chessground).

Thanks to Lichess for the puzzle dump, and to [chessground](https://github.com/lichess-org/chessground) and [chess.js](https://github.com/jhlywa/chess.js) for the board and rules.

## Download

Pre-built binaries are on the **[Releases](https://github.com/DSerejo/lichess-puzzle-mixer/releases/latest)** page:

| Platform | File |
|----------|------|
| Linux (64-bit) | `lichess-puzzle-mixer-linux-amd64.tar.gz` |
| Windows (64-bit) | `lichess-puzzle-mixer-windows-amd64.zip` |

Each archive contains a single executable. Extract it, then run it — your browser should open automatically.

**Linux**

```bash
tar xzf lichess-puzzle-mixer-linux-amd64.tar.gz
chmod +x lichess-puzzle-mixer-linux-amd64
./lichess-puzzle-mixer-linux-amd64
```

**Windows**

Unzip `lichess-puzzle-mixer-windows-amd64.zip` and double-click `lichess-puzzle-mixer-windows-amd64.exe`, or run it from a terminal. Windows may show a SmartScreen prompt the first time; choose “More info” → “Run anyway” if you trust this build.

## First run

1. Start the app. It listens on `http://127.0.0.1:7777` by default and opens your browser.
2. On the setup screen, open **Database** and either:
   - **Import sample (~500)** — quick start, no download, or
   - **Download from Lichess** — full database (~330 MB compressed; needs internet and disk space), or
   - **Upload CSV** — use a `lichess_db_puzzle.csv` or `.csv.bz2` file you already have from [database.lichess.org](https://database.lichess.org/#puzzles).
3. Build **theme groups** (see below), add any other filters, then click **Start training →**.

Your database and settings are stored under the app data folder (see below), not next to the executable.

## Combining themes

Each **group** is OR: the puzzle needs at least one theme from that group. **Groups** are ANDed together.

| What you want | How to set it up |
|---------------|------------------|
| fork **or** pin | One group: `fork`, `pin` |
| fork **and** pin | Two groups: `fork` · `pin` |
| **(fork or pin) and short** | Group 1: `fork`, `pin` — then **+ Add AND group** — Group 2: `short` |

Use **Preview count** to see how many puzzles match before you train. Rating, side to move, popularity, and length filters apply on top.

## How to play

- Make moves by **clicking** a piece then a square, or by **dragging**, like on Lichess.
- **Get a hint** highlights the piece, then the target square.
- **View solution** plays the full line without counting as a solve.
- The sidebar shows **move history** with restart / step back / forward / last.
- When a puzzle has a source game link, the app loads **moves from that game** before the puzzle position (shown in grey in the history) so you can step through the game line.
- **Puzzle on Lichess** opens the training page for the current puzzle.

Wrong moves are rejected; when you finish the line, the next puzzle loads after you click **Next puzzle**.

## Features

- **Combine multiple themes** with OR inside groups and AND between groups (the reason to use this app)
- Also filter by **rating range**, **side to move**, **minimum popularity**, and **puzzle length**
- **Preview count** before you start training
- **Local SQLite database** — import sample, upload CSV, download from Lichess, or clear and re-import
- **Hints** and **view solution**
- **Move history** with keyboard-style navigation controls
- **Source game preamble** in the history (when Lichess provides a game URL)
- **Links** to the puzzle and original game on Lichess
- Single static binary; no Node, Rust, or Python required to run releases

## Data folder

| OS | Default location |
|----|------------------|
| Linux | `~/.config/lichess-puzzle-mixer/` |
| Windows | `%AppData%\lichess-puzzle-mixer\` |
| macOS | `~/Library/Application Support/lichess-puzzle-mixer/` |

Contains `puzzles.db` and `config.json`. Override with `--data-dir /path/to/dir`.

## Command-line options

```
  -dev              Serve web files from ./web (development only)
  -data-dir string  App data directory (default: OS path above)
  -port int         HTTP port (default 7777)
  -no-browser       Do not open a browser on startup
  -import-csv path  Import a Lichess CSV (.csv or .csv.bz2), then start
  -import-only      Exit after -import-csv
```

Examples:

```bash
./lichess-puzzle-mixer-linux-amd64 --port 8080
./lichess-puzzle-mixer-linux-amd64 --no-browser
./lichess-puzzle-mixer-linux-amd64 --import-csv ~/Downloads/lichess_db_puzzle.csv.bz2 --import-only
```

## Building from source

Requires [Go 1.26+](https://go.dev/dl/).

```bash
git clone https://github.com/DSerejo/lichess-puzzle-mixer.git
cd lichess-puzzle-mixer
make release          # bin/lichess-puzzle-mixer
make release-all      # dist/ for Linux, Windows, macOS (amd64/arm64)
make dev              # dev server with live ./web assets on :7777
```

Run tests and end-to-end checks:

```bash
make verify
```

## Creating a release (maintainers)

Tag a version and push; GitHub Actions builds Linux and Windows assets and attaches them to the release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Or run the **Release** workflow manually from the Actions tab.

## Possible use cases

- Build a custom mix: e.g. **fork OR pin** at 1500–1800, or **mateIn2 AND endgame** only
- Practice tactics **offline** after importing the full database
- Teach one motif, then tighten with AND (e.g. **deflection AND sacrifice**)
- Review the **real game** leading up to a puzzle when preamble is available

Feedback and bug reports are welcome in [Issues](https://github.com/DSerejo/lichess-puzzle-mixer/issues).

## License

MIT — see [LICENSE](LICENSE).

### Third-party assets

- Board and pieces use **chessground** (Lichess) and the **cburnett** piece set (CC BY-SA 3.0).
- Puzzle data is from the [Lichess open database](https://database.lichess.org/) (ODbL for the dataset; check Lichess terms for redistribution).
