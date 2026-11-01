---
id: TASK-38
title: Capture only interfaces with an assigned IP (stop exhausting /dev/bpf*)
status: In Progress
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 15:25'
updated_date: '2026-11-01 15:28'
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Root-caused live from a user test: tcpdump saw 541 pkts on en0:5056 but companion-cli -iface en0 failed with /dev/bpf5 Permission denied; getAllPhysicalInterface returned 24 ifaces, exhausting the grant-capture'd bpf pool. Added ifaceHasUsableIP (shared) requiring a non-loopback/non-link-local/non-unspecified assigned IP; applied in _darwin and _linux selectors (also fixed darwin FlagUp check from '==1' to '!=0' to match linux). Windows uses the IpAdapterAddresses path, left unchanged (own follow-up). Result on this Mac: 24 -> 4 interfaces (en0 + 3 VM bridges); live capture now packets=279 decodeErr=0 with no permission error, no grant-capture re-run. Extracted isUsableCaptureIP as a pure predicate with a table test + a host-enumeration sanity test.
<!-- SECTION:NOTES:END -->
