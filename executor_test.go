package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Suíte da phase 4.1: prova o comportamento do RunCompacted (streaming
// tar -> rclone rcat) sem Dropbox real — binários fake num bin/ de
// t.TempDir(), remoto fake = diretório local. Sem t.Parallel: t.Setenv
// de PATH é incompatível com paralelismo.

// rcloneFake implementa só os subcomandos que o código invoca, mapeando
// "conta:caminho" para $FAKE_REMOTE_ROOT/caminho e registrando cada
// invocação em $FAKE_RCLONE_LOG. Falhas são ligadas por env.
const rcloneFake = `#!/bin/sh
echo "rclone $*" >> "$FAKE_RCLONE_LOG"
cmd="$1"
shift
alvo="${1#*:}"
case "$cmd" in
rcat)
	if [ -n "$FAKE_RCLONE_FAIL_RCAT" ]; then
		head -c 16 >/dev/null 2>&1
		exit 1
	fi
	dest="$FAKE_REMOTE_ROOT/$alvo"
	mkdir -p "$(dirname "$dest")"
	cat > "$dest"
	;;
copy)
	if [ -n "$FAKE_RCLONE_FAIL_COPY" ]; then
		exit 1
	fi
	src="$1"
	dest="$FAKE_REMOTE_ROOT/${2#*:}"
	mkdir -p "$dest"
	cp -r "$src"/. "$dest/"
	;;
deletefile)
	if [ -n "$FAKE_RCLONE_FAIL_DELETEFILE" ]; then
		exit 1
	fi
	rm -f "$FAKE_REMOTE_ROOT/$alvo"
	;;
purge)
	rm -rf "$FAKE_REMOTE_ROOT/$alvo"
	;;
lsf)
	dir="$FAKE_REMOTE_ROOT/$alvo"
	[ -d "$dir" ] || exit 0
	case " $* " in
	*" --dirs-only "*)
		for f in "$dir"/*; do
			[ -d "$f" ] && echo "$(basename "$f")/"
		done
		;;
	*)
		for f in "$dir"/*; do
			[ -f "$f" ] && echo "$(basename "$f");$(stat -c %Y "$f")"
		done
		;;
	esac
	;;
*)
	echo "rclone fake: subcomando não suportado: $cmd" >&2
	exit 2
	;;
esac
`

// tarFakeMorreNoMeio emite bytes parciais e sai não-zero — falha no meio
// do stream, como um tar que encontra erro de leitura depois de iniciar.
const tarFakeMorreNoMeio = `#!/bin/sh
echo "tar $*" >> "$FAKE_TAR_LOG"
printf 'conteudo-parcial-do-stream'
exit 1
`

// tarFakeSoLoga registra a invocação e falha — usado nos casos de aborto
// antecipado, em que o tar nunca deveria ser chamado.
const tarFakeSoLoga = `#!/bin/sh
echo "tar $*" >> "$FAKE_TAR_LOG"
exit 1
`

// ambienteFake isola um teste: bin/ de fakes no PATH, remoto local e logs.
type ambienteFake struct {
	tmp        string
	bin        string
	remoteRoot string
	rcloneLog  string
	tarLog     string
}

func novoAmbienteFake(t *testing.T) *ambienteFake {
	t.Helper()
	tmp := t.TempDir()
	env := &ambienteFake{
		tmp:        tmp,
		bin:        filepath.Join(tmp, "bin"),
		remoteRoot: filepath.Join(tmp, "remote"),
		rcloneLog:  filepath.Join(tmp, "rclone.log"),
		tarLog:     filepath.Join(tmp, "tar.log"),
	}
	for _, dir := range []string{env.bin, env.remoteRoot} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("criando %s: %v", dir, err)
		}
	}
	t.Setenv("FAKE_REMOTE_ROOT", env.remoteRoot)
	t.Setenv("FAKE_RCLONE_LOG", env.rcloneLog)
	t.Setenv("FAKE_TAR_LOG", env.tarLog)
	t.Setenv("PATH", env.bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return env
}

func (e *ambienteFake) instalaFake(t *testing.T, nome, conteudo string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(e.bin, nome), []byte(conteudo), 0755); err != nil {
		t.Fatalf("instalando fake %s: %v", nome, err)
	}
}

