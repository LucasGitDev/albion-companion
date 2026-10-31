---
id: decision-1
title: Go engine + Wails v2 + React/shadcn, single binary
date: '2026-10-31T19:20:00Z'
status: accepted
---
## Context
Upstream tools are Windows-only (WPF/.NET) or Node+Electron. The albiondata-client fork already has a Go Protocol18 parser and pcap capture that work on macOS.
## Decision
Go engine reused from the fork, Wails v2 UI with React + shadcn/ui, everything in one process. No Electron, no IPC tunnel.
## Consequences
cgo required (libpcap/Npcap); native build per OS in CI.
