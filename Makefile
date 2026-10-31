.PHONY: check test race cli run-replay fixture codes refs tidy dev build

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

tidy:
	go mod tidy

## dev / build: Wails app (after TASK "Wails scaffold" is done)
dev:
	wails dev

build:
	wails build