// criaOrigem monta um diretório fixture com um arquivo dentro.
func (e *ambienteFake) criaOrigem(t *testing.T) string {
	t.Helper()
	origem := filepath.Join(e.tmp, "origem")
	if err := os.MkdirAll(origem, 0755); err != nil {
		t.Fatalf("criando origem: %v", err)
	}
	if err := os.WriteFile(filepath.Join(origem, "ola.txt"), []byte("conteúdo de teste"), 0644); err != nil {
		t.Fatalf("criando arquivo da origem: %v", err)
	}
	return origem
}

func (e *ambienteFake) backup(origem string, maxBackups int) Backup {
	return Backup{
		Path:          origem,
		RcloneAccount: "fake",
		RemotePath:    "backups/x",
		MaxBackups:    maxBackups,
		Type:          "compacted",
	}
}

func leLog(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatalf("lendo log %s: %v", path, err)
	}
	return string(data)
}

// tarGzEm lista os .tar.gz diretamente sob um diretório (vazio se não existir).
func tarGzEm(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("listando %s: %v", dir, err)
	}
	var nomes []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".tar.gz") {
			nomes = append(nomes, e.Name())
		}
	}
	return nomes
}

// assertSemTarGzFora prova que nenhum .tar.gz ficou em disco local fora do
// remoto fake — o streaming nunca materializa arquivo temporário local.
func (e *ambienteFake) assertSemTarGzFora(t *testing.T, exceto string) {
	t.Helper()
	err := filepath.WalkDir(e.tmp, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".tar.gz") && !strings.HasPrefix(path, exceto+string(os.PathSeparator)) {
			t.Errorf(".tar.gz inesperado em disco local: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("varrendo %s: %v", e.tmp, err)
	}
}

// assertAbortoSemPipe cobre os casos de falha rápida: erro identifica a
// origem e nem tar nem rclone chegaram a ser invocados.
func assertAbortoSemPipe(t *testing.T, env *ambienteFake, b Backup, origem string) {
	t.Helper()
	err := RunCompacted(b)
	if err == nil {
		t.Fatal("esperava erro de aborto, obteve sucesso")
	}
	if !strings.Contains(err.Error(), origem) {
		t.Errorf("erro deveria identificar a origem %s: %v", origem, err)
	}
	if log := leLog(t, env.rcloneLog); log != "" {
		t.Errorf("rclone não deveria ter sido invocado, log: %s", log)
	}
	if log := leLog(t, env.tarLog); log != "" {
		t.Errorf("tar não deveria ter sido invocado, log: %s", log)
	}
}

func TestRunCompactedOrigemInexistente(t *testing.T) {
	// Origem que não existe aborta antes de qualquer compactação/upload.
	env := novoAmbienteFake(t)
	env.instalaFake(t, "rclone", rcloneFake)
	env.instalaFake(t, "tar", tarFakeSoLoga)
	origem := filepath.Join(env.tmp, "nao-existe")

	assertAbortoSemPipe(t, env, env.backup(origem, 5), origem)
}

func TestRunCompactedOrigemArquivoRegular(t *testing.T) {
	// Origem que é arquivo regular (não-diretório) aborta da mesma forma.
	env := novoAmbienteFake(t)
	env.instalaFake(t, "rclone", rcloneFake)
	env.instalaFake(t, "tar", tarFakeSoLoga)
	origem := filepath.Join(env.tmp, "arquivo.txt")
	if err := os.WriteFile(origem, []byte("sou um arquivo"), 0644); err != nil {
		t.Fatalf("criando fixture: %v", err)
	}

	assertAbortoSemPipe(t, env, env.backup(origem, 5), origem)
}

func TestRunCompactedOrigemSemPermissaoLeitura(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignora permissão de leitura — caso só é válido sem root")
	}
	// Diretório chmod 000 aborta da mesma forma (critério phase 2).
	env := novoAmbienteFake(t)
	env.instalaFake(t, "rclone", rcloneFake)
	env.instalaFake(t, "tar", tarFakeSoLoga)
	origem := env.criaOrigem(t)
	if err := os.Chmod(origem, 0000); err != nil {
		t.Fatalf("aplicando chmod 000: %v", err)
	}
	// Restaura a permissão para o cleanup do t.TempDir conseguir remover.
	t.Cleanup(func() { os.Chmod(origem, 0755) })

	assertAbortoSemPipe(t, env, env.backup(origem, 5), origem)
}

