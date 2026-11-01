---
id: TASK-26
title: 'UI: passe de acessibilidade (teclado + contraste AA)'
status: In Review
assignee:
  - '@lucas'
created_date: '2026-11-01 02:40'
updated_date: '2026-11-01 03:12'
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
- [x] #1 Elementos interativos que não são <button>/<a> (ex.: linhas de tabela filtráveis em Loot) são operáveis por teclado (foco + Enter/Space) com role apropriado
- [x] #2 Auditoria de contraste: texto normal ≥ 4.5:1 e texto grande/UI ≥ 3:1; corrigir combinações muted + opacidade + tamanho pequeno que falham
- [x] #3 Nenhuma regressão no focus-visible existente
- [x] #4 make check verde
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Tornar linhas clicáveis (Loot tr onClick) operáveis por teclado: role=button/row + tabIndex + Enter/Space, ou converter em botão. 2. Auditar contraste: muted-foreground com /70-/80 em texto 10-11px vs AA (4.5:1 normal, 3:1 grande/UI); ajustar tokens/opacidades que falham. 3. Preservar focus-visible global. 4. npm run build.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Linha de filtro (Loot) role=button/tabIndex/aria-pressed/onKeyDown; aria-pressed/aria-current em DPS/Timeline/Sidebar. Contraste: --muted-foreground 60%->68% + remoção de /70-/80 em texto essencial; ratios AA computados (ex.: outsider tag 3.71->4.57). focus-visible intacto (diff não toca linhas 97-101). AC1 (teclado) é comportamental: handlers presentes, mas verificação por interação pendente de app.

Revisão visual (ui-reviewer via Playwright + vite/mock @1440): teclado OK mas role=button no <tr> quebrava semântica de tabela ARIA. Corrigido: <tr> volta a ser linha, nome vira <button> (aria-pressed + foco-visível nativo). Build verde. AC1 agora verificado por árvore de acessibilidade (linhas = row, controle = button) + revisão. Nota: desktop-only (ver memória do projeto) — achados de 390px descartados.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
A11y: linha filtrável operável por teclado (handlers), aria-states nos toggles, bump do token muted + fix de opacidade => contraste AA (ratios computados). Verificado: contraste por matemática WCAG, focus-visible intacto por diff, make check verde. AC1 (interação teclado) pendente de verificação com app rodando.
<!-- SECTION:FINAL_SUMMARY:END -->
