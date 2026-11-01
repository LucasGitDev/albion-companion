# PRODUCT.md — Albion Companion

> Contexto de produto que guia as decisões de UI/UX. Loot logger + DPS meter passivo
> para Albion Online, multiplataforma (macOS, Windows, Linux).

## Público (ICP)
- **Quem usa:** jogadores de Albion Online que fazem conteúdo em grupo — PvP de ZvZ,
  ganking, dungeons, roaming e farm de facção. Técnicos o suficiente para instalar um
  app desktop e conceder permissão de captura de rede; querem números, não configuração.
- **Contexto de uso:** app desktop aberto num **segundo monitor**, lido de relance
  durante o jogo (muitas vezes à noite). Precisa ser legível a distância e não roubar foco.
- **Dor principal:** saber, ao vivo e depois, **quem pegou qual loot/quanto silver** e
  **quem está causando/levando dano** — para dividir loot com justiça e avaliar performance.
- **Alternativas atuais:** StatisticsAnalysis (Triky313) e ao-loot-logger — ambos
  **só Windows**; e planilhas manuais. O diferencial é ser multiplataforma (macOS first-class)
  e 100% passivo.

## Problema → Valor
- **Job-to-be-done:** "Quando jogo em grupo, quero ver ao vivo o loot e o dano de cada um,
  para dividir o silver de forma justa e medir quem contribuiu."
- **Proposta de valor:** o companheiro passivo e multiplataforma que transforma os pacotes
  de rede do Albion em um ledger de loot e um medidor de DPS confiáveis — sem tocar no jogo.

## Tom de voz
- **Personalidade:** direto, técnico e honesto. Transparente sobre limites ("passive capture",
  o que é e o que não é observável). Nada de hype.
- **Faz:** frases curtas e concretas; deixa claro que é passivo e não envia nada ao jogo;
  usa o vocabulário do jogo (loot, silver, fame, party, zone).
- **Não faz:** jargão vazio, promessas de vantagem competitiva/automação, emoji como ícone,
  linguagem que sugira cheating.

## Ação principal por tela
| Tela | Objetivo único | Ação / CTA primário | Sinal de sucesso |
|------|----------------|---------------------|------------------|
| **Loot** | Ver quem pegou o quê e quanto silver | Iniciar captura · **Export CSV** · filtrar por jogador | Ledger preenchido; CSV exportado no formato ao-loot-logger |
| **DPS** | Comparar dano/DPS por jogador no combate | Alternar **Party only** / **Players-only targets** | Ranking de dano com share % ao vivo |
| **Party** | Confirmar quem está no grupo (base do filtro) | — (reflete o estado da party) | Membros corretos listados; filtro de party coerente |
| **Timeline** | Reconstruir a sessão (loot, silver, fame) em ordem | Rolar o feed cronológico | Eventos recentes legíveis em tempo real |
| **Settings** | Ajustar captura e overrides sem quebrar nada | Selecionar fontes / arquivo de codes override | Captura ativa; codes atualizados/não-outdated |

## Restrições
- **Stack:** Go (engine) + Wails v2 + React/TS + Tailwind v4 + shadcn/ui.
- **Passivo apenas:** só lê pacotes; nunca envia, injeta, lê memória ou automatiza input.
- **Acessibilidade:** meta WCAG AA (contraste, foco visível, navegação por teclado).
- **Plataformas:** desktop macOS (first-class), Windows, Linux. Licença GPL-3.0.
- **Limite conhecido:** túneis VPN/ExitLag escondem o tráfego do pcap (igual aos tools upstream).
