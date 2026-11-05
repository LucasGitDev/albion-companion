---
id: TASK-68
title: Fullscreen/large-width layout breaks
status: Done
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 15:07'
updated_date: '2026-11-05 17:07'
labels:
  - ui
dependencies: []
priority: high
ordinal: 68000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
In fullscreen (very large width) the app layout looks bugged. Audit the shell/tab layouts at wide widths (content not constrained, stat strips/grids stretching oddly, tables sprawling) and constrain/align so the app looks intentional at fullscreen as well as 1440px.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 At fullscreen / very wide widths the layout stays intentional (content max-width or sensible stretch; no broken alignment)
- [x] #2 No regressions at 1440px
- [x] #3 ui-reviewer at a wide width (e.g. 1920/2560) + 1440px, no criticals
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Constrained content at fullscreen/ultrawide. main content is now wrapped in a centered mx-auto max-w-[1600px] container (App.tsx) so tabs don't sprawl edge-to-edge; no regression at 1440 (cap not reached, container ~1200px as before). Fixed left-hugging columns that read as broken at wide widths: Settings and Party max-w-2xl columns centered (mx-auto); Route's fixed left column (minmax(0,42rem)) changed to minmax(0,1fr) so the right column anchors. ui-reviewer at 2560/1920/1440: no horizontal overflow, nothing escaping the container, no console errors across 30 tab×width combos; remaining items were balance-only lows. make check EXIT=0.
<!-- SECTION:FINAL_SUMMARY:END -->
