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
	Name          string `json:"name"`
	RepeatCicle   string `json:"repeat_cicle"` // 15m|30m|1h|3h|6h|12h|24h; vazio = 24h (1x/dia)
	MaxBackups    int    `json:"max_backups"`  // só tem efeito no tipo compacted
	Type          string `json:"type"`         // compacted | folder-backup | folder-sync
}

// repeatCicles são os ciclos aceitos; todos dividem 24h, mantendo o padrão
// diário de slots estável.
var repeatCicles = []string{"15m", "30m", "1h", "3h", "6h", "12h", "24h"}

// repeatCicleDurations mapeia cada ciclo do enum para sua duração.
var repeatCicleDurations = map[string]time.Duration{
	"15m": 15 * time.Minute,
	"30m": 30 * time.Minute,
	"1h":  time.Hour,
	"3h":  3 * time.Hour,
	"6h":  6 * time.Hour,
	"12h": 12 * time.Hour,
	"24h": 24 * time.Hour,
}

// cicleDuration resolve a duração do ciclo; vazio equivale a 24h (1x/dia).
func cicleDuration(repeatCicle string) (time.Duration, error) {
	if repeatCicle == "" {
		return 24 * time.Hour, nil
	}
	d, ok := repeatCicleDurations[repeatCicle]
	if !ok {
		return 0, fmt.Errorf("repeat_cicle inválido %q", repeatCicle)
	}
	return d, nil
}

// LoadBackups lê o arquivo JSON e retorna a lista de backups.
// Entradas legadas sem name ou repeat_cicle carregam normalmente — a
// validação é camada separada do parse.
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

// nextSlot retorna a próxima ocorrência >= now entre os slots do dia
// (backup_time + k*ciclo, enquanto < 24h). Como todo ciclo divide 24h,
// os slots se repetem diariamente.
func nextSlot(b Backup, now time.Time) (time.Time, error) {
	base, err := ParseBackupTime(b.BackupTime, now)
	if err != nil {
		return time.Time{}, err
	}
	cicle, err := cicleDuration(b.RepeatCicle)
	if err != nil {
		return time.Time{}, err
	}

	if now.Before(base) {
		return base, nil
	}
	elapsed := now.Sub(base)
	next := base.Add(elapsed / cicle * cicle)
	if next.Before(now) {
		next = next.Add(cicle)
	}
	return next, nil
}

// FilterDue retorna os backups cuja próxima ocorrência cai dentro do
// intervalo [now, now+interval) — uma execução por janela por entrada.
func FilterDue(backups []Backup, now time.Time, intervalMinutes int) ([]Backup, error) {
	interval := time.Duration(intervalMinutes) * time.Minute
	windowEnd := now.Add(interval)

	var due []Backup
	for _, b := range backups {
		next, err := nextSlot(b, now)
		if err != nil {
			return nil, err
		}

		if next.Before(windowEnd) {
			due = append(due, b)
		}
	}

	return due, nil
}
