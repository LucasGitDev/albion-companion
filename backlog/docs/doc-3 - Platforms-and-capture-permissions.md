---
id: doc-3
title: Platforms and capture permissions
type: guide
created_date: '2026-10-31T19:20:00Z'
---
# Platforms and capture permissions

## Capture requirements per OS

| OS | Library | Permission | Notes |
|---|---|---|---|
| macOS | libpcap (ships with macOS) | read on `/dev/bpf*` | sudo works for the CLI; the app must not run as root |
| Linux | libpcap (`libpcap-dev` to build) | CAP_NET_RAW + CAP_NET_ADMIN | `setcap` on the binary |
| Windows | Npcap (WinPcap-compatible mode) | Npcap installed | interface names are `\Device\NPF_{GUID}` |

Build: cgo is required (gopacket/pcap). Cross-compiling needs the target SDK; release
builds run natively per OS in CI.

## macOS strategy (desktop app)

1. Detect: try `pcap.OpenLive` on the first interface; on `capture.ErrPermission`
   the UI shows a "Grant capture permission" screen (Status.Hint = "permission").
2. Fix (one-time, admin password): install a LaunchDaemon that, at boot, sets
   `/dev/bpf*` group to `access_bpf` and adds the user to that group — same approach
   as Wireshark's ChmodBPF. Run it via `osascript -e 'do shell script ... with administrator privileges'`.
3. If Wireshark is already installed, ChmodBPF is already there → just works.
4. Never run the GUI as root (WebView + root = bad idea).

## Gotchas

- VPN / ExitLag / GeForce Now: traffic goes through a tunnel, pcap on the physical NIC
  sees nothing. ExitLag users can switch to NDIS (Legacy) redirection on Windows.
- macOS touchbar NIC (`ac:de:48:00:11:22`) and VM NICs are filtered out.
- Apple Silicon running the game natively: fine. Under CrossOver/Whisky: still the
  same host NIC, should work — verify.
- Wails on Windows needs WebView2 (bundled bootstrapper in installer).
