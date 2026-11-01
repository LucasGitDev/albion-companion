---
id: TASK-1
title: Validate engine on macOS with a live game session
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-01 01:33'
labels:
  - capture
  - core
dependencies: []
priority: high
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
go mod tidy (go.sum is not committed yet), make check, then run the CLI live while playing: sudo go run ./cmd/companion-cli -record refs/session-1.aorec -dump -only Join,OtherGrabbedLoot,HealthUpdate,HealthUpdates,PartyJoined. Compare params with docs/Protocol spec.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 make check green on macOS
- [x] #2 Live capture shows Join + loot + damage with decodeErr=0
- [x] #3 Silver FixPoint scale confirmed (or fixed) against in-game value
- [x] #4 Mismatches recorded in task notes / Protocol spec
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. go mod tidy (generate go.sum) on macOS.
2. make check; fix macOS breakage (cgo/libpcap, build tags, vet) without behavior change.
3. make run-replay; confirm decodeErr=0, loot+DPS tables filled.
4. make refs.
5. Build CLI; give user exact sudo record command for refs/session-1.aorec. Wait.
6. On user signal: replay recording, compare params to Protocol spec + refs/StatisticsAnalysis, fix divergences with tests. Verify silver FixPoint scale vs in-game value.
7. Record divergences in notes + update Protocol spec. make check, commit, move to In Review.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
go.sum generated via go mod tidy. make check GREEN on macOS unchanged — cgo/libpcap compiled clean, no build-tag/vet fixes needed (golangci-lint not installed → skipped). run-replay: packets=296 msgs=296 unknown=0 decodeErr=0; loot+DPS tables filled (AC#1 ✓, AC#2 partial: replay only). refs cloned: StatisticsAnalysis@8ad2067, ao-loot-logger@854ea44. CLI built, all 6 -only names valid. Awaiting live capture.

Live capture refs/session-1.aorec: 80268 packets / 159938 msgs. Initial replay decodeErr=97, unknown=0. Divergences vs decoders (all confirmed against refs/StatisticsAnalysis, fixed with table tests):
1. HealthUpdate Change (param 2) is OPTIONAL — omitted on zero-delta health syncs (83x). SA HealthUpdateEvent reads via TryGetValue->0. Decoder made it mandatory. Fixed.
2. HealthUpdate CauserID (param 6) is OPTIONAL — server/env damage (EffectOrigin ServerAuthority=8) has no causer (12x). SA TryGetValue->0. Fixed.
3. PartyJoined GUIDs (param 8) is a FLAT []byte (Nx16 concatenated), not array-of-arrays (SA GetPartyUserGuidsFromBytes); absent 8/9 -> empty party, no error (2x). protocol.Params.GUIDs now chunks flat []byte/[]int8. Side effect: party now populates so party-only DPS/loot shows all members (was only local player).
4. Silver FixPoint /10000 CONFIRMED vs in-game: local player session total ~57652 matched user's in-game silver. Scale unchanged. (Raw param5 is fractional on wire, e.g. 1312000/10000=131.2; display rounds.)
After fixes: replay decodeErr=0, unknown=0; DPS+loot tables populated for all 6 party members. Protocol spec (doc-2) updated.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Validated the engine on macOS against a live game session. go mod tidy generated go.sum; make check green (cgo/libpcap compiled clean, no platform fixes needed). Live capture (refs/session-1.aorec, 80268 pkts) initially had decodeErr=97 from three decoder divergences, all confirmed vs refs/StatisticsAnalysis and fixed with table tests: HealthUpdate Change(2) and CauserID(6) are optional; PartyJoined GUIDs(8) is a flat []byte of concatenated 16-byte GUIDs (fixing this also populated the party so party-only DPS/loot now lists all members). Silver FixPoint /10000 confirmed against the in-game total (~57652). Post-fix replay: decodeErr=0, unknown=0, DPS+loot tables populated. Protocol spec (doc-2) updated. Verified via go test -race (make check) and CLI replay summary. Commit c81a5d0.
<!-- SECTION:FINAL_SUMMARY:END -->
