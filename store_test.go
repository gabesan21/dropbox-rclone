package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validBackup() Backup {
	return Backup{
		Path:          "/srv/dados",
		RcloneAccount: "dropbox",
		RemotePath:    "backups/dados",
		BackupTime:    "03:00",
		MaxBackups:    5,
		Type:          "compacted",
	}
}

func TestSaveBackupsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backups.json")

	want := []Backup{
		validBackup(),
		{
			Path:          "/var/log/app",
			RcloneAccount: "dropbox",
			RemotePath:    "backups/logs",
			BackupTime:    "23:45",
			MaxBackups:    1,
			Type:          "folder-sync",
		},
	}

	if err := SaveBackups(path, want); err != nil {
		t.Fatalf("SaveBackups: %v", err)
	}

	got, err := LoadBackups(path)
	if err != nil {
		t.Fatalf("LoadBackups: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("entradas: got %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entrada %d: got %+v, want %+v", i, got[i], want[i])
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("lendo arquivo salvo: %v", err)
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Errorf("arquivo salvo sem newline final")
	}
}

func TestValidateBackupValidTypes(t *testing.T) {
	for _, typ := range backupTypes {
		b := validBackup()
		b.Type = typ
		if err := ValidateBackup(b); err != nil {
			t.Errorf("type %q deveria ser válido: %v", typ, err)
		}
	}
}

func TestValidateBackupInvalid(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Backup)
	}{
		{"path vazio", func(b *Backup) { b.Path = "  " }},
		{"rclone_account vazio", func(b *Backup) { b.RcloneAccount = "" }},
		{"remote_path vazio", func(b *Backup) { b.RemotePath = "" }},
		{"horário inválido", func(b *Backup) { b.BackupTime = "25:00" }},
		{"horário sem minutos", func(b *Backup) { b.BackupTime = "3" }},
		{"max_backups zero", func(b *Backup) { b.MaxBackups = 0 }},
		{"max_backups negativo", func(b *Backup) { b.MaxBackups = -2 }},
		{"type desconhecido", func(b *Backup) { b.Type = "mirror" }},
		{"type vazio", func(b *Backup) { b.Type = "" }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := validBackup()
			tc.mutate(&b)
			if err := ValidateBackup(b); err == nil {
				t.Errorf("esperava erro para %s, veio nil", tc.name)
			}
		})
	}
}