func TestRunCompactedFalhaNoTar(t *testing.T) {
	// tar morrendo no meio do stream derruba o backup (erro não engolido
	// pelo pipe) e o objeto remoto parcial é removido.
	env := novoAmbienteFake(t)
	env.instalaFake(t, "rclone", rcloneFake)
	env.instalaFake(t, "tar", tarFakeMorreNoMeio)
	origem := env.criaOrigem(t)

	err := RunCompacted(env.backup(origem, 5))
	if err == nil {
		t.Fatal("esperava erro com tar falhando, obteve sucesso")
	}

	if log := leLog(t, env.rcloneLog); !strings.Contains(log, "deletefile") {
		t.Errorf("esperava remoção do parcial remoto, log: %s", log)
	}
	if restos := tarGzEm(t, filepath.Join(env.remoteRoot, "backups/x")); len(restos) != 0 {
		t.Errorf("não deveria sobrar parcial no remoto: %v", restos)
	}
	env.assertSemTarGzFora(t, env.remoteRoot)
}

func TestRunCompactedFalhaNoRcat(t *testing.T) {
	// rclone rcat falhando também derruba o backup e remove o parcial.
	// Aqui o tar é o real do sistema: só o rclone é fake.
	env := novoAmbienteFake(t)
	env.instalaFake(t, "rclone", rcloneFake)
	t.Setenv("FAKE_RCLONE_FAIL_RCAT", "1")
	origem := env.criaOrigem(t)

	err := RunCompacted(env.backup(origem, 5))
	if err == nil {
		t.Fatal("esperava erro com rcat falhando, obteve sucesso")
	}

	if log := leLog(t, env.rcloneLog); !strings.Contains(log, "deletefile") {
		t.Errorf("esperava remoção do parcial remoto, log: %s", log)
	}
	if restos := tarGzEm(t, filepath.Join(env.remoteRoot, "backups/x")); len(restos) != 0 {
		t.Errorf("não deveria sobrar parcial no remoto: %v", restos)
	}
	env.assertSemTarGzFora(t, env.remoteRoot)
}

func TestRunCompactedSucesso(t *testing.T) {
	// Caminho feliz: tar real + rclone fake — prova o stream de ponta a
	// ponta com um .tar.gz genuíno e nenhum temporário local.
	env := novoAmbienteFake(t)
	env.instalaFake(t, "rclone", rcloneFake)
	origem := env.criaOrigem(t)

	if err := RunCompacted(env.backup(origem, 5)); err != nil {
		t.Fatalf("RunCompacted falhou no caminho feliz: %v", err)
	}

	remoto := filepath.Join(env.remoteRoot, "backups/x")
	arquivos := tarGzEm(t, remoto)
	if len(arquivos) != 1 {
		t.Fatalf("esperava exatamente 1 arquivo no remoto, obteve %v", arquivos)
	}
	nome := arquivos[0]
	if !regexp.MustCompile(`^origem-\d{8}-\d{6}\.tar\.gz$`).MatchString(nome) {
		t.Errorf("nome fora do padrão <pasta>-<AAAAMMDD-HHMMSS>.tar.gz: %s", nome)
	}

	// O conteúdo é um tar.gz válido contendo o arquivo da origem.
	f, err := os.Open(filepath.Join(remoto, nome))
	if err != nil {
		t.Fatalf("abrindo arquivo remoto: %v", err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("arquivo remoto não é gzip válido: %v", err)
	}
	tr := tar.NewReader(gz)
	var entradas []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("lendo tar remoto: %v", err)
		}
		entradas = append(entradas, hdr.Name)
	}
	if !strings.Contains(strings.Join(entradas, " "), "origem/ola.txt") {
		t.Errorf("arquivo da origem ausente no tar remoto, entradas: %v", entradas)
	}

	env.assertSemTarGzFora(t, env.remoteRoot)
}

