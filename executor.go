package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RunCompacted executa backup compactado: gera arquivo .tar.gz datado, sobe para o remoto
// e remove os mais antigos quando exceder MaxBackups.
func RunCompacted(b Backup) error {
	// Nome do arquivo: <nome-da-pasta>-<AAAAMMDD-HHMMSS>.tar.gz
	base := filepath.Base(strings.TrimSuffix(b.Path, "/"))
	stamp := time.Now().Format("20060102-150405")
	archiveName := fmt.Sprintf("%s-%s.tar.gz", base, stamp)
	archivePath := filepath.Join(os.TempDir(), archiveName)

	// Compacta a pasta local.
	tarCmd := exec.Command("tar", "-czf", archivePath, "-C", filepath.Dir(b.Path), base)
	tarCmd.Stdout = os.Stdout
	tarCmd.Stderr = os.Stderr
	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("compactando %s: %w", b.Path, err)
	}
	defer os.Remove(archivePath)

	// Sobe para o remoto.
	remote := fmt.Sprintf("%s:%s", b.RcloneAccount, b.RemotePath)
	copyCmd := exec.Command("rclone", "copy", archivePath, remote, "--progress")
	copyCmd.Stdout = os.Stdout
	copyCmd.Stderr = os.Stderr
	if err := copyCmd.Run(); err != nil {
		return fmt.Errorf("subindo %s para %s: %w", archiveName, remote, err)
	}

	// Lista backups existentes no remoto e aplica rotação.
	if err := rotateBackups(remote, base, b.MaxBackups); err != nil {
		return fmt.Errorf("aplicando rotação em %s: %w", remote, err)
	}

	return nil
}

// rotateBackups mantém apenas os MaxBackups arquivos mais recentes no remoto.
func rotateBackups(remote, prefix string, maxBackups int) error {
	if maxBackups <= 0 {
		return nil
	}

	// Lista arquivos no remoto que começam com o prefixo.
	lsCmd := exec.Command("rclone", "lsf", remote, "--format", "pt")
	out, err := lsCmd.Output()
	if err != nil {
		return fmt.Errorf("listando %s: %w", remote, err)
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		// Formato: "nome;timestamp"
		parts := strings.Split(line, ";")
		if len(parts) < 1 {
			continue
		}
		name := parts[0]
		if strings.HasPrefix(name, prefix+"-") && strings.HasSuffix(name, ".tar.gz") {
			files = append(files, name)
		}
	}

	if len(files) <= maxBackups {
		return nil
	}

	// Ordena por nome (que contém timestamp) e remove os mais antigos.
	sort.Strings(files)
	toDelete := files[:len(files)-maxBackups]
	for _, f := range toDelete {
		delCmd := exec.Command("rclone", "deletefile", fmt.Sprintf("%s/%s", remote, f))
		delCmd.Stdout = os.Stdout
		delCmd.Stderr = os.Stderr
		if err := delCmd.Run(); err != nil {
			return fmt.Errorf("removendo %s: %w", f, err)
		}
		fmt.Printf("Removido backup antigo: %s\n", f)
	}

	return nil
}

// RunFolderBackup espelha o conteúdo local no remoto (unidirecional, local como referência).
func RunFolderBackup(b Backup) error {
	remote := fmt.Sprintf("%s:%s", b.RcloneAccount, b.RemotePath)
	cmd := exec.Command("rclone", "sync", b.Path, remote,
		"--progress",
		"--create-empty-src-dirs",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sincronizando %s -> %s: %w", b.Path, remote, err)
	}
	return nil
}

// RunFolderSync sincroniza o conteúdo entre local e remoto respeitando ambas as fontes.
func RunFolderSync(b Backup) error {
	remote := fmt.Sprintf("%s:%s", b.RcloneAccount, b.RemotePath)

	// Tenta bisync (bidirecional real). Se falhar por não estar inicializado,
	// orienta o usuário a rodar com --resync na primeira vez.
	cmd := exec.Command("rclone", "bisync", b.Path, remote,
		"--progress",
		"--create-empty-src-dirs",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sincronizando %s <-> %s: %w\nDica: na primeira execução, rode manualmente: rclone bisync %s %s --resync", b.Path, remote, err, b.Path, remote)
	}
	return nil
}

// RunBackup executa o backup conforme o tipo.
func RunBackup(b Backup) error {
	switch b.Type {
	case "compacted":
		return RunCompacted(b)
	case "folder-backup":
		return RunFolderBackup(b)
	case "folder-sync":
		return RunFolderSync(b)
	default:
		return fmt.Errorf("tipo de backup desconhecido: %s", b.Type)
	}
}
