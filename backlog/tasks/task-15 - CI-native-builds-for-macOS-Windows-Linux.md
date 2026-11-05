---
id: TASK-15
title: 'CI: native builds for macOS, Windows, Linux'
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-10-31T19:20:00Z'
updated_date: '2026-11-05 18:55'
labels:
  - build
  - ci
dependencies:
  - TASK-3
priority: low
ordinal: 15000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
GitHub Actions matrix: make check + wails build per OS (libpcap-dev on Linux, Npcap SDK on Windows). Upload artifacts; release on tag. Unsigned is fine for now (document Gatekeeper/SmartScreen workaround).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Tag produces 3 downloadable builds
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
GitHub Actions CI (.github/workflows/build.yml): a test gate (go vet + go test -race with libpcap) then a native build matrix for Linux/Windows/macOS. Each OS installs its capture + GUI deps (Linux: libgtk-3/libwebkit2gtk-4.1/libpcap, built with -tags webkit2_41; Windows: Npcap SDK 1.13 wired via CGO_CFLAGS/LDFLAGS for gopacket/pcap; macOS: native), installs Wails v2.17.0, builds with CGO, and uploads the binary as an artifact (macOS .app zipped). A release job attaches all three to a GitHub Release on a v* tag. README gained a 'Download (prebuilt)' section with per-OS runtime deps (Npcap on Windows, libpcap+gtk/webkit on Linux) and capture-permission notes. Verified locally: wails build (macOS) succeeds producing build/bin/albion-companion.app; YAML validated; go test/vet green. Windows/Linux CI steps are standard but first exercised on push.
<!-- SECTION:FINAL_SUMMARY:END -->
