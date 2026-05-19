#!/usr/bin/env bash
# Print the CHANGELOG.md section for a release tag (e.g. v0.1.1 → ## [0.1.1]).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TAG="${1:?usage: release-notes.sh <tag> [changelog-file]}"
CHANGELOG="${2:-$ROOT/CHANGELOG.md}"
VERSION="${TAG#v}"

if [[ ! -f "$CHANGELOG" ]]; then
  echo "release-notes: $CHANGELOG not found" >&2
  exit 1
fi

content="$(
  awk -v ver="$VERSION" '
    /^## \[/ {
      if (found) exit
      if ($0 ~ "\\[" ver "\\]") found = 1
      next
    }
    found { print }
  ' "$CHANGELOG"
)"

if [[ -z "${content//[[:space:]]/}" ]]; then
  echo "release-notes: no section ## [$VERSION] in $CHANGELOG" >&2
  exit 1
fi

printf '%s\n' "$content" | sed -e :a -e '/^\n*$/{$d;N;ba' -e '}'
