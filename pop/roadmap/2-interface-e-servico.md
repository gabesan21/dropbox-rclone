# Epoch 2 — Interface e serviço

- **Projeto:** [[pop/PROJECT|dropbox-rclone]] · **Roadmap:** [[pop/ROADMAP|Roadmap]]
- **Status:** concluída
- **Descrição:** TUI de gestão do `backups.json` com as libs CHARM e serviço systemd de execução periódica.
- **Yolo:** sim

> Uma phase por seção; sob cada phase, somente suas tasks ainda abertas — **sempre descrições de uma linha**. Detalhe vai para a spec ou para a pasta da task no kanban. Task iniciada ganha link `[[<id>]]`; ao concluir o `005_closing`, sai da tabela depois de sua memory válida (ver [[WORKFLOW|WORKFLOW]]).
> **Yolo herda:** epoch yolo → phases e tasks herdam; phase yolo → tasks herdam. Opt-out/opt-in por task: anexe ` · yolo: não` (ou ` · yolo: sim`) ao fim da célula Descrição — sem coluna nova. O `new-task` resolve a herança e estampa o card (seção Yolo do [[WORKFLOW|WORKFLOW]]).
> **Size:** o agente sugere `S|M|L` na Descrição; `new-task` estampa no card e o humano corrige em 001. Size orienta tier/esforço; risco, skills, dependências e write sets determinam a topologia no [[WORKFLOW|WORKFLOW]].

## Recon e forks

> Pesquisas em `pop/researches/` que embasaram o detalhamento; o que ficou sem resposta é RECON NEEDED, com o check que resolve. Forks: mudanças de rota pré-identificadas.

- Fork: se `OnUnitActiveSec` não bastar para o agendamento → gerar `OnCalendar=*:0/N` quando o intervalo dividir 60.
- Instalação, habilitação e teste real do serviço em servidor remoto ficam no gate humano — a epoch só entrega os artefatos.

## Phase 2.1 — TUI de gestão do backups.json

- **Status:** concluída
- **Descrição:** TUI em Go com CHARM (lista bubbles + formulários huh) para listar, adicionar, editar e remover entradas do `backups.json`.

| Task | Descrição (≤1 linha) | Status |
|------|----------------------|--------|

> **Toda phase termina com a task `verificacao-da-phase`** (`depends_on` todas as demais): é a única em que testes rodam — seção "Verificação de phase" do [[WORKFLOW|WORKFLOW]].

## Phase 2.2 — Serviço systemd

- **Status:** concluída
- **Descrição:** Script que gera unit + timer systemd a partir do `BACKUP_INTERVAL_MINUTES` e oferece instalar/habilitar/status/remover.

| Task | Descrição (≤1 linha) | Status |
|------|----------------------|--------|

> **Toda phase termina com a task `verificacao-da-phase`** (`depends_on` todas as demais): é a única em que testes rodam — seção "Verificação de phase" do [[WORKFLOW|WORKFLOW]].
