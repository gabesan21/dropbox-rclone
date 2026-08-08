package main

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// backupItem adapta Backup para o componente list do bubbles.
type backupItem struct{ backup Backup }

// displayName identifica a entrada pelo name, com fallback para o path
// em entradas legadas ainda sem nome.
func displayName(b Backup) string {
	if b.Name != "" {
		return b.Name
	}
	return b.Path
}

// displayCicle resolve o ciclo para exibição; vazio equivale a 24h (1x/dia).
func displayCicle(b Backup) string {
	if b.RepeatCicle != "" {
		return b.RepeatCicle
	}
	return "24h"
}

func (i backupItem) Title() string { return displayName(i.backup) }

func (i backupItem) Description() string {
	b := i.backup
	return fmt.Sprintf("%s:%s · %s · %s · ciclo=%s · max=%d",
		b.RcloneAccount, b.RemotePath, b.Type, b.BackupTime, displayCicle(b), b.MaxBackups)
}

func (i backupItem) FilterValue() string { return displayName(i.backup) }

// action é a escolha feita na lista, executada fora do programa bubbletea.
type action int

const (
	actionQuit action = iota
	actionAdd
	actionEdit
	actionDelete
)

var footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginLeft(2)

type listModel struct {
	list   list.Model
	action action
}

func newListModel(backups []Backup) listModel {
	items := make([]list.Item, len(backups))
	for i, b := range backups {
		items[i] = backupItem{b}
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Backups agendados"
	l.DisableQuitKeybindings()
	l.SetShowHelp(false)
	return listModel{list: l, action: actionQuit}
}

func (m listModel) Init() tea.Cmd { return nil }

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-2) // 2 linhas do rodapé
		return m, nil
	case tea.KeyPressMsg:
		// Com o filtro ativo, toda tecla pertence ao componente.
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "ctrl+c", "q":
			m.action = actionQuit
			return m, tea.Quit
		case "a":
			m.action = actionAdd
			return m, tea.Quit
		case "e":
			if m.list.SelectedItem() != nil {
				m.action = actionEdit
				return m, tea.Quit
			}
		case "d":
			if m.list.SelectedItem() != nil {
				m.action = actionDelete
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) View() tea.View {
	footer := footerStyle.Render("a: adicionar · e: editar · d: remover · /: filtrar · q: sair")
	v := tea.NewView(m.list.View() + "\n" + footer)
	v.AltScreen = true
	return v
}

// selectedBackup retorna o Backup sob o cursor, resolvendo o índice real
// na slice (o cursor do list reflete os itens visíveis, não a slice).
func (m listModel) selectedBackup(backups []Backup) int {
	it, ok := m.list.SelectedItem().(backupItem)
	if !ok {
		return -1
	}
	for i, b := range backups {
		if b == it.backup {
			return i
		}
	}
	return -1
}

// runManage abre a TUI de gestão do backups.json: a lista roda em bubbletea
// e as ações (adicionar/editar/remover) rodam em formulários huh standalone,
// fora do programa — I/O nunca acontece dentro do Update.
func runManage(jsonPath string) error {
	for {
		backups, err := loadBackupsOrEmpty(jsonPath)
		if err != nil {
			return err
		}

		final, err := tea.NewProgram(newListModel(backups)).Run()
		if err != nil {
			return fmt.Errorf("TUI: %w", err)
		}
		m := final.(listModel)

		switch m.action {
		case actionQuit:
			return nil
		case actionAdd:
			b, ok, err := backupForm(Backup{})
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			backups = append(backups, b)
		case actionEdit:
			idx := m.selectedBackup(backups)
			if idx < 0 {
				continue
			}
			b, ok, err := backupForm(backups[idx])
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			backups[idx] = b
		case actionDelete:
			idx := m.selectedBackup(backups)
			if idx < 0 {
				continue
			}
			ok, err := confirmDelete(backups[idx])
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			backups = append(backups[:idx], backups[idx+1:]...)
		}

		if err := SaveBackups(jsonPath, backups); err != nil {
			return err
		}
	}
}
