---
name: charm-bubbletea
description: Contrato de uso do ecossistema Charm para TUIs em Go (bubbletea, lipgloss, bubbles, huh, glamour) - modelo mental, imports v2, padrões de código e armadilhas. Use quando for criar ou alterar qualquer código de TUI em Go do vault (core, TUI de leitura do kanban/INBOX, TUI de ações), incluindo renderizar markdown do vault no terminal com glamour.
---

# charm-bubbletea

## O que é / quando usar

Stack oficial de TUIs em Go da Charm. `bubbletea` é o runtime (Elm Architecture), `lipgloss` estiliza e monta layout, `bubbles` dá componentes prontos, `huh` monta formulários e `glamour` renderiza markdown no terminal. É a stack das TUIs Go do vault (core + leitura do kanban/INBOX + ações); glamour é a peça-chave para exibir os markdowns do vault. Leia esta skill antes de codar qualquer TUI — depois siga a doc oficial da lib que for usar.

## Instalação e imports

As libs migraram para `charm.land` com major `v2` (repos no GitHub seguem `charmbracelet/<lib>`). Imports atuais:

```go
go get charm.land/bubbletea/v2   // runtime: tea
go get charm.land/lipgloss/v2    // estilo e layout
go get charm.land/bubbles/v2     // componentes (spinner, list, viewport, textinput...)
go get charm.land/huh/v2         // formulários e prompts
go get charm.land/glamour/v2     // markdown no terminal
```

## Conceitos centrais

- **bubbletea** — Elm Architecture: `Model` (estado, geralmente struct por valor) com 3 métodos: `Init() tea.Cmd` (I/O inicial), `Update(tea.Msg) (tea.Model, tea.Cmd)` (reage a eventos) e `View() tea.View` (declara a UI; o runtime redesenha). `Msg` é qualquer tipo (tecla = `tea.KeyPressMsg`, resize = `tea.WindowSizeMsg`); `Cmd` é `func() tea.Msg` — toda I/O assíncrona vive em Cmds, nunca no corpo do Update. `tea.Quit` encerra.
- **lipgloss** — `lipgloss.NewStyle()` imutável e encadeável (Bold, Foreground, Padding, Border...); `Style.Render(s)` produz string ANSI. Layout com `JoinHorizontal`/`JoinVertical`/`Place`; medida com `lipgloss.Width/Height/Size`. Cores: `lipgloss.Color("63")` (256) ou `"#7D56F4"` (truecolor); downsampling ao perfil do terminal é automático.
- **bubbles** — componentes que são Models: têm `Update`/`View` próprios; você os embute no seu Model, delega o `msg` no Update e concatena os Cmds. `spinner`, `list`, `viewport`, `textinput`, `table`, `progress`, `paginator`, `help`, `key` (bindings com `key.Matches`).
- **huh** — `huh.NewForm(huh.NewGroup(campos...))` (Group = página); campos: `NewInput`, `NewText`, `NewSelect[T]`, `NewMultiSelect[T]`, `NewConfirm`, com `Value(&var)`, `Validate(fn)` e `Options(huh.NewOption(rotulo, valor))`. Standalone via `form.Run()`; embutido, `huh.Form` é um `tea.Model`.
- **glamour** — renderiza markdown para ANSI: atalho `glamour.Render(md, "dark")` ou `glamour.NewTermRenderer(glamour.WithWordWrap(w))` + `r.Render(md)`. Styles prontos ("dark", "light", "notty"...) ou stylesheet próprio; `GLAMOUR_STYLE` como env.

## Padrões de código

Programa mínimo bubbletea v2 (tecla `q` sai; alt screen é campo da View):

```go
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

type model struct{ count int }

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView(fmt.Sprintf("Olá, vault!\n\nq: sair\n"))
	v.AltScreen = true // tela cheia; em v2 não existe tea.WithAltScreen
	return v
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("erro:", err)
	}
}
```

