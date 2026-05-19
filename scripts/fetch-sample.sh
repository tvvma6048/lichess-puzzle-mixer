#!/usr/bin/env bash
# Refresh testdata/lichess_sample.csv from the official Lichess puzzle export (first 500 puzzles).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${ROOT}/testdata/lichess_sample.csv"
ROWS=501 # header + 500 puzzles

mkdir -p "${ROOT}/testdata"

if ! command -v zstd >/dev/null 2>&1; then
  echo "zstd is required (sudo apt install zstd)"
  exit 1
fi

echo "Downloading ${ROWS} rows to ${OUT} ..."
curl -fsSL 'https://database.lichess.org/lichess_db_puzzle.csv.zst' \
  | zstd -d -q -c \
  | head -n "${ROWS}" > "${OUT}"

echo "Done: $(wc -l < "${OUT}") lines"
