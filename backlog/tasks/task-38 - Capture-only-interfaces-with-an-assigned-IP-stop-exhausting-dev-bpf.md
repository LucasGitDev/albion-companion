---
id: TASK-38
title: Capture only interfaces with an assigned IP (stop exhausting /dev/bpf*)
status: To Do
assignee: []
created_date: '2026-11-01 15:25'
labels:
  - capture
  - platform
  - bug
dependencies: []
ordinal: 38000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Live capture opens EVERY physical interface (getAllPhysicalInterface returns ~24 on this Mac: anpi*, en1-6, bridge*, vmenet*, utun*, awdl0, llw0). pcap.OpenLive grabs the lowest free /dev/bpf* per source, so opening 24 interfaces consumes bpf0..bpfN. grant-capture only chowns the /dev/bpf* that existed when it ran; macOS creates new bpf devices on demand owned by root (crw-------), so the game interface (en0) can land on a root-only device (observed: /dev/bpf5) and fails with 'permission to capture on that device', packets=0 -- even though tcpdump on en0 sees 541 pkts on udp/5056. Fix: only open interfaces with an assigned non-link-local IP, cutting bpf usage to ~1-2. Stays passive. Complements TASK-9 (permanent ChmodBPF LaunchDaemon).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Live sources built only for interfaces with an assigned non-loopback, non-link-local IP; dead interfaces (no inet) skipped
- [ ] #2 On a multi-interface Mac, capture succeeds on the game interface without re-running grant-capture after new bpf devices appear
- [ ] #3 Interface selection change has a test; _darwin/_linux/_windows builds unaffected
<!-- AC:END -->