func TestRunCompactedRotacao(t *testing.T) {
	// Com max_backups=2 e 2 arquivos antigos já no remoto, o sucesso deixa
	// exatamente os 2 mais recentes e remove o mais antigo.
	env := novoAmbienteFake(t)
	env.instalaFake(t, "rclone", rcloneFake)
	origem := env.criaOrigem(t)

	remoto := filepath.Join(env.remoteRoot, "backups/x")
	if err := os.MkdirAll(remoto, 0755); err != nil {
		t.Fatalf("criando remoto fake: %v", err)
	}
	antigos := []string{"origem-20200101-000000.tar.gz", "origem-20210101-000000.tar.gz"}
	for _, nome := range antigos {
		if err := os.WriteFile(filepath.Join(remoto, nome), []byte("backup antigo"), 0644); err != nil {
			t.Fatalf("semeando remoto: %v", err)
		}
	}

	if err := RunCompacted(env.backup(origem, 2)); err != nil {
		t.Fatalf("RunCompacted falhou: %v", err)
	}

	restantes := tarGzEm(t, remoto)
	if len(restantes) != 2 {
		t.Fatalf("esperava 2 arquivos após rotação, obteve %v", restantes)
	}
	for _, nome := range restantes {
		if nome == antigos[0] {
			t.Errorf("o mais antigo deveria ter sido removido: %v", restantes)
		}
	}
	if log := leLog(t, env.rcloneLog); !strings.Contains(log, "deletefile fake:backups/x/"+antigos[0]) {
		t.Errorf("esperava deletefile do mais antigo no log: %s", log)
	}
}

// Suíte da task 4.2.1: prova o RunFullFolder (rclone copy para pasta datada
// + rotação por purge) sobre o mesmo ambiente fake — pasta remota fake =
// diretório local. Sem t.Parallel pelos mesmos motivos da suíte compacted.

// dirsEm lista os diretórios diretamente sob um diretório (vazio se não existir).
func dirsEm(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("listando %s: %v", dir, err)
	}
	var nomes []string
	for _, e := range entries {
		if e.IsDir() {
			nomes = append(nomes, e.Name())
		}
	}
	return nomes
}

func (e *ambienteFake) backupFullFolder(origem string, maxBackups int) Backup {
	b := e.backup(origem, maxBackups)
	b.Type = "full-folder"
	return b
}

func TestRunFullFolderSucesso(t *testing.T) {
	// Caminho feliz: a origem é copiada para <base>-<AAAAMMDD-HHMMSS>/ no
	// remoto, com o conteúdo intacto — nenhuma compactação envolvida.
	env := novoAmbienteFake(t)
	env.instalaFake(t, "rclone", rcloneFake)
	origem := env.criaOrigem(t)

	if err := RunFullFolder(env.backupFullFolder(origem, 5)); err != nil {
		t.Fatalf("RunFullFolder falhou no caminho feliz: %v", err)
	}

	remoto := filepath.Join(env.remoteRoot, "backups/x")
	pastas := dirsEm(t, remoto)
	if len(pastas) != 1 {
		t.Fatalf("esperava exatamente 1 pasta no remoto, obteve %v", pastas)
	}
	if !regexp.MustCompile(`^origem-\d{8}-\d{6}$`).MatchString(pastas[0]) {
		t.Errorf("nome fora do padrão <pasta>-<AAAAMMDD-HHMMSS>: %s", pastas[0])
	}

	conteudo, err := os.ReadFile(filepath.Join(remoto, pastas[0], "ola.txt"))
	if err != nil {
		t.Fatalf("arquivo da origem ausente na pasta remota: %v", err)
	}
	if string(conteudo) != "conteúdo de teste" {
		t.Errorf("conteúdo divergente na pasta remota: %q", conteudo)
	}
	if log := leLog(t, env.tarLog); log != "" {
		t.Errorf("tar não deveria ter sido invocado, log: %s", log)
	}
}

