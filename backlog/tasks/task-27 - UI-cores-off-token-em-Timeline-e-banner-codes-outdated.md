---
id: TASK-27
title: 'UI: cores off-token em Timeline e banner codes-outdated'
status: To Do
assignee: []
created_date: '2026-11-01 03:12'
updated_date: '2026-11-01 13:55'
labels:
  - ui
  - design
dependencies: []
priority: high
ordinal: 27000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Revisão visual (ui-reviewer) encontrou cores fora dos tokens do DESIGN.md: a aba Timeline mapeia tipos de evento para uma paleta arco-íris Tailwind crua (sky-400/violet-400/emerald-400/teal-400/rose-400), violando 'nenhuma cor fora dos tokens' e 'um acento por tela'; e o banner de 'codes outdated' no App usa amber-500/400 cru em vez de token. Desktop-only — não há questão de responsivo aqui.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Timeline não usa classes de cor Tailwind cruas; cada tipo de evento mapeia para tokens do DESIGN.md (primary/ember/heal/taken/muted-foreground) ou tokens nomeados novos no style.css
- [ ] #2 Banner de codes-outdated (App.tsx) usa token (ex.: --primary ou --destructive), sem amber-* cru
- [ ] #3 DESIGN.md reflete quaisquer tokens novos adicionados
- [ ] #4 make check verde
<!-- AC:END -->
