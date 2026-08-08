# Frente F03 — TUI, documentação e exemplo — [[M-1.1-ciclo-repeticao-e-restore]]

- **Entrega:** formulário e lista da TUI com `name` e `repeat_cicle`, README e `backups.json` de exemplo alinhados ao novo schema e à CLI.
- **Escopo:** TUI e documentação — não toca lógica Go de agendamento/validação/restore nem scripts (F01/F02).
- **Responsável:** agent.
- **Owns:** `tui_form.go`, `tui.go`, `README.md`, `backups.json`.
- **May read:** card (O quê/Por quê), [[pop/specs/backup-config-e-cli|spec backup-config-e-cli]], `backup.go`/`store.go` (campos e regras pós-F01), skills `charm-huh` e `charm-bubbles` (contratos das libs).
- **Must not edit:** `main.go`, `backup.go`, `store.go`, `executor.go`, `*_test.go`, `service.sh`, `tests/`, `pop/`.
- **Depends on:** F01.
- **Entrada esperada:** campos `Name`/`RepeatCicle` na struct e regras finais de `ValidateBackup` (name obrigatório, repeat_cicle no enum, max_backups só compacted). Ausente/incompatível → `BLOCKED`.
- **Skills:** [[clean-code-change]] — *use ao planejar e executar as edições de código*; [[charm-huh]] — *use ao alterar o formulário*; [[charm-bubbles]] — *use ao alterar a lista*.
- **Critérios:** 6, 8 do [[M-1.1-ciclo-repeticao-e-restore.plan|plano]].

## Decisões vinculantes (do plano — não reabrir)

- Formulário: campo `name` (input obrigatório, não-vazio) e `repeat_cicle` (select com o enum `15m…24h`, default `24h`). `max_backups` mantém-se no form com descrição explicitando que só se aplica ao tipo `compacted` (demais tipos ignoram, vale 1).
- Lista: título/descrição do item passam a identificar a entrada pelo `name` (fallback para `path` em entradas legadas sem nome).
- README: tabela de campos do `backups.json` com `name` e `repeat_cicle` (semântica de ciclos + default 24h); seção nova dos comandos `force`/`restore`/`validate` (uso via service.sh, destrutividade do restore e flag `--yes`); `max_backups` marcado como só-compacted; seção "Estrutura do repositório" revisada se os papéis dos arquivos mudarem (inclui corrigir a descrição de `executor.go` se a frente F02 criar `restore.go`).
- `backups.json` (exemplo versionado): entradas ganham `name` e `repeat_cicle` coerentes com os tipos (ex.: `folder-backup` sem depender de `max_backups`).
- Idioma pt-BR em docs e textos da TUI.

## Contrato de execução

- Entregar somente o escopo e os critérios desta frente.
- **Saber parar:** no máximo 2 tentativas por critério `agent` quando a falha é de ambiente; na segunda, registre `ambiente`, reporte reclassificação para `verify: user` e siga.
- Dependência ou entrada ausente/incompatível → responder `BLOCKED` ao orquestrador com evidência.
- Não alterar caminhos fora de `Owns`; necessidade nova volta ao orquestrador.

## Resultado

- **Status:** concluída.
- **Commit/artefato:** working tree (commit a cargo do orquestrador).
- **Arquivos alterados:** `tui_form.go` (input `name` obrigatório, select `repeat_cicle` com enum e default `24h`, descrição de `max_backups` marcando só-compacted, confirmação de remoção pelo nome), `tui.go` (helpers `displayName`/`displayCicle`, título e filtro da lista pelo `name` com fallback para `path`, descrição com `ciclo=`), `README.md` (tabela de campos com `name`/`repeat_cicle` e `max_backups` só-compacted, exemplo JSON sincronizado, seção nova "Operação manual: force, restore e validate", passo 2 + mermaid do "Como funciona" com semântica de slots, "Estrutura do repositório" atualizada para main.go/store.go), `backups.json` (entradas com `name` e `repeat_cicle`; `folder-backup` sem `max_backups`).
- **Desvios:** nenhum. `restore.go` ainda não existia no momento da execução (F02 em paralelo) — descrição de `executor.go` mantida; se F02 criar o arquivo, ajuste fica para o 005.
- **Evidência:** `gofmt -l .` vazio; `go vet ./...` limpo; `go build ./...` ok; `go test ./...` → `ok github.com/gabesan21/dropbox-rclone 0.004s`; `go run . validate` → "Configuração válida."
