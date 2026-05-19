#!/usr/bin/env bash
# Install launcher entry for GNOME/KDE/etc. (Super key → app search).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN_SRC="${BIN_SRC:-$ROOT/bin/lichess-puzzle-mixer}"
INSTALL_BIN="${INSTALL_BIN:-$HOME/.local/bin/lichess-puzzle-mixer}"
DESKTOP_DIR="$HOME/.local/share/applications"
ICON_DIR="$HOME/.local/share/icons/hicolor/scalable/apps"
DESKTOP_SRC="$ROOT/packaging/linux/lichess-puzzle-mixer.desktop"
ICON_SRC="$ROOT/packaging/linux/lichess-puzzle-mixer.svg"

if [[ ! -x "$BIN_SRC" ]]; then
  echo "Binary not found: $BIN_SRC" >&2
  echo "Run: make release" >&2
  exit 1
fi

mkdir -p "$(dirname "$INSTALL_BIN")" "$DESKTOP_DIR" "$ICON_DIR"
install -m 755 "$BIN_SRC" "$INSTALL_BIN"
install -m 644 "$ICON_SRC" "$ICON_DIR/lichess-puzzle-mixer.svg"

sed "s|@EXEC@|$INSTALL_BIN|g" "$DESKTOP_SRC" >"$DESKTOP_DIR/lichess-puzzle-mixer.desktop"
chmod 644 "$DESKTOP_DIR/lichess-puzzle-mixer.desktop"

if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database "$HOME/.local/share/applications" 2>/dev/null || true
fi
if command -v gtk-update-icon-cache >/dev/null 2>&1; then
  gtk-update-icon-cache -f -t "$HOME/.local/share/icons/hicolor" 2>/dev/null || true
fi

echo "Installed:"
echo "  Binary:  $INSTALL_BIN"
echo "  Launcher: $DESKTOP_DIR/lichess-puzzle-mixer.desktop"
echo ""

if ldd "$INSTALL_BIN" 2>/dev/null | grep -q ayatana-appindicator; then
  echo "System tray: enabled (Ayatana AppIndicator)."
else
  echo "System tray: NOT in this binary. Rebuild after:"
  echo "  sudo apt install libayatana-appindicator3-dev gcc pkg-config"
  echo "  make release && make install-desktop"
fi

echo ""
echo "Open your app menu and search for \"Lichess Puzzle Mixer\" (or \"chess\" / \"puzzle\")."
echo "On GNOME/Pop!_OS, enable the AppIndicator extension if the tray icon is missing."
echo "Ensure ~/.local/bin is on your PATH if you want to run it from a terminal."
