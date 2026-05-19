#!/usr/bin/env bash
# Full verification: Go checks, release build, Playwright UI smoke.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

export GOTOOLCHAIN="${GOTOOLCHAIN:-local}"
if [[ -x "${HOME}/sdk/go/bin/go" ]]; then
  export GOROOT="${HOME}/sdk/go"
  export PATH="${GOROOT}/bin:${HOME}/go/bin:${PATH}"
fi

# nvm (Node) — same as interactive shells / Cursor agent
export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
if [[ -s "${NVM_DIR}/nvm.sh" ]]; then
  # shellcheck source=/dev/null
  . "${NVM_DIR}/nvm.sh"
fi

echo "==> go vet"
go vet ./...

echo "==> go test"
go test ./...

echo "==> release build"
make release

echo "==> import sample database for e2e"
./bin/lichess-puzzle-mixer --import-csv testdata/lichess_sample.csv --data-dir ./.e2e-data --import-only

if ! command -v npm >/dev/null 2>&1; then
  echo "ERROR: npm is required for e2e tests. Install Node.js or run with nvm."
  exit 1
fi

echo "==> e2e dependencies"
(
  cd e2e
  if [[ -f package-lock.json ]]; then
    npm ci --no-fund --no-audit
  else
    npm install --no-fund --no-audit
  fi
)

echo "==> playwright browser"
(
  cd e2e
  if [[ -n "${CI:-}" ]]; then
    npx playwright install --with-deps chromium
  else
    npx playwright install chromium
  fi
)

echo "==> playwright tests (release binary + embedded UI)"
(
  cd e2e
  npm test
)

echo ""
echo "All checks passed."