Render de markdown do vault com glamour (largura real = viewport menos borda/padding do lipgloss):

```go
import "charm.land/glamour/v2"

r, _ := glamour.NewTermRenderer(
	glamour.WithStandardStyle("dark"),
	glamour.WithWordWrap(width), // default é 80: ajuste à largura real
)
out, err := r.Render(markdownString)
```

Formulário huh standalone (útil para confirmações e inputs na TUI de ações):

```go
import "charm.land/huh/v2"

var estagio string
var confirma bool

form := huh.NewForm(huh.NewGroup(
	huh.NewSelect[string]().
		Title("Mover task para").
		Options(huh.NewOptions("002_planning", "004_processing")...).
		Value(&estagio),
	huh.NewConfirm().Title("Confirmar?").Value(&confirma),
))
err := form.Run()
```

Estilo e layout com lipgloss (bloco com borda para cards do kanban):

```go
import "charm.land/lipgloss/v2"

card := lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(0, 1).Width(40).
	Render(titulo + "\n" + descricao)
```

Spinner bubbles embutido (embutir = delegar Update e concatenar Cmd):

```go
import "charm.land/bubbles/v2/spinner"

type model struct{ sp spinner.Model }

func (m model) Init() tea.Cmd { return m.sp.Tick } // inicia a animação

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.sp, cmd = m.sp.Update(msg)
	return m, cmd
}
```

## Composição entre as libs

bubbles e huh são Models bubbletea — rodam dentro do seu `Update`/`View`. lipgloss estiliza tudo (strings que entram na `tea.View` e nos temas do huh). O pipeline de leitura do vault é: markdown do arquivo → glamour (string ANSI) → lipgloss (enquadrar/posicionar) → bubbletea exibe, geralmente dentro de um `viewport` do bubbles para scroll.

## Armadilhas comuns

- **v1 vs v2:** em v2 `View()` retorna `tea.View` (não string), alt screen é `v.AltScreen = true` (não existe `tea.WithAltScreen`), tecla é `tea.KeyPressMsg` e os imports são `charm.land/<lib>/v2`. Exemplos antigos da web quebram — confira a major antes de copiar.
- **Nunca faça I/O ou processamento pesado no Update:** o loop congela e a UI trava. Leitura de arquivo, parse e chamadas externas viram `tea.Cmd` que devolve uma `Msg` com o resultado.
- **Largura do glamour:** o wrap default é 80 colunas. Passe `WithWordWrap` com a largura útil real (viewport − borda − padding do lipgloss), senão o conteúdo estoura ou quebra duas vezes.
- **glamour não faz downsampling de cor** (renderer é puro): fora do bubbletea, imprima a saída via `lipgloss.Print` para respeitar o perfil de cor do terminal. Com bubbletea v2, o downsampling já é embutido.
- **Componente esquecido congela:** ao embutir bubble/huh, todo `msg` precisa chegar ao `Update` do filho e todo `Cmd` dele precisa ser retornado (some com `tea.Batch` quando houver mais de um).

## Links oficiais

- bubbletea: [repo](https://github.com/charmbracelet/bubbletea) · [pkg.go.dev](https://pkg.go.dev/charm.land/bubbletea/v2)
- lipgloss: [repo](https://github.com/charmbracelet/lipgloss) · [pkg.go.dev](https://pkg.go.dev/charm.land/lipgloss/v2)
- bubbles: [repo](https://github.com/charmbracelet/bubbles) · [pkg.go.dev](https://pkg.go.dev/charm.land/bubbles/v2)
- huh: [repo](https://github.com/charmbracelet/huh) · [pkg.go.dev](https://pkg.go.dev/charm.land/huh/v2)
- glamour: [repo](https://github.com/charmbracelet/glamour) · [pkg.go.dev](https://pkg.go.dev/charm.land/glamour/v2)
- Portal Charm: [charm.land](https://charm.land)
