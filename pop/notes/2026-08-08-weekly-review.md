---
author: agent
created: 2026-08-08
---

# Weekly review — 2026-08-08

Primeira revisão do escopo. Rodada fora do kanban, via skill `weekly-review`: scripts (`pop_status`, `pop_validate`, `--check-fresh`) + 3 frentes de coleta em paralelo (arquivos base/DOX; specs/notas/links; memory/roadmap/epochs/gate/worktrees).

## Aguardando você

- [[3.2.2-validacao-pelo-humano]] — em `001_initial_task` desde 2026-08-07, liberação desmarcada **de propósito**: é o gate humano de validação em servidor remoto via SSH. Roteiro: [[pop/notes/references/checklist-validacao-remota|checklist de validação remota]].
- [[2026-08-08-drift-harness-gate-adversarial|Drift de harness gerido (gate adversarial)]] — questão aberta criada nesta revisão; a correção é na origem do harness.

## Ajustado nesta revisão

- `pop/notes/2026-08-08-weekly-review.md` — este relatório (classe: harness próprio, consolidação da skill).
- `pop/INBOX.md` — link deste relatório na seção **Revisões** (única seção manual do arquivo).
- `pop/open_questions/2026-08-08-drift-harness-gate-adversarial.md` — achado fora do alcance do escopo virou questão (classe: harness gerido, nunca editado aqui).

Nenhum conserto de conteúdo foi necessário: a varredura não encontrou link morto real, status inconsistente ou arquivo acima do teto. **Durante a coleta, um processo em paralelo corrigiu na working tree** as únicas pendências medidas (entradas `2.1.1-crud-tui-backups.01-tui-manage.md` 811c→≤800c e `2.2.2-verificacao-da-phase.01-suite-service.md` 803c→≤800c; ledger `D-20260807` substituído por `F-20260807-remove-binario-do-git.md` + entrada granular com evidência). Estado verificado, não alterado por esta revisão — falta commitar.

## Medição (tudo limpo)

- `pop_validate --scope .`: válido, nenhuma violação. `--check-fresh`: harness na versão cf1c380c01e2.
- `AGENTS.md`: 90 linhas, descontado o bloco DOX (30) = 60 — no teto. `pop/PROJECT.md`: 36.
- Memory: 35 arquivos em 2 pastas de data, todos ledgers ≤1200c e entradas ≤800c, todas com evidência linkada.
- Roadmap: zero resíduo — epochs 1 e 2 concluídas e limpas, batem com os 28 ledgers. Statuses do `ROADMAP.md` batem com os arquivos de epoch.
- Wikilinks: zero quebrados (alvos inexistentes são por design: `pop/MODIFICATIONS` "criado sob demanda", placeholders `.plan`/`.approval` do estágio 001).
- Worktrees: vazia; branches: só `main`/`origin/main`; sem `develop` órfã.
- Gate adversarial: único card tem `created: 2026-08-07` ≥ corte (`JUDGE_DREDD_SINCE = 2026-08-04`) — **zero cards pré-corte**.

## Parado

- Nada parado além do gate humano por design (3.2.2 acima).

## Progresso

Sem relatório anterior para comparar. Desde a criação do projeto: Epochs 1 (núcleo de backup) e 2 (TUI + serviço systemd) concluídas; Phase 3.1 (README) e task 3.2.1 (checklist de validação) concluídas; fixes diretos F-20260807 (binário fora do tracking do git) e F-20260808 (README renovado no estilo do espelho público). Projeto depende só da validação remota para fechar a Epoch 3.

## Propostas

1. **Inicializar a árvore DOX.** `tests/` tem 3 scripts shell (stack diferente do Go da raiz) — gatilho objetivo de contrato-filho. A inicialização exige varredura e gate humano 003 (seção DOX do `AGENTS.md`); candidata a task quando a Epoch 3 fechar.
2. **Criar as primeiras specs.** Três epochs de código sem nenhum contrato durável: formato de `backups.json`, tipos de backup (compacted/folder/sync) e comportamento do timer/serviço são os candidatos naturais via skill `write-spec`.
3. **Remover a dívida do gate adversarial na origem.** Zero cards pré-corte aqui, mas a remoção conjunta (cláusula no WORKFLOW, ressalva na spec do gate, constante e testes no validador) toca material **gerido** — decisão e execução são da origem do harness, seguidas de reinstalação. Detalhe e pergunta: [[2026-08-08-drift-harness-gate-adversarial|questão aberta]].
4. **Seção "Abandonar/pausar se" ausente nas epochs 2 e 3** (só a Epoch 1 tem, `roadmap/1-nucleo-backup.md:7`, condição não atingida). Avaliar ao planejar a próxima epoch — não é obrigatória, mas deixa o critério de abandono explícito.
