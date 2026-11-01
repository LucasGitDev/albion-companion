# DESIGN.md — Albion Companion

> Fonte única de verdade visual do app (Go + Wails + React/TS + Tailwind v4 + shadcn).
> Nenhum valor de cor/tipografia/espaçamento/motion deve existir fora dos tokens abaixo.
> Os tokens vivem em `frontend/src/style.css` (`@layer base :root` + `@theme inline`).
> Tema: **"dungeon ledger"** — superfícies escuras e quentes, lidas num segundo monitor
> à noite. Ouro = sinal de loot/silver; ember = sinal de dano. A marca mora nos dados,
> não no cromo.

## Tokens de cor (HSL — consumidos via Tailwind `@theme inline` como `--color-*`)

| Token | HSL | Papel |
|-------|-----|-------|
| `--background` | `222 20% 6.5%` | Fundo da janela (slate quase preto, azulado) |
| `--surface` | `222 17% 9%` | Painéis, sidebar, barras de sessão (`bg-surface`) |
| `--elevated` | `222 15% 12%` | Linhas/chips elevados, hover de tabela |
| `--card` / `--popover` | `222 17% 9%` / `222 16% 11%` | Cards e popovers |
| `--foreground` | `40 12% 93%` | Texto principal (branco levemente quente) |
| `--muted-foreground` | `220 10% 68%` | Texto secundário, labels, headers de tabela (≥4.5:1 sobre o fundo, WCAG AA) |
| `--primary` (ouro) | `41 92% 55%` | **Loot / silver.** Ação primária, acento único por tela |
| `--primary-foreground` | `36 45% 8%` | Texto sobre ouro |
| `--ember` | `18 88% 57%` | **Dano** (DPS, erros de decode) |
| `--heal` | `150 56% 47%` | Cura / estado "capturando" / sucesso / Timeline: entrou na party |
| `--taken` | `2 74% 60%` | Dano recebido / Timeline: kill |
| `--fame` | `265 62% 70%` | **Categoria "fame" no feed da Timeline/Dungeon (violeta).** NÃO é a cor da métrica fame — a métrica fame (TopBar, aba DPS) usa ouro (`--primary`), porque fame é progressão/loot. `--fame` só distingue eventos de fama no feed multicategoria, onde ouro já marca silver. |
| `--zone` | `190 64% 55%` | **Zona / navegação.** Acento da Timeline (ciano) |
| `--warning` | `34 90% 56%` | Banners de cautela (ex.: codes.json desatualizado) — âmbar, não é erro |
| `--destructive` | `2 74% 55%` | Destrutivo / erro de captura |
| `--secondary` / `--muted` / `--accent` | `222 13% 16/15/17%` | Neutros tingidos de superfície |
| `--border` | `222 12% 17%` | Bordas (aplicada a `*` por padrão) |
| `--input` | `222 12% 20%` | Bordas de input |
| `--ring` | `41 92% 55%` | Anel de foco (ouro) |

Regra: **um acento por tela.** Ouro domina Loot/fame; ember domina DPS. Todo o resto
é neutro tingido (nunca cinza puro — todos os neutros carregam a matiz `222`).
Exceção: a **Timeline** é um feed multi-categoria (como uma legenda) e mapeia cada tipo
de evento a um token semântico — loot/silver=`--primary`, fame=`--fame`, zone=`--zone`,
party entrou=`--heal`, party saiu=`--muted-foreground`, kill=`--taken` — nunca a cores
Tailwind cruas. Chips de filtro e linhas do feed usam o mesmo token por tipo.
Contraste: texto normal ≥ 4.5:1, UI/texto grande ≥ 3:1.

Fundo do `body` tem um glow radial de ouro a 5% no canto superior direito — sutil,
não decorativo genérico.

## Tipografia

Fontes empacotadas offline via `@fontsource-variable/*` (Vite bundle, sem CDN —
importadas em `frontend/src/main.tsx`). Todas variáveis.

- **Display — `Space Grotesk Variable`** (`font-display`): títulos de seção e **números
  grandes** (ledger de loot, silver, fame, dano). Grotesk com personalidade e algarismos
  distintos que ancoram leituras numéricas num overlay escuro.
- **Texto/UI — `Geist Variable`** (`font-sans`, default do `body`): corpo, tabelas densas,
  labels, controles. Grotesk neutra e altíssima legibilidade — **não é Inter**.
- **Mono — `JetBrains Mono Variable`** (`font-mono`): valores monoespaçados pontuais.
- **Proibido por padrão:** Inter, Roboto, system-ui, Arial (só com justificativa).
- Números tabulares: classe `.tnum` (`font-variant-numeric: tabular-nums`) em toda leitura
  numérica que alinha em coluna.
- Fallback: fonte escolhida → `ui-sans-serif, system-ui, -apple-system, sans-serif`.
- Removido `font-feature-settings: 'cv11','ss01'` (eram específicos da Inter).
- `letter-spacing: -0.011em` no body; `tracking-tight`/`tracking-wider` pontual em títulos.

Onde o display é aplicado: `SectionTitle` e valores do `Ledger` (Loot), `FameStat` e fame
total (DPS), `h1` da StartScreen, marca da Sidebar, fame na TopBar, `h2` da Party.

## Raio / Sombra / Espaço

```css
--radius:    0.75rem;               /* --radius-lg */
--radius-md: calc(var(--radius) - 4px);  /* 0.5rem  */
--radius-sm: calc(var(--radius) - 6px);  /* 0.375rem */
/* cards/barras: rounded-xl; chips/botões: rounded-md/lg; pills: rounded-full */
```

- **Sombras:** o app é "flat sobre escuro" — hierarquia por `--surface`/`--elevated`, não
  por sombra. Única exceção: popovers usam `shadow-xl shadow-black/40`.
- **Espaço:** escala do Tailwind (base 4px). Gaps típicos: `gap-5` entre seções,
  `gap-3` dentro de seção, `px-5 py-4` em barras de sessão, `px-3 py-2.5` em linhas.

## Motion

- Transições **curtas (< 300ms)**, só `transition-colors`/`transition-transform`/
  `transition-all` com a duração padrão do Tailwind (150ms).
- **Nunca ease-in puro.** Preferir o default (ease) ou saída suave.
- Único movimento contínuo permitido: o `animate-ping` do indicador "Capturing".
- Respeitar `prefers-reduced-motion` ao adicionar qualquer novo movimento.

## Do's
- Hierarquia por tamanho/peso/cor e por `surface`/`elevated`, não por caixas empilhadas.
- Um acento de cor por tela (ouro ou ember); resto em neutros tingidos.
- `.tnum` + `font-display` em todo número grande/alinhado.
- Estados de foco visíveis (anel ouro de 2px) — já global via `:focus-visible`.

## Don'ts
- Inter/Roboto/system-ui/Arial como default.
- Gradiente genérico roxo→azul; grid decorativo de fundo.
- Cinza puro (sempre tingir com a matiz `222`).
- Cards idênticos em grade sem foco; emoji como ícone (usar `lucide-react`).
- Sombras pesadas para criar profundidade; motion > 300ms ou ease-in.
