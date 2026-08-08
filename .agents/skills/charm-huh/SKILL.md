---
name: charm-huh
description: Contrato de uso da biblioteca huh (Charm) para construir formulários e prompts interativos de terminal em Go - forms, groups, fields, validação, temas, spinner e integração com Bubble Tea. Use quando uma task de TUI Go do vault precisar coletar input do usuário (seleção de task, confirmação de ação, edição de campo) ou embutir um formulário num programa bubbletea.
---

# charm-huh

**Princípio: declare o formulário, não o loop.** Você descreve grupos e campos com bindings para variáveis; o huh cuida de navegação, validação, ajuda e renderização em cima do Bubble Tea.

## O que é / quando usar

huh é a biblioteca de formulários de terminal do ecossistema Charm (mesma família de bubbletea, lipgloss, bubbles e glamour). Use nas TUIs Go do vault (epochs 1–3) para toda interação de input: escolher uma task do kanban, confirmar uma transição de estágio, preencher campos de um card. Para exibição pura (listas, markdown renderizado), prefira bubbles + glamour; huh é para **coletar** dados.

## Instalação e imports

```bash
go get charm.land/huh/v2
```

```go
import (
    "charm.land/huh/v2"
    "charm.land/huh/v2/spinner" // spinner standalone (feedback pós-submit)
)
```

A v2 é a linha atual (módulo `charm.land/huh/v2`). A v1 legada vive em `github.com/charmbracelet/huh` — não misture as duas no mesmo módulo.

## Conceitos centrais

- **Form** — o formulário inteiro: coleção de groups exibidos um por vez ("páginas"). `huh.NewForm(groups...)`, roda com `form.Run()`.
- **Group** — uma página do form: `huh.NewGroup(fields...)`. O form só avança de grupo quando todos os campos validam.
- **Field** — um controle: `Input` (linha única), `Text` (multi-linha, abre `$EDITOR`), `Select[T]` / `MultiSelect[T]` (genéricos), `Confirm` (sim/não), `Note` (só exibe texto/markdown), `FilePicker`. Todo field é um `tea.Model`.
- **Binding** — `.Value(&variavel)` liga o field a uma variável sua; ao final do form ela está preenchida. Alternativa: `.Key("nome")` + `form.GetString("nome")` depois.
- **Option** — `huh.NewOption("rótulo visível", valor)` ou `huh.NewOptions(valores...)`; o valor pode ser qualquer tipo comparável.
- **Validação** — `.Validate(func(v T) error)`; prontos: `huh.ValidateNotEmpty()`, `huh.ValidateMinLength(n)` etc.
- **Dinâmico** — `TitleFunc`/`OptionsFunc`/`DescriptionFunc(f, &binding)` recomputam (com cache) quando o binding muda.
- **Tema** — `form.WithTheme(...)`; prontos: Charm, Catppuccin, Dracula, Base16 (`huh.ThemeCharm`, …, funções `func(isDark bool) *Styles`).
- **Estado** — `form.State`: `huh.StateNormal`, `huh.StateCompleted`, `huh.StateAborted`.

## Padrões de código

Form completo standalone (caso mais comum: coletar e seguir):

```go
var (
    task    string
    confirm bool
)

form := huh.NewForm(
    huh.NewGroup(
        huh.NewSelect[string]().
            Title("Qual task avançar?").
            Options(
                huh.NewOption("1.1.1-setup-core", "1.1.1-setup-core"),
                huh.NewOption("1.1.2-models", "1.1.2-models"),
            ).
            Value(&task),
    ),
    huh.NewGroup(
        huh.NewConfirm().
            Title("Confirmar avanço para 002_planning?").
            Affirmative("Sim").
            Negative("Não").
            Value(&confirm),
    ),
)

err := form.Run()
if err != nil {
    log.Fatal(err) // ver armadilha sobre ErrUserAborted
}
```

Prompt único (atalho sem Form explícito — `Run()` bloqueia):

```go
var name string
err := huh.NewInput().
    Title("Slug da nova task?").
    Validate(huh.ValidateNotEmpty()).
    Value(&name).
    Run()
```

Form dinâmico (opções recomputadas quando outro campo muda — note o binding `&country`):

```go
huh.NewSelect[string]().
    Title("Estado").
    OptionsFunc(func() []huh.Option[string] {
        return huh.NewOptions(statesFor(country)...)
    }, &country). // sem este binding, recomputa a cada tecla
    Value(&state)
```

huh dentro de um programa Bubble Tea (`*huh.Form` é um `tea.Model` — delegue Init/Update/View):

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if m.form.State == huh.StateCompleted {
        return m, tea.Quit // valor já está no binding ou em m.form.GetString("key")
    }
    form, cmd := m.form.Update(msg)
    if f, ok := form.(*huh.Form); ok {
        m.form = f
    }
    return m, cmd
}
```

Spinner pós-submit (feedback de ação demorada, ex.: chamada de script `pop_*`):

```go
err := spinner.New().
    Title("Movendo task no kanban...").
    Action(func() { runPopMove(task) }).
    Run()
```

## Composição com as outras libs Charm

huh é construído sobre Bubble Tea — standalone via `form.Run()` ou embutido como `tea.Model` num app maior (como farão as TUIs do vault). Os temas do huh são estilos lipgloss (`huh.ThemeCharm(isDark)` etc.). `form.WithProgramOptions(...)` aceita `tea.ProgramOption` — mas na bubbletea v2 **não existe** option de alt screen: para form em tela cheia, embuta o `*huh.Form` num programa hospedeiro cuja `View()` declare `v.AltScreen = true` (padrão acima). Markdown exibido em `Note`/descriptions pode ser pré-renderizado com glamour antes de virar string.

## Armadilhas comuns

- **`form.Run()` retorna `huh.ErrUserAborted`** quando o usuário sai com esc/ctrl+c antes de submeter — trate como cancelamento normal, não como `log.Fatal` genérico.
- **Esquecer o `&` em `.Value(&var)`** não compila (bom), mas esquecer `.Value`/`.Key` deixa o valor preso no form — todo field de input precisa de binding ou key.
- **`Select`/`MultiSelect` são genéricos:** declare o tipo (`huh.NewSelect[string]()`) e lembre que `NewOption(rotulo, valor)` separa o texto visível do valor gravado.
- **`OptionsFunc` sem o binding correto** (segundo argumento) perde o cache e reexecuta a função a cada input — em função que bate em disco/rede isso trava o form.
- **Validação é por grupo:** um field inválido bloqueia o avanço da página inteira; mensagens de erro só aparecem com `WithShowErrors(true)` (default) — não desligue sem motivo.
- **Acessibilidade:** `form.WithAccessible(os.Getenv("ACCESSIBLE") != "")` troca o TUI por prompts simples para screen readers; `WithTimeout` não funciona em modo acessível (`ErrTimeoutUnsupported`).

## Links oficiais

- Repo: https://github.com/charmbracelet/huh — exemplos completos em `examples/` (dynamic, bubbletea, spinner).
- Docs Charm: https://charm.land — guia e referência de temas.
- API: https://pkg.go.dev/charm.land/huh/v2 — referência completa de fields, keymaps e layouts.
