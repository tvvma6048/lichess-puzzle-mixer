# Changelog

All notable changes to [Lichess Puzzle Mixer](https://github.com/DSerejo/lichess-puzzle-mixer) are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.2] - 2026-05-19

### Added

- `scripts/release-notes.sh` — GitHub releases now use the matching section from `CHANGELOG.md` instead of empty auto-generated notes.

### Changed

- Release workflow publishes release body from `CHANGELOG.md` via `body_path`.
- Maintainer docs: release checklist in README and AGENTS.md.

### Fixed

- CI `make verify`: use `CGO_ENABLED=0` for `go vet` / `go test` on Linux when Ayatana AppIndicator dev libraries are not installed (matches `make release` fallback).
- CI visual regression test: snapshot a single theme-group card so font/wrapping differences on GitHub Actions do not fail the build.

## [0.1.1] - 2026-05-18

### Added

- **System tray** on Linux (Ayatana AppIndicator via `github.com/getlantern/systray`), Windows, and macOS (native CGO build on Mac). Menu: Open in browser, Quit.
- **macOS** release binaries in CI: Intel (`macos-amd64`) and Apple Silicon (`macos-arm64`) tarballs.
- Linux **desktop launcher**: `make install-desktop` installs a Freedesktop `.desktop` entry and icon under `~/.local/share`.
- README **screenshots and demo GIF** in `docs/images/`; regenerate with `make readme-images` (Playwright + optional `ffmpeg`).

### Changed

- `make release` on Linux enables CGO when Ayatana AppIndicator dev libraries are present (tray support in release builds).
- README: download table (macOS), tray setup notes, screenshot section, link to this changelog.

### Fixed

- CI: commit `e2e/package-lock.json` so `actions/setup-node` npm cache resolves.

## [0.1.0] - 2026-05-18

First public release ([`v0.1.0`](https://github.com/DSerejo/lichess-puzzle-mixer/releases/tag/v0.1.0)).

### Added

- Desktop app that serves a local web UI on `http://127.0.0.1:7777` (default) with embedded assets in release builds.
- **Combined theme filters**: OR within each theme group, AND between groups (e.g. `(fork or pin) and short`), with theme autocomplete and live match-count preview.
- Filter API: `POST /api/puzzle/count` with JSON body and `GET` with query parameters (including `theme_groups`).
- Puzzle training: Lichess-style board (chessground), hints, wrong-move handling, links back to Lichess.
- **Move history** with step back/forward through the puzzle line and optional **game preamble** when available.
- **Database panel**: import bundled sample (~500 puzzles), download the full Lichess puzzle CSV, or upload `.csv` / `.csv.bz2`; collapses when the database is ready.
- Persistent app data directory and saved preferred port.
- CI (`make verify`: `go vet`, `go test`, release build, Playwright e2e).
- GitHub release workflow with **Linux** and **Windows** amd64 binaries and `SHA256SUMS`.

[Unreleased]: https://github.com/DSerejo/lichess-puzzle-mixer/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/DSerejo/lichess-puzzle-mixer/releases/tag/v0.1.2
[0.1.1]: https://github.com/DSerejo/lichess-puzzle-mixer/releases/tag/v0.1.1
[0.1.0]: https://github.com/DSerejo/lichess-puzzle-mixer/releases/tag/v0.1.0
