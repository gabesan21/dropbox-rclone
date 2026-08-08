---
name: charm-lipgloss
description: Estilos e layout de terminal com a lib Lip Gloss do ecossistema Charm - estilos declarativos (cores, bordas, padding, alinhamento), composição de blocos (join/place), subpacotes table/list/tree e medidas de células. Use quando for escrever ou estilizar a View de uma TUI Go no vault (core Go, TUI de leitura, TUI de ações), montar layouts de terminal com blocos lado a lado, renderizar tabelas/listas/árvores estáticas ou padronizar a identidade visual dos TUIs do PoP.
---

# charm-lipgloss

**Princípio: Lip Gloss é a camada de apresentação, não a aplicação.** Ele renderiza strings estilizadas; quem roda o loop interativo é o Bubble Tea (skill `charm-bubbletea`). Todo visual das TUIs Go do vault passa por aqui — trate estilos como tokens de design: defina-os uma vez por pacote e reutilize.

## O que é / quando usar

Lip Gloss é a biblioteca de estilos declarativos do ecossistema Charm (modelo mental de CSS para o terminal). Use para estilizar texto e montar o layout da `View()` de um programa Bubble Tea, para saída formatada de CLI sem interatividade e para tabelas/listas/árvores estáticas. É o complemento do glamour: o glamour renderiza o markdown do vault, o lipgloss emoldura, alinha e coloriza o resto da tela.

## Instalação e imports

```bash
go get charm.land/lipgloss/v2
```

```go
import (
    "charm.land/lipgloss/v2"
    "charm.land/lipgloss/v2/table" // subpacotes: table, list, tree
)
```

Use **v2** (`charm.land/lipgloss/v2`) — é a versão atual. A v1 (`github.com/charmbracelet/lipgloss`) só aparece em código legado; se encontrar, siga o upgrade guide oficial antes de mexer.

## Conceitos centrais

- **`Style` é um tipo-valor imutável:** cada método (`Bold`, `Foreground`, `Padding`...) retorna uma cópia nova. Atribuir (`a := b`) já copia — encadeie na declaração e reutilize o estilo como token.
- **Cores:** `lipgloss.Color("205")` (ANSI256) ou `lipgloss.Color("#7D56F4")` (hex); constantes nomeadas para as 16 ANSI (`lipgloss.Red`, `lipgloss.BrightCyan`...). O downsampling para o perfil do terminal é automático ao imprimir com `lipgloss.Println`/`Sprint`/`Fprint` (drop-in do `fmt`) — com Bubble Tea v2, embutido.
- **Block model (CSS):** `Padding`/`Margin` com shorthand de 1 a 4 valores (sentido horário do topo), `Width`/`Height`, `Align(lipgloss.Center)` com posições `Top/Bottom/Center/Left/Right` (0.0 a 1.0), `Border(lipgloss.RoundedBorder(), lados...)` com `BorderForeground`.
- **Layout de blocos prontos:** `JoinHorizontal`/`JoinVertical` colam blocos multi-linha alinhados; `Place`/`PlaceHorizontal`/`PlaceVertical` posicionam um bloco num espaço em branco.
- **Medida em células:** `lipgloss.Width`, `lipgloss.Height`, `lipgloss.Size` ignoram sequências ANSI e contam corretamente emoji/CJK — nunca `len()` em string estilizada.
- **Subpacotes:** `table` (tabelas com `StyleFunc` por linha/coluna), `list` (listas aninhadas com enumeradores), `tree` (árvores de diretório).
- **Camadas:** `NewLayer`/`NewCompositor` compõem conteúdo sobreposto por coordenada (X/Y/Z) — para a maioria das telas, Join + Place bastam.

## Padrões de código

Estilo reutilizável com borda (o padrão "card" das TUIs do vault):

```go
var cardStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("63")).
    Padding(0, 1).
    Width(40)

fmt.Println(cardStyle.Render("Task 1.1.1\nEm planejamento"))
```

Layout em duas colunas com `JoinHorizontal` e `Place` (lista + detalhe):

```go
lista := lipgloss.NewStyle().Width(30).Render("• task A\n• task B")
detalhe := lipgloss.NewStyle().Width(50).Padding(0, 1).Render(card)

tela := lipgloss.JoinHorizontal(lipgloss.Top, lista, detalhe)
// Centralizar na tela inteira (Bubble Tea: use msg.Width/Height do WindowSizeMsg):
view := lipgloss.Place(largura, altura, lipgloss.Center, lipgloss.Center, tela)
```

Cor adaptada ao fundo do terminal (standalone; em Bubble Tea, capture `tea.BackgroundColorMsg`):

```go
hasDarkBG := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
lightDark := lipgloss.LightDark(hasDarkBG)
titulo := lipgloss.NewStyle().
    Bold(true).
    Foreground(lightDark(lipgloss.Color("#1a1a1a"), lipgloss.Color("#FAFAFA")))
```

Tabela estática com o subpacote `table` (kanban em texto):

```go
t := table.New().
    Border(lipgloss.NormalBorder()).
    BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
    Headers("TASK", "ESTÁGIO", "DONO").
    Rows(
        []string{"1.1.1", "004_processing", "agent"},
        []string{"M-2.1", "002_planning", "agent"},
    )

lipgloss.Println(t)
```

## Composição com as outras libs Charm

- **bubbletea:** o loop é dele; a `View()` retorna string montada com lipgloss (Join/Place + estilos). Downsampling de cor é automático na v2.
- **bubbles:** cada componente aceita estilos lipgloss (ex.: `list.Styles.Title`; no textinput v2, via `SetStyles` com estilo por estado) — nunca estilize o componente por fora com `Render` geral, use os campos de estilo dele.
- **glamour:** renderiza os markdowns do vault para string ANSI; enquadre a saída com `lipgloss.NewStyle().Width(...)` ou coloque num viewport estilizado.
- **huh:** temas (`huh.ThemeCharm()` etc.) são construídos com estilos lipgloss.

## Armadilhas comuns

- **Medir string estilizada com `len()`:** sequências ANSI inflam o tamanho e quebram o layout — use sempre `lipgloss.Width`/`Size` para calcular alinhamento e Join.
- **Misturar v1 e v2:** os imports são incompatíveis (`github.com/charmbracelet/lipgloss` vs `charm.land/lipgloss/v2`); um projeto usa uma só. Bubbles/glamour atuais já são v2.
- **`fmt.Println` em saída standalone:** sem os writers do lipgloss (`lipgloss.Println`, `Fprint`) não há downsampling — cores hex vazam como sequências cruas em terminal sem truecolor.
- **Esquecer a moldura no cálculo de largura:** borda + padding + margem somam ao `Width`; para encaixar num espaço de N células, subtraia `GetHorizontalFrameSize()` (ou defina `Width` já descontando).
- **Cor fixa sem checar o fundo:** foreground claro em terminal claro some; use `LightDark`/`HasDarkBackground` (ou `tea.BackgroundColorMsg` em Bubble Tea) em vez de cor única.
- **`Style` não muta:** `s.Bold(true)` sozinho descarta o resultado — sempre atribua (`s = s.Bold(true)`) ou encadeie.

## Links oficiais

- Repo: <https://github.com/charmbracelet/lipgloss> (README = documentação principal, com exemplos visuais)
- Docs da API v2: <https://pkg.go.dev/charm.land/lipgloss/v2>
- Upgrade guide v1 → v2: link "upgrade guide" no README do repo
- Charm: <https://charm.land>
