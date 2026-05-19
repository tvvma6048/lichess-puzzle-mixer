GO ?= go

.PHONY: dev build release release-windows release-all run-release test verify e2e import-sample fetch-sample install-desktop

LDFLAGS_RELEASE = -s -w
DIST = dist

check-go:
	@command -v $(GO) >/dev/null 2>&1 || { \
		echo "Go is not installed or not on PATH."; \
		echo ""; \
		echo "Pop!_OS / Ubuntu (Go 1.22):"; \
		echo "  sudo apt update && sudo apt install -y golang-go"; \
		echo ""; \
		echo "Or latest from https://go.dev/dl/ (extract to /usr/local/go, add to PATH)."; \
		exit 1; \
	}

dev: check-go
	$(GO) run . --dev --data-dir ./.devdata --no-browser --port 7777

build: check-go
	$(GO) build -o bin/lichess-puzzle-mixer .

# Release binary: embedded web assets, stripped symbols (~smaller).
# On Linux, enables CGO for the system tray (libayatana-appindicator3-dev).
release: check-go
	@mkdir -p bin
	@if [ "$$(uname -s)" = "Linux" ]; then \
		if pkg-config --exists ayatana-appindicator3-0.1 2>/dev/null \
		   || pkg-config --exists appindicator3-0.1 2>/dev/null; then \
			CGO_ENABLED=1 $(GO) build -ldflags="$(LDFLAGS_RELEASE)" -o bin/lichess-puzzle-mixer .; \
		else \
			echo "Note: install libayatana-appindicator3-dev for the system tray icon."; \
			echo "  sudo apt install libayatana-appindicator3-dev gcc pkg-config"; \
			CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS_RELEASE)" -o bin/lichess-puzzle-mixer .; \
		fi; \
	else \
		CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS_RELEASE)" -o bin/lichess-puzzle-mixer .; \
	fi

# Windows amd64 from Linux/macOS (no CGO in this project).
release-windows: check-go
	@mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS_RELEASE)" -o $(DIST)/lichess-puzzle-mixer-windows-amd64.exe .

release-all: check-go
	@chmod +x scripts/build-all.sh
	./scripts/build-all.sh

run-release: release
	./bin/lichess-puzzle-mixer --data-dir ./.devdata --port 7777

# Linux app menu entry (~/.local/share/applications). Search "Lichess Puzzle Mixer" after install.
install-desktop: release
	@chmod +x scripts/install-desktop.sh
	./scripts/install-desktop.sh

test: check-go
	$(GO) vet ./...
	$(GO) test ./...

import-sample: check-go
	$(GO) run . --import-csv testdata/lichess_sample.csv --data-dir ./.devdata --import-only

fetch-sample:
	@chmod +x scripts/fetch-sample.sh
	./scripts/fetch-sample.sh

readme-images: release
	@./bin/lichess-puzzle-mixer --import-csv testdata/lichess_sample.csv --data-dir ./.e2e-data --import-only
	@cd e2e && npm install --no-fund --no-audit && npx playwright install chromium
	cd e2e && CAPTURE_README_ASSETS=1 npx playwright test capture-readme-assets

verify:
	@chmod +x scripts/verify.sh
	./scripts/verify.sh

e2e: release
	@cd e2e && npm install --no-fund --no-audit && npx playwright install chromium && npm test
