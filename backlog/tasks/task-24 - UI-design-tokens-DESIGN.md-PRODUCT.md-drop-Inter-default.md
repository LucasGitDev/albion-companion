---
id: TASK-24
title: 'UI: design tokens + DESIGN.md/PRODUCT.md (drop Inter default)'
status: In Progress
assignee:
  - '@lucas'
created_date: '2026-11-01 02:40'
updated_date: '2026-11-01 02:43'
labels:
  - ui
  - design
dependencies: []
ordinal: 24000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Revisão de UI apontou que a stack visual (frontend/src/style.css) usa 'Inter' como --font-sans padrão e não há fonte display, um anti-pattern de 'cara de template'. O projeto também não tem DESIGN.md/PRODUCT.md como fonte de verdade de design (templates existem em ~/.claude/templates). Objetivo: estabelecer tokens intencionais e documentá-los para que futuras telas sigam um sistema, não defaults.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 frontend/src/style.css usa um par fonte display + fonte texto intencional (não Inter/Roboto/system-ui como padrão); fallback de sistema mantido
- [ ] #2 Existe DESIGN.md na raiz do projeto com os tokens reais em uso (cores HSL atuais ou migradas, escala tipográfica, raios, sombras, motion) e do's/don'ts
- [ ] #3 Existe PRODUCT.md com público, tom de voz e ação principal por tela (Loot/DPS/Party/Timeline/Settings)
- [ ] #4 make check verde (frontend build inclui)
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Escolher par fonte display+texto não-Inter via @fontsource (bundle offline p/ Wails). 2. Trocar --font-sans em style.css; adicionar --font-display; aplicar display em títulos/hero. 3. Criar DESIGN.md (tokens HSL reais, escala tipo, raios/sombras/motion, do's/don'ts). 4. Criar PRODUCT.md (público, tom, ação por tela). 5. cd frontend && npm run build (ou make check).
<!-- SECTION:PLAN:END -->
