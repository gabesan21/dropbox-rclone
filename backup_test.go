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
			"name": "data",
			"repeat_cicle": "12h",
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
	if b.Name != "data" {
		t.Errorf("Name: esperava data, obteve %s", b.Name)
	}
	if b.RepeatCicle != "12h" {
		t.Errorf("RepeatCicle: esperava 12h, obteve %s", b.RepeatCicle)
	}
}

func TestLoadBackupsLegadoSemCamposNovos(t *testing.T) {
	// Entradas legadas sem name/repeat_cicle carregam normalmente — o load
	// nunca falha por isso; só a validação aponta.
	jsonPath := filepath.Join(t.TempDir(), "backups.json")
	content := `[{"path": "/a", "rclone_account": "db", "remote_path": "r/a", "backup_time": "02:00", "max_backups": 1, "type": "folder-sync"}]`

	if err := os.WriteFile(jsonPath, []byte(content), 0644); err != nil {
		t.Fatalf("criando fixture: %v", err)
	}

	backups, err := LoadBackups(jsonPath)
	if err != nil {
		t.Fatalf("LoadBackups falhou para entrada legada: %v", err)
	}
	if len(backups) != 1 || backups[0].Name != "" || backups[0].RepeatCicle != "" {
		t.Errorf("entrada legada deveria carregar com Name/RepeatCicle vazios, obteve %+v", backups)
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

func TestFilterDueRepeatCicle(t *testing.T) {
	at := func(hour, minute int) time.Time {
		return time.Date(2026, 8, 7, hour, minute, 0, 0, time.UTC)
	}

	cases := []struct {
		name        string
		repeatCicle string
		now         time.Time
		wantDue     bool
	}{
		// 03:00 + 12h → slots 03:00 e 15:00.
		{"12h slot da tarde", "12h", at(14, 50), true},
		{"12h entre slots", "12h", at(3, 30), false},
		{"12h antes do slot da manhã", "12h", at(2, 50), true},
		// 03:00 + 3h → 8 slots/dia: 03,06,09,12,15,18,21,00.
		{"3h slot do meio-dia", "3h", at(11, 50), true},
		{"3h entre slots", "3h", at(12, 30), false},
		{"3h vira o dia no slot 00:00", "3h", at(23, 50), true},
		// Ausente ou 24h → 1x/dia (comportamento histórico).
		{"default 1x/dia já passou", "", at(14, 50), false},
		{"default antes do slot", "", at(2, 50), true},
		{"24h explícito antes do slot", "24h", at(2, 50), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			backups := []Backup{{Path: "/a", BackupTime: "03:00", RepeatCicle: tc.repeatCicle}}
			due, err := FilterDue(backups, tc.now, 30)
			if err != nil {
				t.Fatalf("FilterDue falhou: %v", err)
			}
			if gotDue := len(due) == 1; gotDue != tc.wantDue {
				t.Errorf("devido: got %v, want %v", gotDue, tc.wantDue)
			}
		})
	}
}

func TestFilterDueRepeatCicleInvalido(t *testing.T) {
	backups := []Backup{{Path: "/a", BackupTime: "03:00", RepeatCicle: "5m"}}
	now := time.Date(2026, 8, 7, 14, 15, 0, 0, time.UTC)

	if _, err := FilterDue(backups, now, 30); err == nil {
		t.Error("esperava erro para repeat_cicle inválido")
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
