---
id: decision-5
title: Passive capture only
date: '2026-10-31T19:20:00Z'
status: accepted
---
## Decision
Only read packets via libpcap/Npcap. Never send, inject, read process memory or automate input. Same model as StatisticsAnalysis/ao-loot-logger, which SBI tolerates.
