# Agent guide — Lichess Puzzle Mixer

## Before finishing any task

Run the full verification suite from the repo root:

```bash
make verify
```

Do **not** claim work is complete until this passes. If a step fails, fix it and re-run.

## What `make verify` does

| Step | Checks |
|------|--------|
| `go vet ./...` | Static analysis |
| `go test ./...` | API + embedded asset HTTP tests |
| `make release` | Production binary builds (`bin/lichess-puzzle-mixer`) |
| Playwright (`e2e/`) | Real browser: page load, `/api/status` UI, CSS applied |

E2E uses the **release** binary (embedded `web/`), not `--dev` mode. That matches what users run.

## Changing the frontend

- Prefer stable selectors: `data-testid="api-status"` (see `web/index.html`).
- After editing `web/*.js`, `web/*.html`, or `web/*.css`, run `make verify`.
- If you change visible text or layout, update `e2e/smoke.spec.ts` (and visual baselines if needed).

### Visual (screenshot) tests

`e2e/visual.spec.ts` compares the success-state UI to committed PNG baselines.

- **Intentional UI change:** update baselines:
  ```bash
  cd e2e && npx playwright test visual --update-snapshots
  ```
- Use screenshots for **layout/color** regressions; use `data-testid` + text assertions for **behavior**.
- Snapshots are pinned to Playwright’s viewport (900×700) in `playwright.config.ts` to reduce cross-machine drift.

## Changing the backend

- Add or extend tests in `internal/server/*_test.go` or `main_test.go` for new routes and embed behavior.
- Run `go test ./...` after handler changes.
- Sample puzzles: `testdata/lichess_sample.csv` (500 rows). Refresh with `make fetch-sample`.
- Import locally: `make import-sample` (writes `./.devdata/puzzles.db`).

## Quick checks (partial)

```bash
make test          # Go tests only
make release       # Build only
cd e2e && npm test # UI only (requires `make release` first)
```

## Requirements

- Go 1.26+ on `PATH` (`~/sdk/go/bin/go` or system install)
- Node.js + npm (for Playwright)
- First e2e run downloads Chromium (~150MB)

## Cross-compile

```bash
make release-windows   # dist/lichess-puzzle-mixer-windows-amd64.exe
make release-all       # all platforms in dist/
```
