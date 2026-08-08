---
id: backup-config-e-cli
project: work/dropbox-rclone
domain: backup
kind: contract
status: active
implementation: implemented
origin: "M-1.1-ciclo-repeticao-e-restore"
created: 2026-08-08
updated: 2026-08-08
supersedes: []
superseded_by:
---

# Spec — Schema do backups.json e CLI de operação

## Contrato

Formato do `backups.json` (configuração declarativa dos backups) e a superfície de comandos do binário `dropbox-rclone` e do script central `service.sh`, incluindo a semântica de agendamento por ciclos e as operações manuais `force`, `restore` e `validate`.

## Comportamento esperado

- Dado um `backup_time` e um `repeat_cicle`, quando o binário roda numa janela, então executa a entrada se algum slot do dia (`backup_time` + k·ciclo, < 24h) cair na janela — uma execução por janela por entrada.
- Dado `repeat_cicle` ausente ou `24h`, quando agenda, então comporta-se como 1x/dia no `backup_time` (comportamento histórico).
- Dado `service.sh force <name>`, quando invocado, então o binário executa a entrada daquele nome imediatamente, ignorando janela e ciclo.
- Dado `service.sh restore <name> --yes`, quando invocado, então o conteúdo da pasta local é limpo e repovoado a partir do remoto (compacted: extrai o `.tar.gz` mais recente; demais tipos: `rclone sync` remoto→local).
- Dado `service.sh validate`, quando o JSON tem problemas, então todos são listados (índice, campo, motivo) e o exit code é não-zero; sem problemas, exit 0.
- Dado um backup `compacted`, quando executa, então a pasta é compactada em streaming direto para o remoto (`tar -cz` pipado para `rclone rcat`), sem nenhum arquivo temporário local; em sucesso, o remoto ganha `<pasta>-<AAAAMMDD-HHMMSS>.tar.gz` e a rotação por `max_backups` remove os mais antigos.
- Dada origem de `compacted` inexistente, não-diretório ou sem permissão de leitura, quando o backup inicia, então aborta com erro que identifica a origem, antes de qualquer compactação ou upload.

## Invariantes

- Todo ciclo do enum divide 24h, mantendo o padrão diário de slots estável.
- `name` é obrigatório e único para entradas criadas/editadas; entradas legadas sem `name` continuam carregando, mas falham na validação e não são endereçáveis por `force`/`restore`.
- `max_backups` só tem efeito no tipo `compacted`; nos demais é ignorado e vale 1.
- `repeat_cicle` menor que `BACKUP_INTERVAL_MINUTES` é configuração inválida (slots seriam perdidos).
- `restore` nunca executa sem a flag `--yes` e recusa path vazio, `/`, inexistente ou não-diretório.
- Load do JSON nunca falha por ausência de `name` ou `repeat_cicle` — validação é camada separada do parse.
- Backup `compacted` nunca cria arquivo temporário local, em sucesso ou em falha — a compactação flui em streaming para o remoto.
- Falha na compactação ou no upload de um `compacted` remove o objeto remoto parcial (best-effort, com aviso se a remoção falhar) — nenhum parcial sobra, nem local nem remoto.
- Erro de qualquer lado do pipe (compactação ou upload) derruba o backup — nenhum dos dois erros é engolido pelo stream.

## Interfaces

- **Entrada (`backups.json`):** array de objetos com `path` (string, obrig.), `rclone_account` (string, obrig.), `remote_path` (string, obrig.), `backup_time` (string `HH:MM`, obrig.), `name` (string, obrig. em entradas novas), `repeat_cicle` (`15m|30m|1h|3h|6h|12h|24h`, default `24h`), `max_backups` (int ≥ 1, só `compacted`), `type` (`compacted|folder-backup|folder-sync`).
- **CLI do binário:** `dropbox-rclone` (agendamento por janela) · `manage` (TUI) · `validate` · `force <name>` · `restore <name> --yes`.
- **CLI do service.sh:** comandos systemd atuais + `force|restore|validate` repassados ao binário.
- **Saída:** `validate` imprime um problema por linha; erros de operação vão para stderr com exit não-zero.
- **Compatibilidade:** JSONs antigos sem `name`/`repeat_cicle` seguem agendando normalmente.

## Erros e limites

- **Nome ausente/desconhecido em force/restore:** erro em stderr listando os nomes disponíveis, exit 1.
- **Restore sem `--yes`:** aborta explicando a destrutividade, exit 2.
- **Validate com problemas:** lista completa dos problemas, exit 1.
- **Limite:** uma execução por janela por entrada — ciclos menores que o intervalo do timer são rejeitados na validação, não executados em lote.

## Critérios de conformidade

- [ ] 03:00 com `12h` agenda 03:00 e 15:00; com `3h` agenda 8 slots/dia; sem `repeat_cicle` agenda 1x/dia.
- [ ] `validate` agrega múltiplos problemas de múltiplas entradas e ajusta o exit code.
- [ ] `restore` exige `--yes`, aplica os guards de path e escolhe o `.tar.gz` mais recente no tipo compacted.
- [ ] `max_backups` fora de `compacted` não gera erro nem rotação.
- [ ] `compacted` aborta antes de compactar quando a origem é inexistente, não-diretório ou ilegível; em falha de stream não sobra parcial local nem remoto.

## Fora de escopo

- Instalação de rclone/Go e configuração OAuth do Dropbox (install.sh, setup-rclone.sh, README).
- Ciclo de vida do unit/timer systemd (service.sh install/enable/disable/status/remove), inalterado por este contrato.

## Referências relacionadas

- [`AGENTS.md`](../../AGENTS.md) — *siga antes de alterar qualquer arquivo do código (verificações e regras DOX)*.
