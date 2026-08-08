# Epoch 1 — Núcleo: instalação, configuração e motor de backup

- **Projeto:** [[pop/PROJECT|dropbox-rclone]] · **Roadmap:** [[pop/ROADMAP|Roadmap]]
- **Status:** concluída
- **Descrição:** Script de instalação, configuração rclone/Dropbox, arquivos .env/JSON e motor de agendamento em Go.
- **Yolo:** sim
- **Abandonar/pausar se:** rclone não suportar Dropbox de forma estável nas distros alvo.

> Uma phase por seção; sob cada phase, somente suas tasks ainda abertas — **sempre descrições de uma linha**. Detalhe vai para a spec ou para a pasta da task no kanban. Task iniciada ganha link `[[<id>]]`; ao concluir o `005_closing`, sai da tabela depois de sua memory válida (ver [[WORKFLOW|WORKFLOW]]).
> **Yolo herda:** epoch yolo → phases e tasks herdam; phase yolo → tasks herdam. Opt-out/opt-in por task: anexe ` · yolo: não` (ou ` · yolo: sim`) ao fim da célula Descrição — sem coluna nova. O `new-task` resolve a herança e estampa o card (seção Yolo do [[WORKFLOW|WORKFLOW]]).
> **Size:** o agente sugere `S|M|L` na Descrição; `new-task` estampa no card e o humano corrige em 001. Size orienta tier/esforço; risco, skills, dependências e write sets determinam a topologia no [[WORKFLOW|WORKFLOW]].

## Recon e forks

> Pesquisas em `pop/researches/` (escopo com o harness na própria raiz: sem o prefixo `pop/`) que embasaram o detalhamento; o que ficou sem resposta é RECON NEEDED, com o check que resolve. Forks: mudanças de rota pré-identificadas.

- [ ] RECON NEEDED: formato exato do token de configuração rclone para Dropbox em modo headless — check: pesquisa na documentação oficial rclone sobre `rclone config` e backend Dropbox.
- Fork: se rclone exigir interação web para OAuth → README ganha instrução de configuração manual assistida via SSH port-forward.

## Phase 1.1 — Script de instalação e detecção de distro

- **Status:** concluída
- **Descrição:** Shell script que detecta a distro e instala rclone e Go via package manager correto, pulando o que já existe.

| Task | Descrição (≤1 linha) | Status |
|------|----------------------|--------|

> **Toda phase termina com a task `verificacao-da-phase`** (`depends_on` todas as demais): é a única em que testes rodam — seção "Verificação de phase" do [[WORKFLOW|WORKFLOW]].

## Phase 1.2 — Configuração rclone/Dropbox e arquivos .env/JSON

- **Status:** concluída
- **Descrição:** Configuração do backend Dropbox no rclone e criação dos arquivos de configuração `.env` e `backups.json`.

| Task | Descrição (≤1 linha) | Status |
|------|----------------------|--------|

## Phase 1.3 — Motor de agendamento e execução de backup

- **Status:** concluída
- **Descrição:** Programa em Go que lê o JSON, filtra backups no intervalo atual e executa o tipo de backup correto (`compacted`, `folder-backup` ou `folder-sync`) com política de retenção.

| Task | Descrição (≤1 linha) | Status |
|------|----------------------|--------|
