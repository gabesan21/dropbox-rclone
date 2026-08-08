# Epoch 4 — Robustez do compacted e tipo full-folder

- **Projeto:** [[pop/PROJECT|dropbox-rclone]] · **Roadmap:** [[pop/ROADMAP|Roadmap]]
- **Status:** em andamento
- **Descrição:** compacted para de derrubar a máquina (memória + pré-check de disco) e nasce o tipo full-folder, cópia datada sem compactação com rotação por max_backups.
- **Yolo:** sim

> Uma phase por seção; sob cada phase, somente suas tasks ainda abertas — **sempre descrições de uma linha**. Detalhe vai para a spec ou para a pasta da task no kanban. Task iniciada ganha link `[[<id>]]`; ao concluir o `005_closing`, sai da tabela depois de sua memory válida (ver [[WORKFLOW|WORKFLOW]]).
> **Yolo herda:** epoch yolo → phases e tasks herdam; phase yolo → tasks herdam. Opt-out/opt-in por task: anexe ` · yolo: não` (ou ` · yolo: sim`) ao fim da célula Descrição — sem coluna nova. O `new-task` resolve a herança e estampa o card (seção Yolo do [[WORKFLOW|WORKFLOW]]).
> **Size:** o agente sugere `S|M|L` na Descrição; `new-task` estampa no card e o humano corrige em 001. Size orienta tier/esforço; risco, skills, dependências e write sets determinam a topologia no [[WORKFLOW|WORKFLOW]].

## Recon e forks

- Incidente real (servidor do usuário, 2026-08-08): `force` de um compacted morreu com `gzip: stdout: No space left on device` e deixou `.tar.gz` parcial em `/tmp`; em `/tmp` tmpfs o arquivo temporário consome RAM — é a raiz provável do "consome toda a memória".
- Fork: se o streaming `tar | rclone rcat` eliminar o temporário, a pré-checagem de disco vira pré-checagem de leitura da origem (ou desaparece) — o plano da task decide.

## Phase 4.1 — Compacted robusto

- **Status:** concluída
- **Descrição:** compactação sem esgotar memória/disco e aborto antecipado quando não houver espaço suficiente antes de começar.
- **Specs:** [[pop/specs/backup-config-e-cli|backup-config-e-cli]]

| Task | Descrição (≤1 linha) | Status |
|------|----------------------|--------|

> **Toda phase termina com a task `verificacao-da-phase`** (`depends_on` todas as demais): é a única em que testes rodam — seção "Verificação de phase" do [[WORKFLOW|WORKFLOW]].

## Phase 4.2 — Tipo full-folder

- **Status:** pendente
- **Descrição:** novo `type: full-folder` — cópia datada do folder sem compactação, com rotação por max_backups como o compacted.
- **Specs:** [[pop/specs/backup-config-e-cli|backup-config-e-cli]]

| Task | Descrição (≤1 linha) | Status |
|------|----------------------|--------|
| [[4.2.2-verificacao-da-phase]] | Escreve/roda a suíte da phase (critérios `verify: phase`) e conserta o que ela pegar. · size: M | 001_initial_task |

> **Toda phase termina com a task `verificacao-da-phase`** (`depends_on` todas as demais): é a única em que testes rodam — seção "Verificação de phase" do [[WORKFLOW|WORKFLOW]].
