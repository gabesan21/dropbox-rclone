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

// ValidationIssue descreve um problema de configuração de uma entrada.
// Index é a posição da entrada no JSON, ou -1 quando não se aplica.
type ValidationIssue struct {
	Index  int
	Field  string
	Reason string
}

func (i ValidationIssue) String() string {
	if i.Index >= 0 {
		return fmt.Sprintf("entrada %d: %s: %s", i.Index, i.Field, i.Reason)
	}
	return fmt.Sprintf("%s: %s", i.Field, i.Reason)
}

// validateFields aplica todas as regras de campo de uma entrada, sem fail-fast.
func validateFields(b Backup) []ValidationIssue {
	var issues []ValidationIssue
	add := func(field, reason string) {
		issues = append(issues, ValidationIssue{Index: -1, Field: field, Reason: reason})
	}

	if strings.TrimSpace(b.Name) == "" {
		add("name", "obrigatório e não pode ser vazio")
	}
	if strings.TrimSpace(b.Path) == "" {
		add("path", "obrigatório")
	}
	if strings.TrimSpace(b.RcloneAccount) == "" {
		add("rclone_account", "obrigatório")
	}
	if strings.TrimSpace(b.RemotePath) == "" {
		add("remote_path", "obrigatório")
	}
	if _, err := time.Parse("15:04", b.BackupTime); err != nil {
		add("backup_time", fmt.Sprintf("inválido %q: use HH:MM", b.BackupTime))
	}
	if b.RepeatCicle != "" {
		if _, err := cicleDuration(b.RepeatCicle); err != nil {
			add("repeat_cicle", fmt.Sprintf("inválido %q: use %s", b.RepeatCicle, strings.Join(repeatCicles, ", ")))
		}
	}
	// max_backups só tem efeito no tipo compacted; nos demais é ignorado.
	if b.Type == "compacted" && b.MaxBackups < 1 {
		add("max_backups", "deve ser >= 1 no tipo compacted")
	}

	validType := false
	for _, t := range backupTypes {
		if b.Type == t {
			validType = true
			break
		}
	}
	if !validType {
		add("type", fmt.Sprintf("inválido %q: use %s", b.Type, strings.Join(backupTypes, ", ")))
	}

	return issues
}

// ValidateBackup confere os campos de uma entrada antes de persistir (fail-fast).
func ValidateBackup(b Backup) error {
	if issues := validateFields(b); len(issues) > 0 {
		return fmt.Errorf("%s", issues[0])
	}
	return nil
}

// ValidateBackups agrega os problemas de todas as entradas: regras de campo,
// unicidade de name e repeat_cicle >= intervalo do timer (slots menores que
// o intervalo seriam perdidos).
func ValidateBackups(backups []Backup, intervalMinutes int) []ValidationIssue {
	var issues []ValidationIssue
	seenNames := map[string]int{}

	for i, b := range backups {
		for _, issue := range validateFields(b) {
			issue.Index = i
			issues = append(issues, issue)
		}

		name := strings.TrimSpace(b.Name)
		if name != "" {
			if first, dup := seenNames[name]; dup {
				issues = append(issues, ValidationIssue{i, "name",
					fmt.Sprintf("duplicado (já usado na entrada %d)", first)})
			} else {
				seenNames[name] = i
			}
		}

		if cicle, err := cicleDuration(b.RepeatCicle); err == nil && cicle < time.Duration(intervalMinutes)*time.Minute {
			issues = append(issues, ValidationIssue{i, "repeat_cicle",
				fmt.Sprintf("ciclo %s menor que o intervalo do timer (%dmin): slots seriam perdidos", b.RepeatCicle, intervalMinutes)})
		}
	}

	return issues
}

// loadBackupsOrEmpty lê o JSON, tratando arquivo inexistente como lista vazia.
func loadBackupsOrEmpty(path string) ([]Backup, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	return LoadBackups(path)
}
