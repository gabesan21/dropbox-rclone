---
name: charm-bubbles
description: Contrato de uso da biblioteca Bubbles (charm.land/bubbles/v2) — componentes TUI prontos (list, viewport, textinput, spinner, progress, table, paginator, help, key) para Bubble Tea. Use quando for codar ou revisar as TUIs Go do vault (leitura do kanban/INBOX, TUI de ações) e precisar montar telas interativas com os componentes oficiais da Charm.
---

# charm-bubbles

## O que é / quando usar

Bubbles é a coleção oficial de componentes de UI para Bubble Tea — cada componente (um "bubble") é um `tea.Model` pronto: lista navegável, viewport com scroll, inputs, spinner, barra de progresso etc. Use em toda task de código das TUIs do vault (epochs 2 e 3 do [[ROADMAP|ROADMAP]]): leitura do kanban/INBOX pede `list` + `viewport` + `glamour`; ações (mover card, criar task) pedem `textinput`/`textarea`, `spinner` e `help`. Esta skill cobre a **v2** (`charm.land/bubbles/v2`), que exige Bubble Tea v2 e Lip Gloss v2 — docs e exemplos v1 (`github.com/charmbracelet/bubbles`) estão desatualizados; não copie API deles.

## Instalação e imports

```sh
go get charm.land/bubbles/v2@latest
go get charm.land/bubbletea/v2@latest
go get charm.land/lipgloss/v2@latest
```

Cada componente é um subpacote próprio — importe só o que usa:

```go
import (
    tea "charm.land/bubbletea/v2"
    "charm.land/bubbles/v2/help"
    "charm.land/bubbles/v2/key"
    "charm.land/bubbles/v2/list"
    "charm.land/bubbles/v2/paginator"
    "charm.land/bubbles/v2/progress"
    "charm.land/bubbles/v2/spinner"
    "charm.land/bubbles/v2/table"
    "charm.land/bubbles/v2/textarea"
    "charm.land/bubbles/v2/textinput"
    "charm.land/bubbles/v2/viewport"
    "charm.land/lipgloss/v2"
)
```

## Conceitos centrais

- **Componente = `tea.Model`:** todo bubble tem `Init()`, `Update(msg) (Model, tea.Cmd)` e `View() string`. O model hospedeiro embute os bubbles como campos, encaminha as mensagens no `Update` e concatena os `View()`s no layout.
- **`key.Binding` é o contrato de teclas:** declare um `KeyMap` próprio com `key.NewBinding(key.WithKeys(...), key.WithHelp(...))`, teste com `key.Matches(msg, binding)` e sirva o mesmo mapa ao `help` — tecla e ajuda nunca divergem.
- **Mensagens próprias:** cada bubble tem seus tipos de msg (`spinner.TickMsg`, `progress.FrameMsg`). Animação só anda se você retornar o `tea.Cmd` do componente.
- **Construção por opções:** `viewport.New(viewport.WithWidth(80))`, `spinner.New(spinner.WithSpinner(spinner.Dot))`. Tamanho se ajusta com `SetWidth`/`SetHeight`/`SetSize` (campos `Width`/`Height` exportados não existem mais na v2).
- **Estilos claro/escuro explícitos:** a v2 não detecta o fundo do terminal — `DefaultStyles(isDark bool)` em `list`, `help`, `textinput`, `textarea`. Obtenha `isDark` com `tea.RequestBackgroundColor` (resposta em `tea.BackgroundColorMsg`) ou `lipgloss.HasDarkBackground(os.Stdin, os.Stdout)` (função da raiz do lipgloss v2; no subpacote `compat` é variável, não função).

## Padrões de código

### Lista navegável (leitura do kanban)

```go
type card struct{ titulo, estagio string }

func (c card) Title() string       { return c.titulo }
func (c card) Description() string { return c.estagio }
func (c card) FilterValue() string { return c.titulo } // habilita fuzzy filter

items := []list.Item{card{titulo: "1.1.1-setup", estagio: "004_processing"}}
l := list.New(items, list.NewDefaultDelegate(), 0, 0)
l.Title = "Kanban"
// Update do hospedeiro:
//   case tea.WindowSizeMsg: m.list.SetSize(msg.Width, msg.Height)
//   m.list, cmd = m.list.Update(msg)   // encaminha TODAS as msgs
```

### Viewport com conteúdo renderizado (markdown via glamour)

