---
id: TASK-59
title: Known-players registry + typeahead for Loot Guard roster input
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 13:09'
updated_date: '2026-11-05 13:17'
labels:
  - app
  - ui
dependencies: []
ordinal: 59000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Loot Guard roster input must keep accepting any free-typed name, but suggest matches from the set of all player nicks ever captured (across sessions). Persist an aggregated known-players list to a new JSON file and feed it from observed players (looters, combat participants, party members). The roster input becomes a typeahead: as the user types, show matching known nicks to pick; typing a brand-new name still works.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A known-players set is persisted (new JSON in the config dir) and accumulates player names across sessions
- [x] #2 Names are harvested from observed players (loot looters, combat rows, party members); case-insensitive de-dup, original casing kept
- [x] #3 Wails binding returns the known-players list for the UI
- [x] #4 Roster input suggests matching known nicks as the user types (substring, case-insensitive), excluding already-added collectors
- [x] #5 Free-typed names not in the registry are still addable
- [x] #6 Suggestion list is keyboard-navigable and accessible (combobox semantics)
- [x] #7 UI follows DESIGN.md; states covered; ui-reviewer at 1440px, no criticals
- [x] #8 Tests cover the registry add/dedup/persist and harvest
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. internal/app/players.go: knownPlayers set (lower->display), load/save <config>/albion-companion/known-players.json (atomic), Add(names...), List() sorted.
2. Service: field known; load in NewService(ConfigDir); feed looter via OnLoot; KnownPlayers() harvests from loot ByPlayer + combat rows + party, merges+persists, returns sorted.
3. app.go binding KnownPlayers() []string.
4. Tests: players_test.go (add/dedup/persist round-trip) + service harvest.
5. Frontend: wails.d.ts+bindings api.knownPlayers(); mock. RosterEditor input → accessible combobox typeahead (substring filter, exclude already-added, keyboard arrows/enter/esc, click pick, free text still addable).
6. ui-reviewer 1440px, fix, make check, finalize.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verified: 4 Go table tests (add/dedup case-insensitive, persist round-trip, harvest from loot/combat/party views, OnLoot feed) + make check EXIT=0. Registry internal/app/players.go persists <config>/albion-companion/known-players.json; fed live via OnLoot looter + lazily harvested from loot ByPlayer/combat Snapshot/party/local player on KnownPlayers(). Frontend: accessible combobox (role=combobox/listbox/option, aria-activedescendant, arrow/enter/esc, scrollIntoView, IME guard, substring-match bold). ui-reviewer: 2 rounds, ended 'sem críticos' (fixed: active-option scroll into view, active-state AA contrast bg-primary/primary-foreground, bg-popover+shadow-xl, new-name + duplicate hints with aria-live). Desktop-only (390px ignored).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added a known-players registry that accumulates every observed nick across sessions (internal/app/players.go, persisted to known-players.json; fed from loot looters live + harvested from loot/combat/party views) and turned the Loot Guard roster input into an accessible typeahead combobox that suggests matching known nicks (bold substring match, keyboard nav, scroll-into-view) while still accepting any free-typed name. Binding KnownPlayers(). Verified with 4 table tests, make check EXIT=0, and 2 ui-reviewer rounds ending with no criticals at 1440px.
<!-- SECTION:FINAL_SUMMARY:END -->
