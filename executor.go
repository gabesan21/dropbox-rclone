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

// RunCompacted executa backup compactado: compacta a pasta em streaming direto
// para o remoto (tar -cz | rclone rcat), sem nenhum arquivo temporário local.
// Aborta antes de iniciar quando a origem não é um diretório legível; em
// qualquer falha no meio do stream remove o objeto remoto parcial. Em sucesso,
// remove os mais antigos quando exceder MaxBackups.
func RunCompacted(b Backup) error {
	src := strings.TrimSuffix(b.Path, "/")
	base := filepath.Base(src)

	// Falha rápida: sem diretório legível, nada é compactado nem enviado.
	if err := checkReadableDir(src); err != nil {
		return err
	}
	logSourceSize(src)

	// Nome do arquivo: <nome-da-pasta>-<AAAAMMDD-HHMMSS>.tar.gz
	stamp := time.Now().Format("20060102-150405")
	archiveName := fmt.Sprintf("%s-%s.tar.gz", base, stamp)
	remote := fmt.Sprintf("%s:%s", b.RcloneAccount, b.RemotePath)
	remoteFile := fmt.Sprintf("%s/%s", remote, archiveName)

	// Pipe tar -> rclone: a saída da compactação alimenta o upload direto.
	tarCmd := exec.Command("tar", "-cz", "-C", filepath.Dir(src), base)
	tarCmd.Stderr = os.Stderr
	tarOut, err := tarCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("preparando compactação de %s: %w", src, err)
	}
	rcatCmd := exec.Command("rclone", "rcat", remoteFile)
	rcatCmd.Stdin = tarOut
	rcatCmd.Stdout = os.Stdout
	rcatCmd.Stderr = os.Stderr

	if err := rcatCmd.Start(); err != nil {
		return fmt.Errorf("iniciando upload para %s: %w", remoteFile, err)
	}
	if err := tarCmd.Start(); err != nil {
		tarOut.Close() // EOF no stdin do rclone, destravando o Wait
		rcatCmd.Wait()
		removeRemotePartial(remoteFile)
		return fmt.Errorf("iniciando compactação de %s: %w", src, err)
	}

	// Espera os dois lados do pipe: erro de qualquer um derruba o backup
	// (equivalente ao pipefail do shell). tar primeiro: seu Wait fecha o
	// pipe, dando EOF ao rclone, que então termina e libera o Wait do rcat.
	tarErr := tarCmd.Wait()
	rcatErr := rcatCmd.Wait()
	if tarErr != nil || rcatErr != nil {
		removeRemotePartial(remoteFile)
		return fmt.Errorf("backup de %s falhou (compactação: %v; upload: %v)", src, tarErr, rcatErr)
	}

	// Lista backups existentes no remoto e aplica rotação.
	if err := rotateBackups(remote, base, b.MaxBackups); err != nil {
		return fmt.Errorf("aplicando rotação em %s: %w", remote, err)
	}

	return nil
}

// checkReadableDir garante que a origem existe, é diretório e pode ser aberta
// para leitura — pré-condição para iniciar o stream.
func checkReadableDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("origem %s inacessível: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("origem %s não é um diretório", path)
	}
	dir, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("origem %s sem permissão de leitura: %w", path, err)
	}
	dir.Close()
	return nil
}

// logSourceSize registra o tamanho estimado da origem como informação —
// best-effort, nunca impede o backup.
func logSourceSize(path string) {
	out, err := exec.Command("du", "-sb", path).Output()
	if err != nil {
		return
	}
	fields := strings.Fields(string(out))
	if len(fields) > 0 {
		fmt.Printf("Tamanho estimado da origem %s: %s bytes\n", path, fields[0])
	}
}

// removeRemotePartial apaga o objeto remoto deixado por um stream que falhou
// no meio — best-effort: se a remoção também falhar, só registra o aviso.
func removeRemotePartial(remoteFile string) {
	delCmd := exec.Command("rclone", "deletefile", remoteFile)
	if err := delCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "aviso: não foi possível remover o parcial remoto %s: %v\n", remoteFile, err)
	}
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
