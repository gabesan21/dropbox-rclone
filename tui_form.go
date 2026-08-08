package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/huh/v2"
)

// backupForm abre o formulário huh de criação/edição de uma entrada.
// ok é false quando o usuário aborta (esc/ctrl+c) — cancelamento normal.
func backupForm(initial Backup) (b Backup, ok bool, err error) {
	b = initial
	if b.BackupTime == "" {
		b.BackupTime = "03:00"
	}
	if b.Type == "" {
		b.Type = backupTypes[0]
	}
	if b.RepeatCicle == "" {
		b.RepeatCicle = "24h"
	}
	maxStr := "5"
	if b.MaxBackups > 0 {
		maxStr = strconv.Itoa(b.MaxBackups)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Nome").
				Description("Identificador único da entrada (usado pelos comandos force/restore).").
				Value(&b.Name).
				Validate(huh.ValidateNotEmpty()),
			huh.NewInput().
				Title("Path local").
				Description("Diretório ou arquivo local de origem.").
				Value(&b.Path).
				Validate(huh.ValidateNotEmpty()),
			huh.NewInput().
				Title("Conta rclone").
				Description("Nome do remote configurado no rclone (ex.: dropbox).").
				Value(&b.RcloneAccount).
				Validate(huh.ValidateNotEmpty()),
			huh.NewInput().
				Title("Path remoto").
				Description("Diretório de destino no remote.").
				Value(&b.RemotePath).
				Validate(huh.ValidateNotEmpty()),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Horário do backup (HH:MM)").
				Value(&b.BackupTime).
				Validate(func(s string) error {
					if _, err := time.Parse("15:04", s); err != nil {
						return fmt.Errorf("use o formato HH:MM")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Ciclo de repetição").
				Description("Intervalo entre execuções ao longo do dia (24h = 1x/dia no horário).").
				Options(huh.NewOptions(repeatCicles...)...).
				Value(&b.RepeatCicle),
			huh.NewInput().
				Title("Máximo de backups").
				Description("Quantos backups manter no remoto. Só se aplica ao tipo compacted; os demais tipos ignoram (vale 1).").
				Value(&maxStr).
				Validate(func(s string) error {
					n, err := strconv.Atoi(strings.TrimSpace(s))
					if err != nil || n < 1 {
						return fmt.Errorf("use um inteiro >= 1")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Tipo de backup").
				Options(huh.NewOptions(backupTypes...)...).
				Value(&b.Type),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return Backup{}, false, nil
		}
		return Backup{}, false, fmt.Errorf("formulário: %w", err)
	}

	b.MaxBackups, _ = strconv.Atoi(strings.TrimSpace(maxStr))
	if err := ValidateBackup(b); err != nil {
		return Backup{}, false, err
	}
	return b, true, nil
}

// confirmDelete pede confirmação antes de remover uma entrada.
// ok é false quando o usuário nega ou aborta.
func confirmDelete(b Backup) (ok bool, err error) {
	var confirm bool
	err = huh.NewConfirm().
		Title(fmt.Sprintf("Remover o backup %s?", displayName(b))).
		Affirmative("Remover").
		Negative("Cancelar").
		Value(&confirm).
		Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, nil
		}
		return false, fmt.Errorf("confirmação: %w", err)
	}
	return confirm, nil
}
