# albion-companion

Loot logger + DPS meter for Albion Online that runs on **macOS**, Windows and Linux.

Passive: it only reads the game's network traffic (UDP 5056), like
StatisticsAnalysis and ao-loot-logger. It never touches the game client.

> Status: engine + CLI working against synthetic sessions. Desktop UI (Wails) in progress.

## Download (prebuilt)

CI builds the desktop app for **macOS, Windows and Linux** on every push to
`main` (see `.github/workflows/build.yml`). Grab a build from the **Actions** tab
(artifacts, kept 14 days) or, for a tagged version (`git tag vX.Y.Z && git push
--tags`), from the **Releases** page.

Runtime requirements on the machine that runs the app:

- **Windows**: install [Npcap](https://npcap.com) (default options) — the app
  reads packets through it. Then run `albion-companion.exe`.
- **Linux**: `libpcap`, `libgtk-3` and `libwebkit2gtk-4.1` (e.g.
  `sudo apt install libpcap0.8 libgtk-3-0 libwebkit2gtk-4.1-0`). Grant capture
  permission with `sudo setcap cap_net_raw,cap_net_admin=eip ./albion-companion`.
- **macOS**: unzip the `.app`; give packet-capture permission (see below). The
  build is unsigned, so first launch may need right-click → Open.

## Quick start (dev)

```bash
make test          # unit + end-to-end engine tests
make run-replay    # replay a synthetic fight + loot session
make cli && sudo bin/companion-cli   # live capture (play the game meanwhile)
```

Live capture needs packet-capture permission:

- **macOS**: `sudo`, or give your user read access to `/dev/bpf*`
  (Wireshark's "ChmodBPF" does exactly this).
- **Linux**: `sudo setcap cap_net_raw,cap_net_admin=eip bin/companion-cli`
- **Windows**: install [Npcap](https://npcap.com) (WinPcap compatible mode).

VPN/accelerators (ExitLag, etc.) can hide the traffic from capture.

## Features (roadmap)

- Loot log per player / per item, Party Mode, silver, CSV export compatible with
  [AO Loot Logger Viewer](https://matheus.sampaio.us/ao-loot-logger-viewer)
- DPS / HPS / damage taken meter, party-only and PvP (players-only) modes
- Session history, chest report, item values (Albion Online Data prices)

## Credits

Built on the work of StatisticsAnalysis (Triky313), ao-loot-logger (Matheus Sampaio)
and the Albion Data Project. See [NOTICE](NOTICE). Licensed under GPL-3.0.