func TestRunFullFolderOrigemInexistente(t *testing.T) {
	// Origem inexistente aborta antes de qualquer cópia, com erro que a identifica.
	env := novoAmbienteFake(t)
	env.instalaFake(t, "rclone", rcloneFake)
	origem := filepath.Join(env.tmp, "nao-existe")

	err := RunFullFolder(env.backupFullFolder(origem, 5))
	if err == nil {
		t.Fatal("esperava erro de aborto, obteve sucesso")
	}
	if !strings.Contains(err.Error(), origem) {
		t.Errorf("erro deveria identificar a origem %s: %v", origem, err)
	}
	if log := leLog(t, env.rcloneLog); log != "" {
		t.Errorf("rclone não deveria ter sido invocado, log: %s", log)
	}
}

func TestRunFullFolderFalhaNoCopy(t *testing.T) {
	// rclone copy falhando derruba o backup e a rotação não roda.
	env := novoAmbienteFake(t)
	env.instalaFake(t, "rclone", rcloneFake)
	t.Setenv("FAKE_RCLONE_FAIL_COPY", "1")
	origem := env.criaOrigem(t)

	if err := RunFullFolder(env.backupFullFolder(origem, 5)); err == nil {
		t.Fatal("esperava erro com copy falhando, obteve sucesso")
	}
	if log := leLog(t, env.rcloneLog); strings.Contains(log, "purge") {
		t.Errorf("rotação não deveria rodar após falha no copy, log: %s", log)
	}
}

func TestRunFullFolderRotacao(t *testing.T) {
	// Com max_backups=2 e 2 pastas antigas já no remoto, o sucesso deixa
	// exatamente as 2 mais recentes e faz purge da mais antiga.
	env := novoAmbienteFake(t)
	env.instalaFake(t, "rclone", rcloneFake)
	origem := env.criaOrigem(t)

	remoto := filepath.Join(env.remoteRoot, "backups/x")
	antigas := []string{"origem-20200101-000000", "origem-20210101-000000"}
	for _, nome := range antigas {
		if err := os.MkdirAll(filepath.Join(remoto, nome), 0755); err != nil {
			t.Fatalf("semeando remoto: %v", err)
		}
	}

	if err := RunFullFolder(env.backupFullFolder(origem, 2)); err != nil {
		t.Fatalf("RunFullFolder falhou: %v", err)
	}

	restantes := dirsEm(t, remoto)
	if len(restantes) != 2 {
		t.Fatalf("esperava 2 pastas após rotação, obteve %v", restantes)
	}
	for _, nome := range restantes {
		if nome == antigas[0] {
			t.Errorf("a mais antiga deveria ter sido removida: %v", restantes)
		}
	}
	if log := leLog(t, env.rcloneLog); !strings.Contains(log, "purge fake:backups/x/"+antigas[0]) {
		t.Errorf("esperava purge da mais antiga no log: %s", log)
	}
}

func TestPastasExcedentes(t *testing.T) {
	// Seleção pura da rotação de pastas: filtra pelo prefixo, aceita nomes
	// com e sem "/" final (lsf --dirs-only devolve com "/"), ignora o que
	// não é pasta datada do backup e devolve as mais antigas a remover.
	cases := []struct {
		name       string
		names      []string
		maxBackups int
		want       []string
	}{
		{
			name:       "com e sem barra final",
			names:      []string{"origem-20200101-000000/", "origem-20210101-000000", "origem-20220101-000000/"},
			maxBackups: 2,
			want:       []string{"origem-20200101-000000"},
		},
		{
			name:       "sem excedente",
			names:      []string{"origem-20200101-000000/"},
			maxBackups: 2,
			want:       nil,
		},
		{
			name:       "ignora outros prefixos e arquivos",
			names:      []string{"origem-20200101-000000/", "origem-20210101-000000/", "outra-20190101-000000/", "origem-20200101-000000.tar.gz"},
			maxBackups: 1,
			want:       []string{"origem-20200101-000000"},
		},
		{
			name:       "ordena cronologicamente pelo nome",
			names:      []string{"origem-20220101-000000/", "origem-20200101-000000/", "origem-20210101-000000/"},
			maxBackups: 1,
			want:       []string{"origem-20200101-000000", "origem-20210101-000000"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := pastasExcedentes(tc.names, "origem", tc.maxBackups)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("posição %d: got %s, want %s", i, got[i], tc.want[i])
				}
			}
		})
	}
}
