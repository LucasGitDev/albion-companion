---
id: decision-4
title: 'UI contract: pull-on-notify with throttled topic events'
date: '2026-10-31T19:20:00Z'
status: accepted
---
## Decision
internal/app.Service exposes views (LootView, CombatView, PartyView) as bound methods. Engine changes emit 'topic:changed' (throttled 250ms + Flush ticker); the frontend re-fetches. Only loot:new carries a payload.
## Why
Small payloads, no state drift between Go and React, protects the WebView during big fights.
