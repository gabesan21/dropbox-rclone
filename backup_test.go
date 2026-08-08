package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadBackups(t *testing.T) {
	tmpdir := t.TempDir()
	jsonPath := filepath.Join(tmpdir, "backups.json")

	content := `[
		{
			"path": "/home/user/data",
			"rclone_account": "dropbox",
			"remote_path": "backups/server/data",
			"backup_time": "02:00",
			"max_backups": 7,
			"type": "compacted"
		}
	]`

	if err := os.WriteFile(jsonPath, []byte(content), 0644); err != nil {
		t.Fatalf("criando fixture: %v", err)
	}

	backups, err := LoadBackups(jsonPath)
	if err != nil {
		t.Fatalf("LoadBackups falhou: %v", err)
	}

	if len(backups) != 1 {
		t.Fatalf("esperava 1 backup, obteve %d", len(backups))
	}

	b := backups[0]
	if b.Path != "/home/user/data" {
		t.Errorf("Path: esperava /home/user/data, obteve %s", b.Path)
	}
	if b.Type != "compacted" {
		t.Errorf("Type: esperava compacted, obteve %s", b.Type)
	}
}

func TestParseBackupTime(t *testing.T) {
	ref := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)

	parsed, err := ParseBackupTime("14:30", ref)
	if err != nil {
		t.Fatalf("ParseBackupTime falhou: %v", err)
	}

	if parsed.Hour() != 14 || parsed.Minute() != 30 {
		t.Errorf("esperava 14:30, obteve %02d:%02d", parsed.Hour(), parsed.Minute())
	}

	_, err = ParseBackupTime("invalid", ref)
	if err == nil {
		t.Error("esperava erro para horário inválido")
	}
}

func TestFilterDue(t *testing.T) {
	now := time.Date(2026, 8, 7, 14, 15, 0, 0, time.UTC)

	backups := []Backup{
		{Path: "/a", BackupTime: "14:00", Type: "compacted"},
		{Path: "/b", BackupTime: "14:20", Type: "folder-backup"},
		{Path: "/c", BackupTime: "15:00", Type: "folder-sync"},
	}

	due, err := FilterDue(backups, now, 30)
	if err != nil {
		t.Fatalf("FilterDue falhou: %v", err)
	}

	// 14:00 já passou (vai para amanhã), 14:20 está dentro, 15:00 está fora.
	if len(due) != 1 {
		t.Fatalf("esperava 1 backup devido, obteve %d", len(due))
	}
	if due[0].Path != "/b" {
		t.Errorf("esperava /b, obteve %s", due[0].Path)
	}
}

func TestRunBackupDispatch(t *testing.T) {
	// Apenas verifica que o dispatch existe e rejeita tipo desconhecido.
	b := Backup{Type: "desconhecido"}
	err := RunBackup(b)
	if err == nil {
		t.Error("esperava erro para tipo desconhecido")
	}
}
