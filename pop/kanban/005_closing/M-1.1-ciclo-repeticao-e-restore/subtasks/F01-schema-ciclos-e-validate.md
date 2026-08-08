# Frente F01 — schema, ciclos e validate — [[M-1.1-ciclo-repeticao-e-restore]]

- **Entrega:** struct `Backup` com `name` e `repeat_cicle`, `FilterDue` por ciclos intra-dia, validação agregada, subcomando `validate` no binário, testes Go atualizados e novos.
- **Escopo:** somente schema de dados, agendamento e validação — nada de force/restore, TUI ou docs.
- **Responsável:** agent.
- **Owns:** `backup.go`, `store.go`, `main.go`, `backup_test.go`, `store_test.go`.
- **May read:** card (O quê/Por quê), [[pop/specs/backup-config-e-cli|spec backup-config-e-cli]], `executor.go` (uso de `MaxBackups` na rotação), `tui_form.go` (uso de `ValidateBackup`).
- **Must not edit:** `executor.go`, `service.sh`, `tui.go`, `tui_form.go`, `README.md`, `backups.json`, `tests/`, `pop/`.
- **Depends on:** nenhuma.
- **Entrada esperada:** nenhuma.
- **Skills:** [[clean-code-change]] — *use ao planejar e executar as edições de código*.
- **Critérios:** 1, 2, 7 do [[M-1.1-ciclo-repeticao-e-restore.plan|plano]].

## Decisões vinculantes (do plano — não reabrir)

- JSON ganha `name` (string) e `repeat_cicle` (string; enum `15m, 30m, 1h, 3h, 6h, 12h, 24h`). `repeat_cicle` ausente ou `24h` = comportamento atual (1x/dia).
- Slots do dia = `backup_time` + k·ciclo enquanto < 24h (todo ciclo divide 24h; ex.: 03:00+`12h` → 03:00 e 15:00; 03:00+`3h` → 03,06,...,21,00). `FilterDue` calcula a próxima ocorrência ≥ now e marca devido se cair na janela [now, now+intervalo); uma execução por janela por entrada, como hoje.
- `ValidateBackup` (fail-fast, usada pela TUI): `name` obrigatório não-vazio; `repeat_cicle` vazio ou no enum; `max_backups >= 1` somente quando `type=compacted` (demais tipos ignoram o campo).
- Validação agregada nova (todas as entradas, todos os problemas: índice, campo, motivo) alimenta o subcomando `validate` do binário: cobre as regras de `ValidateBackup` + unicidade de `name` + `repeat_cicle` ≥ `BACKUP_INTERVAL_MINUTES` (env, default 30). Exit 1 com qualquer problema, 0 se limpo; um problema por linha.
- Entradas legadas sem `name` carregam normalmente (load nunca falha por isso); só a validação aponta.
- Fixtures e testes existentes precisam do campo `name` para continuar verdes; acrescentar testes de ciclo (12h, 3h, default) e da validação agregada (múltiplos problemas, max_backups ignorado fora de compacted).
- `main.go` hoje só aceita `manage`: adicionar `validate` ao dispatch e ao texto de uso; `force`/`restore` são da F02 — deixar o dispatch fácil de estender, sem implementá-los.

## Contrato de execução

- Entregar somente o escopo e os critérios desta frente.
- **Saber parar:** no máximo 2 tentativas por critério `agent` quando a falha é de ambiente; na segunda, registre `ambiente`, reporte reclassificação para `verify: user` e siga.
- Dependência ou entrada ausente/incompatível → responder `BLOCKED` ao orquestrador com evidência.
- Não alterar caminhos fora de `Owns`; necessidade nova volta ao orquestrador.

## Resultado

- **Status:** concluída.
- **Commit/artefato:** binário `/tmp/dropbox-rclone-f01` (build de verificação); merge fica para a rodada do orquestrador.
- **Arquivos alterados:** `backup.go` (campos `Name`/`RepeatCicle`, enum+durações, `nextSlot`, `FilterDue` por ciclos), `store.go` (`ValidationIssue`, `validateFields` compartilhado, `ValidateBackup` fail-fast com as regras novas, `ValidateBackups` agregada), `main.go` (dispatch por switch com `validate`, `backupInterval`, `runValidate`, `fatal`), `backup_test.go`, `store_test.go` (fixtures com `name`, testes de ciclo 12h/3h/24h, validação agregada, max_backups ignorado fora de compacted, legado sem campos novos).
- **Desvios:** nenhum.
- **Evidência:** `gofmt -l .` vazio; `go vet ./...` limpo; `go test ./...` ok. Run do binário: fixture suja → 6 problemas, um por linha (`entrada N: campo: motivo`), exit 1; fixture limpa → "Configuração válida.", exit 0; comando desconhecido → uso, exit 2.
