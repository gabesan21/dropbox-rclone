package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Backup representa uma entrada de agendamento no backups.json.
type Backup struct {
	Path          string `json:"path"`
	RcloneAccount string `json:"rclone_account"`
	RemotePath    string `json:"remote_path"`
	BackupTime    string `json:"backup_time"` // formato "HH:MM"
	MaxBackups    int    `json:"max_backups"`
	Type          string `json:"type"` // compacted | folder-backup | folder-sync
}

// LoadBackups lê o arquivo JSON e retorna a lista de backups.
func LoadBackups(path string) ([]Backup, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("lendo %s: %w", path, err)
	}

	var backups []Backup
	if err := json.Unmarshal(data, &backups); err != nil {
		return nil, fmt.Errorf("parseando %s: %w", path, err)
	}

	return backups, nil
}

// ParseBackupTime converte "HH:MM" em time.Time no dia de hoje.
func ParseBackupTime(s string, ref time.Time) (time.Time, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("formato de horário inválido %q: %w", s, err)
	}
	return time.Date(ref.Year(), ref.Month(), ref.Day(), t.Hour(), t.Minute(), 0, 0, ref.Location()), nil
}

// FilterDue retorna os backups cujo horário cai dentro do intervalo [now, now+interval).
func FilterDue(backups []Backup, now time.Time, intervalMinutes int) ([]Backup, error) {
	interval := time.Duration(intervalMinutes) * time.Minute
	windowEnd := now.Add(interval)

	var due []Backup
	for _, b := range backups {
		t, err := ParseBackupTime(b.BackupTime, now)
		if err != nil {
			return nil, err
		}

		// Se o horário já passou hoje, considera o de amanhã (para virada de dia).
		if t.Before(now) {
			t = t.Add(24 * time.Hour)
		}

		if !t.Before(now) && t.Before(windowEnd) {
			due = append(due, b)
		}
	}

	return due, nil
}
