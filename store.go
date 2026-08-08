package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// backupTypes são os tipos de backup aceitos no backups.json.
var backupTypes = []string{"compacted", "folder-backup", "folder-sync"}

// SaveBackups grava a lista de backups no arquivo JSON, com indentação.
func SaveBackups(path string, backups []Backup) error {
	data, err := json.MarshalIndent(backups, "", "  ")
	if err != nil {
		return fmt.Errorf("serializando backups: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("gravando %s: %w", path, err)
	}
	return nil
}

// ValidateBackup confere os campos de uma entrada antes de persistir.
func ValidateBackup(b Backup) error {
	if strings.TrimSpace(b.Path) == "" {
		return fmt.Errorf("path é obrigatório")
	}
	if strings.TrimSpace(b.RcloneAccount) == "" {
		return fmt.Errorf("rclone_account é obrigatório")
	}
	if strings.TrimSpace(b.RemotePath) == "" {
		return fmt.Errorf("remote_path é obrigatório")
	}
	if _, err := time.Parse("15:04", b.BackupTime); err != nil {
		return fmt.Errorf("backup_time inválido %q: use HH:MM", b.BackupTime)
	}
	if b.MaxBackups < 1 {
		return fmt.Errorf("max_backups deve ser >= 1")
	}
	for _, t := range backupTypes {
		if b.Type == t {
			return nil
		}
	}
	return fmt.Errorf("type inválido %q: use %s", b.Type, strings.Join(backupTypes, ", "))
}

// loadBackupsOrEmpty lê o JSON, tratando arquivo inexistente como lista vazia.
func loadBackupsOrEmpty(path string) ([]Backup, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	return LoadBackups(path)
}
