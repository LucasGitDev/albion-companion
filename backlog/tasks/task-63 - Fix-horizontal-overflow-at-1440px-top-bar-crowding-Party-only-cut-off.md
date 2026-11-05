---
id: TASK-63
title: Fix horizontal overflow at 1440px (top bar crowding / Party only cut off)
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 14:51'
updated_date: '2026-11-05 14:55'
labels:
  - ui
dependencies: []
ordinal: 63000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
At 1440px the page shows a horizontal scrollbar and the top bar's right-most control (the 'Party only' toggle) is clipped. The header is a single non-wrapping flex row (capture button, session chip, zone/player chips, fame chip, diagnostics, party toggle) whose intrinsic width can exceed the viewport; chips don't shrink. Make the top bar fit any supported width without horizontal page scroll (shrink/truncate the flexible chips, or wrap), and confirm no element is clipped.
<!-- SECTION:DESCRIPTION:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Could NOT reproduce a horizontal scrollbar at 1440px: measured scrollWidth==clientWidth (overflow 0) on every tab (Loot/DPS/Dungeons/Gathering/Route/Timeline/Loot Guard/History), and the Party-only toggle's right edge was 1424 < 1440 (not clipped). The app root is already overflow-hidden, so the page can't scroll horizontally. Applied defensive hardening so a long zone/session name can never push controls off: shrink-0 on capture button, fame chip, diagnostics wrapper, Party-only toggle; min-w-0 on header + zone/player Chip + SessionControl (button min-w-0 max-w-[28ch]); title on the Chip value; shrink-0 on the Chip icon. Stress-tested with an injected 60-char zone name: overflow still 0, toggle right still 1424 (chip truncates). make check EXIT=0. ui-reviewer: no criticals; 2 mediums fixed (Chip title, SessionControl shrink).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Hardened the top bar against horizontal overflow. The reported scrollbar/clipping was not reproducible at 1440px (measured zero overflow on all tabs; root is overflow-hidden), but the header lacked shrink guards so a long zone/session name could clip the Party-only toggle. Added shrink-0 to the fixed controls and min-w-0 + truncation to the flexible chips (zone/player/session), plus a title on the zone chip. Verified by measurement incl. a 60-char zone-name stress test (overflow stays 0, nothing clipped) and make check EXIT=0.
<!-- SECTION:FINAL_SUMMARY:END -->
