---
id: TASK-36
title: Validação manual no jogo (checklist do Lucas)
status: To Do
assignee:
  - '@devlucasteles'
created_date: '2026-11-01 14:53'
labels:
  - testing
dependencies: []
priority: high
ordinal: 36000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Checklist de aceitação das features já implementadas (cobertas por teste/mock, mas sem validação ao vivo). COMO RODAR: `make grant-capture` (1x por boot) e `make run` — OU `open build/bin/albion-companion.app` (destacado do terminal, fecha com Cmd+Q; NÃO use make run se travar o terminal). Sem VPN/ExitLag/GeForce Now (túnel esconde tráfego). Clica Start capture e joga. Pra cada item: faça a ação, veja o resultado no app, e me diga o que bateu/divergiu (print + número do jogo quando houver). Marque o AC quando validar.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 LOOT corpo/sacolinha: mate algo ou /die e pegue itens do corpo/sacolinha -> aparecem em Loot (feed+tabela) com lootedFrom = nome do corpo (TASK-35, era o bug)
- [ ] #2 LOOT baú de masmorra: abra um baú de dungeon e pegue -> itens contam em Loot (TASK-21)
- [ ] #3 NOMES PT-BR: itens aparecem com nome traduzido (ex. "Bolsa do Adepto"), id cru só no tooltip (TASK-18)
- [ ] #4 PRATA: soma do silver no app bate com o total que você pegou no jogo (floor por pickup)
- [ ] #5 COLETA: NÃO deve aparecer ainda (TASK-29 gathering não implementada) — confirme que está ausente, não quebrado
- [ ] #6 FAMA: total, fama/hora, amarela (Aprendizado Rápido num nó 100/100) e premium (+50%) batem com o clog; chip de fama no topo atualiza
- [ ] #7 DPS per-skill: expanda um jogador na aba DPS -> linhas por habilidade com dano e % do próprio dano (nomes = Spell #N por ora)
- [ ] #8 KILLFEED: uma morte próxima aparece no Timeline (filtro Kill) com vítima [guild] killed by killer [guild]; knockdown marcado
- [ ] #9 RATES: aba DPS mostra silver/h, respec/h, might/h além de fame/h; toggle Session/Per-hour muda os valores
- [ ] #10 DUNGEON TRACKER: faça uma dungeon (ex. RANDOMDUNGEON) -> cria um card com fame/silver/loot/baús/duração + snapshot DPS; rename e tier funcionam
- [ ] #11 TIMELINE: feed cronológico único com loot/silver/fame/kill/zone/party, filtros por tipo funcionam
- [ ] #12 VALOR DE LOOT: em Settings escolha a região (ou auto) + Refresh prices; na aba Loot aparece Est. value por player e total; offline mostra "—" e "Offline · no prices"
- [ ] #13 EXPORT CSV: botão Export CSV na aba Loot -> o arquivo abre em https://matheus.sampaio.us/ao-loot-logger-viewer (TASK-8)
- [ ] #14 PACOTES: contador de pkts no topo sobe durante a captura; abrir o popover mostra packets/messages/unknown/decodeErr/encrypted
- [ ] #15 PARTY: entre/saia de grupo -> roster atualiza; toggle Party only filtra loot/DPS pros membros
- [ ] #16 ESTABILIDADE: app não trava durante captura normal; se travar, anote o que estava fazendo (aba, tráfego, VPN)
<!-- AC:END -->
