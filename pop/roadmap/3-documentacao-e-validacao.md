# Epoch 3 — Documentação e validação

- **Projeto:** [[pop/PROJECT|dropbox-rclone]] · **Roadmap:** [[pop/ROADMAP|Roadmap]]
- **Status:** em andamento
- **Descrição:** README completo (instalação rclone/Dropbox headless, configuração, TUI, serviço) e validação em servidor remoto via SSH.
- **Yolo:** sim

> Uma phase por seção; sob cada phase, somente suas tasks ainda abertas — **sempre descrições de uma linha**. Detalhe vai para a spec ou para a pasta da task no kanban. Task iniciada ganha link `[[<id>]]`; ao concluir o `005_closing`, sai da tabela depois de sua memory válida (ver [[WORKFLOW|WORKFLOW]]).
> **Yolo herda:** epoch yolo → phases e tasks herdam; phase yolo → tasks herdam. Opt-out/opt-in por task: anexe ` · yolo: não` (ou ` · yolo: sim`) ao fim da célula Descrição — sem coluna nova. O `new-task` resolve a herança e estampa o card (seção Yolo do [[WORKFLOW|WORKFLOW]]).
> **Size:** o agente sugere `S|M|L` na Descrição; `new-task` estampa no card e o humano corrige em 001. Size orienta tier/esforço; risco, skills, dependências e write sets determinam a topologia no [[WORKFLOW|WORKFLOW]].

## Recon e forks

> Pesquisas em `pop/researches/` que embasaram o detalhamento; o que ficou sem resposta é RECON NEEDED, com o check que resolve. Forks: mudanças de rota pré-identificadas.

- [ ] RECON NEEDED (herdado da Epoch 1): fluxo de token rclone Dropbox sem browser — check: o README deve documentar o fork de port-forward SSH (`ssh -L`) e/ou `rclone authorize` na máquina local; a validação humana confirma qual funciona.
- Instalação, configuração, teste e execução real em servidor remoto são gate humano — a epoch entrega README e checklist; o humano executa no final (decisão registrada: "testes de gate humano ficam para o final de tudo").

## Phase 3.1 — README completo

- **Status:** pendente
- **Descrição:** README pt-BR cobrindo instalação rclone+Go, configuração Dropbox headless via SSH, `.env`/`backups.json`, TUI `manage` e serviço `service.sh`.

| Task | Descrição (≤1 linha) | Status |
|------|----------------------|--------|
| `3.1.1-readme-completo` | Escreve o README pt-BR: pré-requisitos, install.sh, rclone Dropbox headless (fork SSH port-forward), .env/backups.json, manage, service.sh. · size: M | não iniciada |
| `3.1.2-verificacao-da-phase` | Confere o README contra os artefatos reais (comandos, flags, caminhos) e conserta divergências. · size: S | não iniciada |

> **Toda phase termina com a task `verificacao-da-phase`** (`depends_on` todas as demais): é a única em que testes rodam — seção "Verificação de phase" do [[WORKFLOW|WORKFLOW]].

## Phase 3.2 — Validação em servidor remoto

- **Status:** pendente
- **Descrição:** Checklist de validação escrito pelo agente; execução (instalar, configurar, testar, rodar) é gate humano no final de tudo.

| Task | Descrição (≤1 linha) | Status |
|------|----------------------|--------|
| `3.2.1-checklist-validacao-remota` | Escreve `pop/` checklist de validação em servidor remoto via SSH (instalação → token → backup dos 3 tipos → timer). · size: S | não iniciada |
| `3.2.2-validacao-pelo-humano` | (user) Humano executa o checklist em servidor remoto e reporta; agente corrige o que surgir. · size: M · yolo: não | não iniciada |

> **Toda phase termina com a task `verificacao-da-phase`** (`depends_on` todas as demais): é a única em que testes rodam — seção "Verificação de phase" do [[WORKFLOW|WORKFLOW]].
> Phase 3.2 dispensa `verificacao-da-phase`: a verificação é a própria execução humana do checklist.
