GO ?= go

.PHONY: dev build release release-windows release-all run-release test verify e2e import-sample fetch-sample

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
release: check-go
	@mkdir -p bin
	$(GO) build -ldflags="$(LDFLAGS_RELEASE)" -o bin/lichess-puzzle-mixer .

# Windows amd64 from Linux/macOS (no CGO in this project).
release-windows: check-go
	@mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS_RELEASE)" -o $(DIST)/lichess-puzzle-mixer-windows-amd64.exe .

release-all: check-go
	@chmod +x scripts/build-all.sh
	./scripts/build-all.sh

run-release: release
	./bin/lichess-puzzle-mixer --data-dir ./.devdata --port 7777

test: check-go
	$(GO) vet ./...
	$(GO) test ./...

import-sample: check-go
	$(GO) run . --import-csv testdata/lichess_sample.csv --data-dir ./.devdata --import-only

fetch-sample:
	@chmod +x scripts/fetch-sample.sh
	./scripts/fetch-sample.sh

verify:
	@chmod +x scripts/verify.sh
	./scripts/verify.sh

e2e: release
	@cd e2e && npm install --no-fund --no-audit && npx playwright install chromium && npm test
