---
id: TASK-25
title: 'UI: estados de loading e error nas abas'
status: In Progress
assignee:
  - '@lucas'
created_date: '2026-11-01 02:40'
updated_date: '2026-11-01 02:47'
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
- [ ] #4 make check verde
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Mapear hooks (useLoot/useCombat/useStatus/useTopic) e chamadas api.*. 2. Distinguir loading de empty: expor estado (undefined=loading vs vazio=empty) e render skeleton/indicador em Loot/DPS/Party/Timeline. 3. Adicionar .catch nas chamadas api.* + estado de erro com retry. 4. Export CSV e Reset: feedback de falha. 5. npm run build.
<!-- SECTION:PLAN:END -->