```go
vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(24))
vp.SetContent(markdownRenderizado) // string ANSI vinda do glamour
// vp.SoftWrap = true              // quebra suave opcional
// Update: m.viewport, cmd = m.viewport.Update(msg)  // scroll por tecla/mouse
```

### Spinner + Cmd assíncrono (ação em andamento)

```go
s := spinner.New(spinner.WithSpinner(spinner.Dot))
// Init do hospedeiro: return tea.Batch(m.spinner.Tick, minhaCmdAssincrona)
// Update:
case spinner.TickMsg:
    var cmd tea.Cmd
    m.spinner, cmd = m.spinner.Update(msg)
    return m, cmd
```

### Teclas + help (qualquer tela)

```go
type keyMap struct{ Sair key.Binding }

var keys = keyMap{
    Sair: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "sair")),
}

func (k keyMap) ShortHelp() []key.Binding   { return []key.Binding{k.Sair} }
func (k keyMap) FullHelp() [][]key.Binding  { return [][]key.Binding{{k.Sair}} }

// Update: case key.Matches(msg, keys.Sair): return m, tea.Quit
// View: m.help.View(keys)  // help.New(), com m.help.SetWidth(w)
```

### Referência rápida dos demais

- **textinput/textarea:** `textinput.New()`, depois `ti.Placeholder`, `ti.Focus()`, `ti.CharLimit`, `ti.SetWidth(40)`; valor em `ti.Value()`. `textarea` para multi-linha; estilos via `DefaultStyles(isDark)` + `SetStyles`.
- **progress:** `progress.New(progress.WithDefaultBlend())`; estático com `p.ViewAs(0.65)`, animado encaminhando `progress.FrameMsg` no `Update`.
- **table:** `table.New(table.WithColumns([]table.Column{{Title: "Task", Width: 30}}), table.WithRows(rows), table.WithFocused(true))`; seleção em `t.SelectedRow()`.
- **paginator:** `paginator.New()`, `p.Type = paginator.Dots`, `p.PerPage = 10`, `p.SetTotalPages(n)`; customize `p.KeyMap` (os toggles `UseJKKeys` etc. sumiram na v2).

## Composição com as outras libs Charm

Bubbles não roda sozinho: exige Bubble Tea v2 (o loop `Model/Update/View`, `tea.Cmd`, `tea.KeyPressMsg`) e se estiliza com Lip Gloss v2. O glamour gera a string ANSI do markdown do vault e o `viewport` a exibe — a largura do renderer do glamour deve bater com a do viewport. Para formulários declarativos completos (wizard de `new-task`), avalie `huh` em vez de compor `textinput` na mão.

## Armadilhas comuns

- **Esquecer de encaminhar msgs:** se o `Update` do hospedeiro não repassar a msg ao bubble (ou descartar o `tea.Cmd` retornado), spinner congela, lista não filtra e progress não anima.
- **API v1 fantasma:** `tea.KeyMsg` (→ `tea.KeyPressMsg`), `NewModel()` (→ `New()`), `DefaultKeyMap` variável (→ função), campos `Width`/`Height` diretos (→ setters) são v1 — a v2 quebra em todos; guia oficial em `UPGRADE_GUIDE_V2.md` no repo.
- **Bloquear o loop:** nunca faça I/O ou `time.Sleep` no `Update` — trabalho lento vai numa `tea.Cmd` (assíncrona), com `spinner`/`progress` dando feedback.
- **ANSI quebrado no viewport:** conteúdo do glamour com largura maior que o viewport corta sequências de escape ao quebrar linha; iguale a largura do renderer à do viewport (ou `SoftWrap` + reflow).
- **Alt screen é campo da `tea.View`, não ProgramOption:** na v2 **não existe** `tea.WithAltScreen()` (API v1 fantasma). Tela cheia se declara na `View()` do model hospedeiro: `v := tea.NewView(conteudo); v.AltScreen = true; return v`. TUIs full-screen (lista + viewport) precisam disso para não poluir o scrollback; inputs inline simples não.

## Links oficiais

- Repo: <https://github.com/charmbracelet/bubbles> (exemplos por componente em cada subpacote)
- Docs/API: <https://pkg.go.dev/charm.land/bubbles/v2>
- Guia de migração v1→v2: <https://github.com/charmbracelet/bubbles/blob/main/UPGRADE_GUIDE_V2.md>
- Charm: <https://charm.land>
