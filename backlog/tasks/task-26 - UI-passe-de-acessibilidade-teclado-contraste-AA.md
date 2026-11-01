---
id: TASK-26
title: 'UI: passe de acessibilidade (teclado + contraste AA)'
status: In Progress
assignee:
  - '@lucas'
created_date: '2026-11-01 02:40'
updated_date: '2026-11-01 02:52'
labels:
  - ui
  - a11y
dependencies: []
ordinal: 26000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Revisão de UI encontrou: linhas de tabela clicáveis em Loot (tr onClick, ~linha 88) sem suporte a teclado (sem role/tabindex/Enter); uso intenso de text-muted-foreground com opacidade reduzida (/70, /80) em texto pequeno (10–11px uppercase) que provavelmente falha WCAG AA em fundo escuro. Focus-visible global já existe (bom) — manter.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Elementos interativos que não são <button>/<a> (ex.: linhas de tabela filtráveis em Loot) são operáveis por teclado (foco + Enter/Space) com role apropriado
- [ ] #2 Auditoria de contraste: texto normal ≥ 4.5:1 e texto grande/UI ≥ 3:1; corrigir combinações muted + opacidade + tamanho pequeno que falham
- [ ] #3 Nenhuma regressão no focus-visible existente
- [ ] #4 make check verde
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Tornar linhas clicáveis (Loot tr onClick) operáveis por teclado: role=button/row + tabIndex + Enter/Space, ou converter em botão. 2. Auditar contraste: muted-foreground com /70-/80 em texto 10-11px vs AA (4.5:1 normal, 3:1 grande/UI); ajustar tokens/opacidades que falham. 3. Preservar focus-visible global. 4. npm run build.
<!-- SECTION:PLAN:END -->
