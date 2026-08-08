package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// RunRestore restaura o conteúdo local a partir do remoto, que vira a
// referência: o que só existe localmente é removido. Operação destrutiva —
// o chamador (CLI) exige confirmação explícita (--yes).
func RunRestore(b Backup) error {
	if err := guardRestorePath(b.Path); err != nil {
		return err
	}

	switch b.Type {
	case "compacted":
		return restoreCompacted(b)
	case "full-folder":
		return restoreFullFolder(b)
	case "folder-backup", "folder-sync":
		return restoreSync(b)
	default:
		return fmt.Errorf("tipo de backup desconhecido: %s", b.Type)
	}
}

// guardRestorePath recusa alvos perigosos ou inválidos para o restore.
func guardRestorePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path vazio: restore exige um diretório local")
	}
	if filepath.Clean(path) == "/" {
		return fmt.Errorf("path %q recusado: restaurar na raiz destruiria o sistema", path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path %q inacessível: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path %q não é um diretório", path)
	}
	return nil
}

// restoreSync espelha o remoto no local (unidirecional, remoto como referência).
func restoreSync(b Backup) error {
	remote := fmt.Sprintf("%s:%s", b.RcloneAccount, b.RemotePath)
	cmd := exec.Command("rclone", "sync", remote, b.Path,
		"--progress",
		"--create-empty-src-dirs",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restaurando %s -> %s: %w", remote, b.Path, err)
	}
	return nil
}

// restoreCompacted baixa o .tar.gz mais recente do remoto, limpa a pasta
// local e extrai o arquivo nela.
func restoreCompacted(b Backup) error {
	remote := fmt.Sprintf("%s:%s", b.RcloneAccount, b.RemotePath)
	base := filepath.Base(strings.TrimSuffix(b.Path, "/"))

	latest, err := latestRemoteArchive(remote, base)
	if err != nil {
		return err
	}

	archivePath := filepath.Join(os.TempDir(), latest)
	defer os.Remove(archivePath)

	dlCmd := exec.Command("rclone", "copyto", fmt.Sprintf("%s/%s", remote, latest), archivePath, "--progress")
	dlCmd.Stdout = os.Stdout
	dlCmd.Stderr = os.Stderr
	if err := dlCmd.Run(); err != nil {
		return fmt.Errorf("baixando %s de %s: %w", latest, remote, err)
	}

	if err := cleanDir(b.Path); err != nil {
		return fmt.Errorf("limpando %s: %w", b.Path, err)
	}

	// O tar contém a pasta base no topo; --strip-components=1 despeja o
	// conteúdo diretamente no path restaurado.
	tarCmd := exec.Command("tar", "-xzf", archivePath, "-C", b.Path, "--strip-components=1")
	tarCmd.Stdout = os.Stdout
	tarCmd.Stderr = os.Stderr
	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("extraindo %s em %s: %w", latest, b.Path, err)
	}
	return nil
}

// restoreFullFolder limpa a pasta local e a repovoa com o conteúdo da
// pasta datada <base>-<AAAAMMDD-HHMMSS> mais recente do remoto.
func restoreFullFolder(b Backup) error {
	remote := fmt.Sprintf("%s:%s", b.RcloneAccount, b.RemotePath)
	base := filepath.Base(strings.TrimSuffix(b.Path, "/"))

	latest, err := latestRemoteFolder(remote, base)
	if err != nil {
		return err
	}

	if err := cleanDir(b.Path); err != nil {
		return fmt.Errorf("limpando %s: %w", b.Path, err)
	}

	remoteFolder := fmt.Sprintf("%s/%s", remote, latest)
	cpCmd := exec.Command("rclone", "copy", remoteFolder, b.Path, "--progress")
	cpCmd.Stdout = os.Stdout
	cpCmd.Stderr = os.Stderr
	if err := cpCmd.Run(); err != nil {
		return fmt.Errorf("copiando %s -> %s: %w", remoteFolder, b.Path, err)
	}
	return nil
}

// latestRemoteFolder retorna a pasta <prefix>-* mais recente do remoto.
func latestRemoteFolder(remote, prefix string) (string, error) {
	lsCmd := exec.Command("rclone", "lsf", remote, "--dirs-only")
	out, err := lsCmd.Output()
	if err != nil {
		return "", fmt.Errorf("listando %s: %w", remote, err)
	}
	latest, ok := pickLatestFolder(strings.Split(strings.TrimSpace(string(out)), "\n"), prefix)
	if !ok {
		return "", fmt.Errorf("nenhuma pasta %s-* encontrada em %s", prefix, remote)
	}
	return latest, nil
}

// pickLatestFolder escolhe, entre os nomes listados, a pasta <prefix>-*
// mais recente (o timestamp no nome torna a ordenação lexicográfica
// cronológica). `rclone lsf --dirs-only` devolve nomes com "/" final —
// tratado no filtro e descartado no resultado.
func pickLatestFolder(names []string, prefix string) (string, bool) {
	var folders []string
	for _, name := range names {
		name = strings.TrimSuffix(name, "/")
		if name != "" && strings.HasPrefix(name, prefix+"-") {
			folders = append(folders, name)
		}
	}
	if len(folders) == 0 {
		return "", false
	}
	sort.Strings(folders)
	return folders[len(folders)-1], true
}

// latestRemoteArchive retorna o arquivo <prefix>-*.tar.gz mais recente do remoto.
func latestRemoteArchive(remote, prefix string) (string, error) {
	lsCmd := exec.Command("rclone", "lsf", remote)
	out, err := lsCmd.Output()
	if err != nil {
		return "", fmt.Errorf("listando %s: %w", remote, err)
	}
	latest, ok := pickLatestArchive(strings.Split(strings.TrimSpace(string(out)), "\n"), prefix)
	if !ok {
		return "", fmt.Errorf("nenhum arquivo %s-*.tar.gz encontrado em %s", prefix, remote)
	}
	return latest, nil
}

// pickLatestArchive escolhe, entre os nomes listados, o <prefix>-*.tar.gz
// mais recente (o timestamp no nome torna a ordenação lexicográfica cronológica).
func pickLatestArchive(names []string, prefix string) (string, bool) {
	var archives []string
	for _, name := range names {
		if strings.HasPrefix(name, prefix+"-") && strings.HasSuffix(name, ".tar.gz") {
			archives = append(archives, name)
		}
	}
	if len(archives) == 0 {
		return "", false
	}
	sort.Strings(archives)
	return archives[len(archives)-1], true
}

// cleanDir remove todo o conteúdo de dir, preservando o próprio diretório.
func cleanDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}
