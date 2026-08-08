# Frente F02 — force/restore e service.sh — [[M-1.1-ciclo-repeticao-e-restore]]

- **Entrega:** subcomandos `force <name>` e `restore <name> --yes` no binário, restore por tipo com guards, repasse `force|restore|validate` no service.sh, testes shell atualizados.
- **Escopo:** operação manual (force/restore) e o script central — não toca agendamento, validação (F01) nem TUI/docs (F03).
- **Responsável:** agent.
- **Owns:** `main.go`, `executor.go`, `service.sh`, `tests/test_service.sh`, e um arquivo Go novo de restore se fizer sentido (ex.: `restore.go` + `restore_test.go`).
- **May read:** card (O quê/Por quê), [[pop/specs/backup-config-e-cli|spec backup-config-e-cli]], `backup.go`/`store.go` (struct e validação pós-F01), `tests/test_setup.sh` (padrão das suítes shell).
- **Must not edit:** `backup.go`, `store.go`, `backup_test.go`, `store_test.go`, `tui.go`, `tui_form.go`, `README.md`, `backups.json`, `pop/`, demais arquivos de `tests/`.
- **Depends on:** F01.
- **Entrada esperada:** struct `Backup` final com `Name`/`RepeatCicle`, dispatch de subcomandos em `main.go` estendível, validação agregada da F01. Ausente/incompatível → `BLOCKED`.
- **Skills:** [[clean-code-change]] — *use ao planejar e executar as edições de código*.
- **Critérios:** 3, 4, 5, 7 do [[M-1.1-ciclo-repeticao-e-restore.plan|plano]].

## Decisões vinculantes (do plano — não reabrir)

- Lookup por `name`: argumento ausente, nome desconhecido ou entrada sem `name` → erro em stderr (exit 1) listando os nomes disponíveis.
- `force <name>`: executa `RunBackup` da entrada imediatamente, ignorando janela/ciclo; exit code reflete sucesso/falha.
- `restore <name> --yes`: sem `--yes`, aborta explicando a destrutividade (exit 2). Guards: recusa `path` vazio, `/`, inexistente ou não-diretório.
  - `folder-backup` e `folder-sync`: `rclone sync` remoto → local (remoto vira referência; o que só existe localmente é removido = "limpa o conteúdo").
  - `compacted`: lista o remoto com o mesmo prefixo da rotação (`<base>-*.tar.gz`), baixa o mais recente para temp, limpa o conteúdo da pasta local e extrai o arquivo nela; remove o temp ao final.
- `service.sh` ganha `force`, `restore` e `validate` repassando argumentos ao binário (incl. `--yes`), com texto de uso atualizado; mantém o padrão de erro dos comandos atuais. Nada muda em install/enable/disable/status/remove.
- `tests/test_service.sh`: novos casos cobrindo o repasse (ex.: binário stub em `BINARY`/PATH ou checagem de dispatch/uso), seguindo o padrão pass/fail da suíte; `bash -n` continua verde.
- Textos de uso do binário e do script listam os comandos novos.

## Contrato de execução

- Entregar somente o escopo e os critérios desta frente.
- **Saber parar:** no máximo 2 tentativas por critério `agent` quando a falha é de ambiente (rclone ausente na máquina não é defeito); na segunda, registre `ambiente`, reporte reclassificação para `verify: user` e siga.
- Dependência ou entrada ausente/incompatível → responder `BLOCKED` ao orquestrador com evidência.
- Não alterar caminhos fora de `Owns`; necessidade nova volta ao orquestrador.

## Resultado

- **Status:** concluída.
- **Commit/artefato:** binário `dropbox-rclone` recompilado com `force`/`restore`.
- **Arquivos alterados:** `main.go` (dispatch + `runForce`/`runRestore`/`findByName`/`backupNames`/`exitUnknownName`, uso atualizado), `restore.go` (novo — `RunRestore` com guards de path, restore por tipo, `pickLatestArchive`, `cleanDir`), `restore_test.go` (novo), `service.sh` (repasse `force|restore|validate`, uso atualizado), `tests/test_service.sh` (4 casos novos: uso lista comandos + 3 repasses com binário stub).
- **Desvios:** nenhum. Sem `--yes`, o restore aborta com exit 2 antes de carregar o JSON (flag verificada primeiro); com `--yes` e nome desconhecido, exit 1 listando nomes.
- **Evidência:** `gofmt -l .` vazio; `go vet ./...` limpo; `go test -count=1 ./...` ok; `bash tests/test_service.sh` 12 pass / 0 fail; `./dropbox-rclone restore qualquer-nome` (sem `--yes`) → exit 2 com explicação da destrutividade; `./dropbox-rclone force nome-inexistente` e `restore nome-inexistente --yes` → exit 1 com "Nomes disponíveis: dados-compacted, nginx-config". rclone presente na máquina (/usr/bin/rclone); execução real contra o remoto não foi feita por ser destrutiva.
