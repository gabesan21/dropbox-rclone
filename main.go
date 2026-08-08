package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "manage":
			if err := runManage("backups.json"); err != nil {
				fatal(err)
			}
		case "validate":
			runValidate("backups.json")
		case "force":
			runForce("backups.json", os.Args[2:])
		case "restore":
			runRestore("backups.json", os.Args[2:])
		default:
			fmt.Fprintf(os.Stderr, "uso: dropbox-rclone [manage|validate|force <nome>|restore <nome> --yes]\n")
			os.Exit(2)
		}
		return
	}

	interval := backupInterval()

	backups, err := LoadBackups("backups.json")
	if err != nil {
		fatal(err)
	}

	now := time.Now()
	due, err := FilterDue(backups, now, interval)
	if err != nil {
		fatal(err)
	}

	if len(due) == 0 {
		fmt.Printf("Nenhum backup agendado para o intervalo atual (%s - %s).\n",
			now.Format("15:04"), now.Add(time.Duration(interval)*time.Minute).Format("15:04"))
		return
	}

	fmt.Printf("Backups agendados para %s - %s:\n",
		now.Format("15:04"), now.Add(time.Duration(interval)*time.Minute).Format("15:04"))

	var failed bool
	for _, b := range due {
		fmt.Printf("  - %s -> %s:%s (%s, max=%d)\n",
			b.Path, b.RcloneAccount, b.RemotePath, b.Type, b.MaxBackups)

		fmt.Printf("    Executando...\n")
		if err := RunBackup(b); err != nil {
			fmt.Fprintf(os.Stderr, "    ERRO: %v\n", err)
			failed = true
		} else {
			fmt.Printf("    OK\n")
		}
	}

	if failed {
		os.Exit(1)
	}
}

// backupInterval resolve BACKUP_INTERVAL_MINUTES (default 30 minutos).
func backupInterval() int {
	intervalStr := os.Getenv("BACKUP_INTERVAL_MINUTES")
	if intervalStr == "" {
		return 30
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERRO: BACKUP_INTERVAL_MINUTES inválido: %s\n", intervalStr)
		os.Exit(1)
	}
	return interval
}

// runValidate lista todos os problemas do arquivo de configuração, um por
// linha, e sai com código 1 se houver qualquer problema.
func runValidate(path string) {
	backups, err := LoadBackups(path)
	if err != nil {
		fatal(err)
	}

	issues := ValidateBackups(backups, backupInterval())
	if len(issues) == 0 {
		fmt.Println("Configuração válida.")
		return
	}

	for _, issue := range issues {
		fmt.Println(issue)
	}
	os.Exit(1)
}

// runForce executa imediatamente o backup da entrada com o nome dado,
// ignorando janela e ciclo; o exit code reflete sucesso/falha.
func runForce(configPath string, args []string) {
	backups, err := LoadBackups(configPath)
	if err != nil {
		fatal(err)
	}
	if len(args) < 1 {
		exitUnknownName(backups, "uso: dropbox-rclone force <nome>")
	}
	b, ok := findByName(backups, args[0])
	if !ok {
		exitUnknownName(backups, fmt.Sprintf("backup %q não encontrado", args[0]))
	}

	fmt.Printf("Executando backup de %q (%s) fora do agendamento...\n", b.Name, b.Path)
	if err := RunBackup(b); err != nil {
		fatal(err)
	}
	fmt.Println("OK")
}

// runRestore restaura o conteúdo local da entrada a partir do remoto.
// Sem --yes, aborta explicando a destrutividade (exit 2).
func runRestore(configPath string, args []string) {
	var name string
	var yes bool
	for _, arg := range args {
		switch {
		case arg == "--yes":
			yes = true
		case name == "":
			name = arg
		default:
			fmt.Fprintf(os.Stderr, "ERRO: argumento inesperado %q\nuso: dropbox-rclone restore <nome> --yes\n", arg)
			os.Exit(2)
		}
	}

	if !yes {
		fmt.Fprintln(os.Stderr, "ERRO: restore é destrutivo — limpa o conteúdo local e o repovoa a partir do remoto.")
		fmt.Fprintln(os.Stderr, "Confirme explicitamente com: dropbox-rclone restore <nome> --yes")
		os.Exit(2)
	}

	backups, err := LoadBackups(configPath)
	if err != nil {
		fatal(err)
	}
	if name == "" {
		exitUnknownName(backups, "uso: dropbox-rclone restore <nome> --yes")
	}
	b, ok := findByName(backups, name)
	if !ok {
		exitUnknownName(backups, fmt.Sprintf("backup %q não encontrado", name))
	}

	fmt.Printf("Restaurando %q (%s) a partir do remoto...\n", b.Name, b.Path)
	if err := RunRestore(b); err != nil {
		fatal(err)
	}
	fmt.Println("OK")
}

// findByName localiza a entrada pelo campo name; entradas legadas sem name
// não são endereçáveis por force/restore.
func findByName(backups []Backup, name string) (Backup, bool) {
	for _, b := range backups {
		if name != "" && b.Name == name {
			return b, true
		}
	}
	return Backup{}, false
}

// backupNames lista os nomes disponíveis para force/restore.
func backupNames(backups []Backup) []string {
	var names []string
	for _, b := range backups {
		if strings.TrimSpace(b.Name) != "" {
			names = append(names, b.Name)
		}
	}
	return names
}

// exitUnknownName reporta nome ausente/desconhecido listando os nomes
// disponíveis e sai com código 1.
func exitUnknownName(backups []Backup, msg string) {
	fmt.Fprintf(os.Stderr, "ERRO: %s\n", msg)
	if names := backupNames(backups); len(names) > 0 {
		fmt.Fprintf(os.Stderr, "Nomes disponíveis: %s\n", strings.Join(names, ", "))
	} else {
		fmt.Fprintln(os.Stderr, "Nenhuma entrada com name configurada no backups.json.")
	}
	os.Exit(1)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "ERRO: %v\n", err)
	os.Exit(1)
}
