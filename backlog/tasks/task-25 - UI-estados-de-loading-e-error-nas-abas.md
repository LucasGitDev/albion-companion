---
id: TASK-25
title: 'UI: estados de loading e error nas abas'
status: In Review
assignee:
  - '@lucas'
created_date: '2026-11-01 02:40'
updated_date: '2026-11-01 03:16'
labels:
  - ui
  - frontend
dependencies: []
ordinal: 25000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Revisão de UI: chamadas de api.* (useLoot/useCombat/useStatus, exportLootCSV, resetLoot) usam .then sem .catch — falha de binding fica silenciosa e não há estado de erro. Loading é confundido com empty (ex.: Loot mostra 'No loot recorded yet' enquanto ainda carrega; só Settings mostra 'loading…'). O fluxo de UI exige loading/empty/error explícitos.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Estado de loading distinto de empty nas abas Loot, DPS, Party, Timeline (skeleton ou indicador enquanto o primeiro dado não chegou)
- [ ] #2 Chamadas de api.* tratam erro (catch) e a UI mostra estado de erro com ação de retry onde fizer sentido
- [ ] #3 Export CSV e Reset dão feedback de falha, não só de sucesso
- [x] #4 make check verde
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Mapear hooks (useLoot/useCombat/useStatus/useTopic) e chamadas api.*. 2. Distinguir loading de empty: expor estado (undefined=loading vs vazio=empty) e render skeleton/indicador em Loot/DPS/Party/Timeline. 3. Adicionar .catch nas chamadas api.* + estado de erro com retry. 4. Export CSV e Reset: feedback de falha. 5. npm run build.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Padrão Async<T>{data,status,retry} em useTopic; useLoot/useCombat delegam. states.tsx (Skeleton/Loading/ErrorState+retry). Loot/DPS/Party/Timeline gated. Export/Reset flasham falha. AC1-3 são comportamentais: código presente e make check verde (exit 0), mas verificação visual/interativa (loading vs empty, erro+retry) pendente de app Wails rodando + ui-reviewer — não disponível nesta sessão.

Merged em main (12b5e47) junto com 24/26. AC1-3 seguem abertas: loading/error não validados visualmente (mock entrega dados prontos). Validar com app Wails rodando antes de fechar como Done.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: @lucas
created: 2026-11-01 02:56
---
AC1-3 não marcadas: exigem verificação comportamental (app Wails + screenshots/ui-reviewer) indisponível nesta sessão. Código implementado e compila.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Estados loading/error explícitos via Async<T> em useTopic + states.tsx; tabs gated; feedback de falha no export/reset. Verificado build/make check verde; AC1-3 (comportamento visual) pendentes de revisão com app rodando.
<!-- SECTION:FINAL_SUMMARY:END -->
