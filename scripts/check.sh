#!/usr/bin/env bash
# Quality gate. Exit 0 = task can move forward (see CLAUDE.md, Definition of Done).
set -euo pipefail
cd "$(dirname "$0")/.."

step() { printf '\n\033[1m==> %s\033[0m\n' "$*"; }

if [ ! -f go.sum ]; then step "go mod tidy (first run)"; go mod tidy; fi

step "gofmt"
unformatted=$(gofmt -l $(git ls-files '*.go' 2>/dev/null || find . -name '*.go' -not -path './refs/*' -not -path './frontend/*'))
if [ -n "$unformatted" ]; then echo "not gofmt'ed:"; echo "$unformatted"; exit 1; fi

step "go build";  go build ./...
step "go vet";    go vet ./...
step "go test";   go test -race ./...

if command -v golangci-lint >/dev/null 2>&1; then
  step "golangci-lint"; golangci-lint run
else
  echo "warn: golangci-lint not installed, skipping"
fi

if [ -f frontend/package.json ]; then
  step "frontend"
  (cd frontend && npm ci && npm run lint --if-present && npx tsc --noEmit && npm run build)
fi

if [ -f wails.json ] && command -v wails >/dev/null 2>&1; then
  step "wails build (native)"; wails build -clean
fi

step "OK"
