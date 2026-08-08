package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGuardRestorePathRecusaVazio(t *testing.T) {
	if err := guardRestorePath(""); err == nil {
		t.Fatal("esperava erro para path vazio")
	}
	if err := guardRestorePath("   "); err == nil {
		t.Fatal("esperava erro para path só com espaços")
	}
}

func TestGuardRestorePathRecusaRaiz(t *testing.T) {
	if err := guardRestorePath("/"); err == nil {
		t.Fatal("esperava erro para /")
	}
}

func TestGuardRestorePathRecusaInexistente(t *testing.T) {
	if err := guardRestorePath(filepath.Join(t.TempDir(), "nao-existe")); err == nil {
		t.Fatal("esperava erro para path inexistente")
	}
}

func TestGuardRestorePathRecusaArquivo(t *testing.T) {
	dir := t.TempDir()
	arquivo := filepath.Join(dir, "arquivo.txt")
	if err := os.WriteFile(arquivo, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := guardRestorePath(arquivo); err == nil {
		t.Fatal("esperava erro para path que não é diretório")
	}
}

func TestGuardRestorePathAceitaDiretorio(t *testing.T) {
	if err := guardRestorePath(t.TempDir()); err != nil {
		t.Fatalf("esperava aceitar diretório existente: %v", err)
	}
}

func TestPickLatestArchive(t *testing.T) {
	names := []string{
		"docs-20240101-000000.tar.gz",
		"docs-20240601-120000.tar.gz",
		"docs-20240301-060000.tar.gz",
		"outro-20249999-000000.tar.gz",
		"docs-sem-sufixo.txt",
		"",
	}
	got, ok := pickLatestArchive(names, "docs")
	if !ok || got != "docs-20240601-120000.tar.gz" {
		t.Fatalf("esperava docs-20240601-120000.tar.gz, veio %q (ok=%v)", got, ok)
	}
}

func TestPickLatestArchiveSemArquivoDoPrefixo(t *testing.T) {
	if _, ok := pickLatestArchive([]string{"outro-20240101-000000.tar.gz"}, "docs"); ok {
		t.Fatal("esperava ok=false sem arquivos do prefixo")
	}
}

func TestPickLatestFolder(t *testing.T) {
	names := []string{
		"docs-20240101-000000/",
		"docs-20240601-120000/",
		"docs-20240301-060000/",
		"outro-20249999-000000/",
		"docs-arquivo.tar.gz",
		"",
	}
	got, ok := pickLatestFolder(names, "docs")
	if !ok || got != "docs-20240601-120000" {
		t.Fatalf("esperava docs-20240601-120000, veio %q (ok=%v)", got, ok)
	}
}

func TestPickLatestFolderSemBarraFinal(t *testing.T) {
	got, ok := pickLatestFolder([]string{"docs-20240101-000000", "docs-20240301-060000/"}, "docs")
	if !ok || got != "docs-20240301-060000" {
		t.Fatalf("esperava docs-20240301-060000, veio %q (ok=%v)", got, ok)
	}
}

func TestPickLatestFolderSemPastaDoPrefixo(t *testing.T) {
	if _, ok := pickLatestFolder([]string{"outro-20240101-000000/"}, "docs"); ok {
		t.Fatal("esperava ok=false sem pastas do prefixo")
	}
}

func TestCleanDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := cleanDir(dir); err != nil {
		t.Fatalf("cleanDir: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("esperava diretório vazio, restaram %d entradas", len(entries))
	}
}

// Suíte do restore full-folder (critério phase 4): ponta a ponta via caso
// full-folder do switch de RunRestore, sobre o mesmo ambiente fake — o
// rcloneFake decide a direção do copy pelo prefixo "conta:" no 1º argumento.
// Sem t.Parallel pelos mesmos motivos das suítes de executor.

// semeiaPastaDatada cria <remoto>/<nome> com um arquivo dentro.
func semeiaPastaDatada(t *testing.T, remoto, nome, arquivo, conteudo string) {
	t.Helper()
	pasta := filepath.Join(remoto, nome)
	if err := os.MkdirAll(pasta, 0o755); err != nil {
		t.Fatalf("semeando remoto: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pasta, arquivo), []byte(conteudo), 0o644); err != nil {
		t.Fatalf("semeando remoto: %v", err)
	}
}

func TestRestoreFullFolder(t *testing.T) {
	// O restore limpa o local e o repovoa só com o conteúdo da pasta datada
	// mais recente do remoto; lixo fora do padrão é ignorado na seleção.
	env := novoAmbienteFake(t)
	env.instalaFake(t, "rclone", rcloneFake)

	remoto := filepath.Join(env.remoteRoot, "backups/x")
	semeiaPastaDatada(t, remoto, "origem-20200101-000000", "velho.txt", "backup antigo")
	semeiaPastaDatada(t, remoto, "origem-20210101-000000", "novo.txt", "backup recente")
	semeiaPastaDatada(t, remoto, "origem-lixo", "lixo.txt", "fora do padrão")

	destino := env.criaOrigem(t)

	if err := RunRestore(env.backupFullFolder(destino, 5)); err != nil {
		t.Fatalf("RunRestore falhou no caminho feliz: %v", err)
	}

	restantes := dirsEm(t, destino)
	if len(restantes) != 0 {
		t.Errorf("destino deveria ter só arquivos, restaram pastas: %v", restantes)
	}
	conteudo, err := os.ReadFile(filepath.Join(destino, "novo.txt"))
	if err != nil {
		t.Fatalf("arquivo da pasta mais recente ausente no destino: %v", err)
	}
	if string(conteudo) != "backup recente" {
		t.Errorf("conteúdo divergente no destino: %q", conteudo)
	}
	if _, err := os.Stat(filepath.Join(destino, "velho.txt")); !os.IsNotExist(err) {
		t.Error("arquivo da pasta antiga não deveria estar no destino")
	}
	if _, err := os.Stat(filepath.Join(destino, "ola.txt")); !os.IsNotExist(err) {
		t.Error("conteúdo prévio do destino deveria ter sido limpo")
	}

	log := leLog(t, env.rcloneLog)
	if !strings.Contains(log, "lsf fake:backups/x --dirs-only") {
		t.Errorf("esperava lsf --dirs-only no log: %s", log)
	}
	if !strings.Contains(log, "copy fake:backups/x/origem-20210101-000000") {
		t.Errorf("esperava copy da pasta mais recente no log: %s", log)
	}
}

func TestRestoreFullFolderSemPasta(t *testing.T) {
	// Remoto sem pasta datada do prefixo: erro e destino INTOCADO — a
	// limpeza não roda antes da seleção da pasta.
	env := novoAmbienteFake(t)
	env.instalaFake(t, "rclone", rcloneFake)

	remoto := filepath.Join(env.remoteRoot, "backups/x")
	semeiaPastaDatada(t, remoto, "origem-lixo", "lixo.txt", "fora do padrão")

	destino := env.criaOrigem(t)

	err := RunRestore(env.backupFullFolder(destino, 5))
	if err == nil {
		t.Fatal("esperava erro sem pasta datada no remoto, obteve sucesso")
	}
	if !strings.Contains(err.Error(), "nenhuma pasta origem-*") {
		t.Errorf("erro deveria identificar a ausência de pasta do prefixo: %v", err)
	}

	conteudo, lerr := os.ReadFile(filepath.Join(destino, "ola.txt"))
	if lerr != nil {
		t.Fatalf("destino deveria estar intocado: %v", lerr)
	}
	if string(conteudo) != "conteúdo de teste" {
		t.Errorf("conteúdo do destino foi alterado: %q", conteudo)
	}
	if log := leLog(t, env.rcloneLog); strings.Contains(log, "copy") {
		t.Errorf("copy não deveria ter sido invocado, log: %s", log)
	}
}
