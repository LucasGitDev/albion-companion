.PHONY: check test race cli run-live run-replay fixture codes refs tidy dev build run run-dev run-replay-app grant-capture

# Resolve the wails CLI whether or not GOPATH/bin is on PATH.
WAILS := $(shell command -v wails 2>/dev/null || echo "$(shell go env GOPATH)/bin/wails")
APP_BIN := build/bin/albion-companion.app/Contents/MacOS/albion-companion

## check: quality gate (Definition of Done) — see scripts/check.sh
check:
	scripts/check.sh

test:
	go test ./...

race:
	go test -race ./...

## cli: build the headless dev tool
cli:
	go build -o bin/companion-cli ./cmd/companion-cli

## run-live: capture live and dump loot/DPS (needs capture permission)
run-live: cli
	sudo bin/companion-cli

## fixture: regenerate the synthetic session used for UI/CLI dev
fixture:
	go run ./tools/mkfixture -out testdata/synthetic.aorec

## run-replay: replay the synthetic session through the engine
run-replay: fixture
	go run ./cmd/companion-cli -replay testdata/synthetic.aorec -every 0

## refs: clone upstream reference projects into refs/ (gitignored)
refs:
	scripts/fetch-refs.sh

## codes: regenerate internal/protocol/codes.json from StatisticsAnalysis (after a game patch)
codes: refs
	git -C refs/StatisticsAnalysis pull --ff-only
	go run ./tools/gencodes -sa refs/StatisticsAnalysis -out internal/protocol/codes.json

## itemnames: regenerate internal/itemdb/localized.tsv from ao-bin-dumps items.json (after a game patch)
itemnames:
	go run ./tools/genlocalized -out internal/itemdb/localized.tsv

tidy:
	go mod tidy

## dev / build: Wails app
dev:
	$(WAILS) dev

build:
	$(WAILS) build

## grant-capture: let your user read /dev/bpf* so the GUI can capture (macOS, until reboot)
grant-capture:
	sudo scripts/grant-capture.sh

## run: build and launch the app for a REAL live-capture test (needs 'make grant-capture' first).
## Open Albion, click Start capture; watch packets/decodeErr in the diagnostics popover.
run: build
	@echo "Launching Albion Companion. If it shows the permission screen, run 'make grant-capture' and retry."
	$(APP_BIN)

## run-dev: app with hot-reload (frontend edits live). Also needs capture permission.
run-dev:
	$(WAILS) dev

## run-replay-app: launch the GUI replaying a recording — no game or permission needed.
## Great to see the UI fill with real decoded data from the captured fixture.
run-replay-app: build
	ALBION_COMPANION_REPLAY=testdata/party-dungeon.aorec $(APP_BIN)
