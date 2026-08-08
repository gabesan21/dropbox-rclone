package main

import (
	"os"
	"path/filepath"
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
