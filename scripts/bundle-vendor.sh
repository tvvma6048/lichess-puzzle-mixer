#!/usr/bin/env bash
# Rebuild vendored chessground bundle after npm install.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
npm install chessground@9.2.1 --no-save --silent
npx esbuild node_modules/chessground/dist/chessground.js \
  --bundle --format=esm \
  --outfile=web/vendor/chessground.bundle.js
cp node_modules/chessground/assets/chessground.cburnett.css web/vendor/
echo "Wrote web/vendor/chessground.bundle.js and chessground.cburnett.css"
