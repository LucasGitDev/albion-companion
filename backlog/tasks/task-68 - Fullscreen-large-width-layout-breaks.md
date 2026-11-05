---
id: TASK-68
title: Fullscreen/large-width layout breaks
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-05 15:07'
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
- [ ] #1 At fullscreen / very wide widths the layout stays intentional (content max-width or sensible stretch; no broken alignment)
- [ ] #2 No regressions at 1440px
- [ ] #3 ui-reviewer at a wide width (e.g. 1920/2560) + 1440px, no criticals
<!-- AC:END -->
