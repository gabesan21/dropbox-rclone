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
		Name:          "dados",
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
			Name:          "logs",
			RepeatCicle:   "6h",
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
		{"name vazio", func(b *Backup) { b.Name = "  " }},
		{"path vazio", func(b *Backup) { b.Path = "  " }},
		{"rclone_account vazio", func(b *Backup) { b.RcloneAccount = "" }},
		{"remote_path vazio", func(b *Backup) { b.RemotePath = "" }},
		{"horário inválido", func(b *Backup) { b.BackupTime = "25:00" }},
		{"horário sem minutos", func(b *Backup) { b.BackupTime = "3" }},
		{"repeat_cicle fora do enum", func(b *Backup) { b.RepeatCicle = "5m" }},
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

func TestValidateBackupRepeatCicleValidos(t *testing.T) {
	// Vazio (default 24h) e todos os ciclos do enum são aceitos.
	for _, cicle := range append([]string{""}, repeatCicles...) {
		b := validBackup()
		b.RepeatCicle = cicle
		if err := ValidateBackup(b); err != nil {
			t.Errorf("repeat_cicle %q deveria ser válido: %v", cicle, err)
		}
	}
}

func TestValidateBackupIgnoraMaxBackupsForaDeCompacted(t *testing.T) {
	// max_backups só é conferido no tipo compacted; nos demais é ignorado.
	for _, typ := range []string{"folder-backup", "folder-sync"} {
		b := validBackup()
		b.Type = typ
		b.MaxBackups = 0
		if err := ValidateBackup(b); err != nil {
			t.Errorf("type %q com max_backups 0 deveria ser válido: %v", typ, err)
		}
	}
}

func TestValidateBackupsAgregaProblemas(t *testing.T) {
	backups := []Backup{
		// Entrada 0: sem name e repeat_cicle fora do enum.
		{Path: "/a", RcloneAccount: "db", RemotePath: "r/a", BackupTime: "03:00", RepeatCicle: "5m", Type: "folder-sync"},
		// Entrada 1: válida, mas o nome é reutilizado pela entrada 2.
		{Name: "dup", Path: "/b", RcloneAccount: "db", RemotePath: "r/b", BackupTime: "04:00", Type: "folder-backup"},
		// Entrada 2: name duplicado e ciclo menor que o intervalo do timer.
		{Name: "dup", Path: "/c", RcloneAccount: "db", RemotePath: "r/c", BackupTime: "05:00", RepeatCicle: "15m", MaxBackups: 1, Type: "compacted"},
	}

	issues := ValidateBackups(backups, 30)

	want := []struct {
		index int
		field string
	}{
		{0, "name"},
		{0, "repeat_cicle"},
		{2, "name"},
		{2, "repeat_cicle"},
	}

	if len(issues) != len(want) {
		t.Fatalf("problemas: got %d (%v), want %d", len(issues), issues, len(want))
	}
	for i, w := range want {
		if issues[i].Index != w.index || issues[i].Field != w.field {
			t.Errorf("problema %d: got entrada %d campo %s, want entrada %d campo %s",
				i, issues[i].Index, issues[i].Field, w.index, w.field)
		}
	}
}

func TestValidateBackupsCicloMenorQueIntervalo(t *testing.T) {
	b := validBackup()
	b.RepeatCicle = "15m"

	// Ciclo 15m é inválido com intervalo de 30min, válido com 15min.
	if issues := ValidateBackups([]Backup{b}, 30); len(issues) != 1 || issues[0].Field != "repeat_cicle" {
		t.Errorf("esperava 1 problema em repeat_cicle, obteve %v", issues)
	}
	if issues := ValidateBackups([]Backup{b}, 15); len(issues) != 0 {
		t.Errorf("esperava configuração limpa, obteve %v", issues)
	}
}

func TestValidateBackupsLimpa(t *testing.T) {
	issues := ValidateBackups([]Backup{validBackup()}, 30)
	if len(issues) != 0 {
		t.Errorf("esperava 0 problemas, obteve %v", issues)
	}
}
