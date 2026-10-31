#!/usr/bin/env bash
# Clones the upstream projects we port from into refs/ (gitignored).
# Agents: grep these before guessing any packet field — see CLAUDE.md.
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p refs
clone() { [ -d "refs/$2/.git" ] && git -C "refs/$2" pull -q --ff-only || git clone -q --depth 1 "$1" "refs/$2"; echo "refs/$2 @ $(git -C refs/$2 rev-parse --short HEAD)"; }
clone https://github.com/Triky313/AlbionOnline-StatisticsAnalysis StatisticsAnalysis
clone https://github.com/matheussampaio/ao-loot-logger ao-loot-logger
