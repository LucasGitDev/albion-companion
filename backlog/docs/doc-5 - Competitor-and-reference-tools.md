---
id: doc-5
title: Competitor and reference tools
type: other
created_date: '2026-11-01 02:19'
updated_date: '2026-11-01 02:20'
---
# Competitor and reference tools

Survey of the 12 Albion tools we reference or compete with, each with its exact
link and feature list. For StatisticsAnalysis the features come from the screens
in the source code (the README only names them); for the others, from the README
or the tool's site. Use this as the input for feature gap analysis — see the
"roadmap: what's worth copying" section at the end.

## Main references

### 1. StatisticsAnalysis (Triky313)
C#/.NET, Windows-only, GPL-3.0, actively maintained (as of 2026-10-30).
- https://github.com/Triky313/AlbionOnline-StatisticsAnalysis
- https://triky313.github.io/AlbionOnline-StatisticsAnalysis/

Features:
- Dashboard: fame, silver, respec and might per hour and per session, with a loot summary
- Damage meter: damage, heal, DPS, % per player and per skill, plus its own overlay window
- Dungeon tracker: fame, silver and loot per dungeon, time, chests opened, entry timer
- Loot logger with a comparator (loot vs chest — to find who did not donate)
- Gathering tracker
- Trade monitoring: buys, sells and market fees
- Auction house reading (prices) and item search with price alerts
- Crafting calculator
- Visited-maps map
- Party: members and equipment
- Guild: guild account logs and donations
- Storage history: contents of each chest/bank
- Loss explorer: losses on death
- Player information: player lookup
- Capture via socket (as admin) or Npcap, with IP/port filter for VPN users

### 2. ao-loot-logger (matheussampaio)
Node.js, Windows and Linux, GPL-3.0.
- https://github.com/matheussampaio/ao-loot-logger

Features:
- Records all loot taken from bodies and bags to a file (CSV with `;`)
- Guild and alliance of the looter and of the looted
- Tracks chests and containers (attach/detach, deposit and item movement)
- The file opens in the online viewer below

### 3. AO Loot Logger Viewer
- https://matheus.sampaio.us/ao-loot-logger-viewer
- Web viewer for ao-loot-logger logs (the site itself does not describe features)

### 4. albiononlinebot.com.br Loot Logger
Closed source, Windows 10/11, beta 0.4.11, free.
- https://www.albiononlinebot.com.br/loot-logger

Features:
- Automatic loot logging
- Party Mode (who is not in your group)
- Live party roster
- Map and zone detection
- Session history
- Chest report
- Ranking of top looters and ranking by guild
- No DPS meter

## Mac competitor

### 5. albibong (imjangkar)
Python (Scapy) with a JS frontend, "optimized for Mac", needs sudo, MIT.
- https://github.com/imjangkar/albibong

Features:
- Damage and heal meter: pause/resume, copy the ranking, reset
- Dungeon tracker: name, fame, silver, respec, start, duration and a DPS snapshot; rename, tag tier, filter by type
- Farming tracker: island harvests and animals, with a daily summary
- No loot

## DPS meter

### 6. albion-online-stats (mazurwiktor)
Python + Qt5, Windows/Linux/Mac, MIT/Apache, dormant since 2023-05.
- https://github.com/mazurwiktor/albion-online-stats

Features:
- DPS meter from network traffic
- Monitors only your own party
- No overlay

## Loot loggers

### 7. LootLogger (juniorafonso)
Fork of ao-loot-logger, Node.js, Windows and Linux, GPL-3.0.
- https://github.com/juniorafonso/LootLogger
- Release: https://github.com/juniorafonso/LootLogger/releases/tag/v1.2.2

Features:
- Logs all loot, no whitelist/blacklist
- Killfeed: deaths and kills
- Market log: buy/sell orders, price change, volume, quality and enchantment
- Might log: ranking and gain/loss over time
- Prebuilt executables and automated builds via GitHub Actions
- Online viewer on recckless.com, with loot and market analysis

### 8. Albion-Online-Loot-Logger (TechieDev22)
Windows (Npcap), MIT.
- https://github.com/TechieDev22/Albion-Online-Loot-Logger

Features:
- Captures loot from the game chat and writes it to its own file
- Total value and item count per player
- Charts
- Filters and settings

### 9. AOLootLogHelper (lo0sr)
Python, MIT, archived in 2021.
- https://github.com/lo0sr/AOLootLogHelper

Features:
- Compares the loot log with the chest log to list gear that was not donated after a CTA

## Libraries and protocol

### 10. photon_decode (mazurwiktor)
Rust, MIT/Apache.
- https://github.com/mazurwiktor/photon_decode
- Photon message decoder, likely for the old protocol (16)

### 11. AlbiPy (banex2win)
Python.
- https://github.com/banex2win/AlbiPy
- Market-data sniffer

### 12. albiondata-client (upstream of our fork)
Go, MIT.
- https://github.com/regner/albiondata-client
- Captures market prices and history and sends them to Albion Online Data

**To find more:** the GitHub topics https://github.com/topics/dps-meter and
https://github.com/topics/albion-damage-meter

## Roadmap: what is worth copying (initial notes)

- Loot vs chest comparator: StatisticsAnalysis and AOLootLogHelper
- Dungeon tracker with a DPS snapshot: albibong and StatisticsAnalysis
- Fame and silver per hour: StatisticsAnalysis (we already have fame/hour)
- Killfeed: juniorafonso's fork
- Party Mode and live roster: albiononlinebot (we already have these)
