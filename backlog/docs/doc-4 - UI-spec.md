---
id: doc-4
title: UI spec
type: specification
created_date: '2026-10-31T19:20:00Z'
---
# UI spec

## Reference

albiononlinebot.com.br/loot-logger — dark theme, capture controls, live party roster,
map/zone detection, session history, loot report with top looters and guild ranking,
Party Mode toggle. They have **no DPS meter**: that is our differentiator.

## Stack

Wails v2 + React + TypeScript + Vite + Tailwind + shadcn/ui. Icons: lucide-react.
Item icons from render.albiononline.com. Dark theme default, light optional.

## Layout

Window ~1100×720, resizable. Left sidebar nav + top status bar.

**Top bar:** capture button (Start/Stop, with state dot), current zone, local player,
packets/s + decode errors (diagnostic popover), Party Mode toggle (global).

**Tabs (sidebar):**

1. **Loot** (default)
   - Summary cards: total items, silver, pickups, session duration.
   - Top looters table (player, guild, items, silver, last loot) — click filters the feed.
   - Items table (icon, name, qty, looters chips).
   - Live feed (newest first): time, looter, icon+name ×qty, from (mob/player).
   - Actions: Reset session, Export CSV (native save dialog), Include trash toggle.
2. **DPS**
   - Bars per player: damage, DPS, % share, heal/HPS, taken. Sortable.
   - Options: Party only, Players-only targets (PvP), auto-reset on zone change (later).
   - Reset button + combat timer.
   - "Overlay" mode: compact always-on-top window (phase 2).
3. **Party** — roster (name, guild, in-zone indicator), zone.
4. **History** — past sessions (phase 2, needs persistence).
5. **Settings** — interface selection, codes source/version + override file, item DB
   refresh, theme, about/credits (GPL notice).

## Frontend ↔ backend contract

Use `internal/app.Service`. Subscribe to `*:changed` and re-fetch the matching view;
`loot:new` → toast/feed highlight; `capture:status` → top bar.
Permission error (`status.hint == "permission"`) → full-screen helper with the
per-OS instructions from the "Platforms and capture permissions" doc.

## Dev without the game

`wails dev` + a dev flag (`ALBION_COMPANION_REPLAY=testdata/synthetic.aorec`) makes
`Service.Sources` replay the fixture at speed 1.
