package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "manage" {
			if err := runManage("backups.json"); err != nil {
				fmt.Fprintf(os.Stderr, "ERRO: %v\n", err)
				os.Exit(1)
			}
			return
		}
		fmt.Fprintf(os.Stderr, "uso: dropbox-rclone [manage]\n")
		os.Exit(2)
	}

	intervalStr := os.Getenv("BACKUP_INTERVAL_MINUTES")
	if intervalStr == "" {
		intervalStr = "30"
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERRO: BACKUP_INTERVAL_MINUTES inválido: %s\n", intervalStr)
		os.Exit(1)
	}

	backups, err := LoadBackups("backups.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERRO: %v\n", err)
		os.Exit(1)
	}

	now := time.Now()
	due, err := FilterDue(backups, now, interval)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERRO: %v\n", err)
		os.Exit(1)
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
