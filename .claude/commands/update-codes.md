---
description: Regenerate Albion event/operation codes after a game patch and verify decoders
---

The game was patched and decoding may be broken (symptoms: `unknown` count rising,
`[decode-error]` lines, empty loot/DPS).

1. `make codes` (pulls StatisticsAnalysis and regenerates `internal/protocol/codes.json`).
2. `git diff --stat internal/protocol/codes.json` and summarize which of the names in
   `internal/events/events.go` changed value.
3. For every event we decode, diff its SA `*Event.cs` file against the field indices
   in `internal/events/events.go` (`git -C refs/StatisticsAnalysis log -p -- <file>`).
   Update decoders + tests if a layout changed.
4. `make check`.
5. If the user has a fresh `.aorec` from after the patch, replay it with
   `go run ./cmd/companion-cli -replay <file> -every 0` and confirm decodeErr=0.
6. Commit as `fix(protocol): update codes for <patch name/date>`.
