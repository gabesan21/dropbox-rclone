# Roadmap — dropbox-rclone

Ficha: [[pop/PROJECT|dropbox-rclone]]

> O que chega **fora do planejamento** (hotfix, ajuste, feature emergente pequena) não entra aqui — vai para o [[pop/MODIFICATIONS|MODIFICATIONS]] ([[_templates/MODIFICATIONS|template]], criado sob demanda). Fronteira no [[AGENTS|AGENTS]].

> Só as **epochs**, uma linha cada. Phases e tasks ficam no arquivo de cada epoch em `pop/roadmap/` ([[_templates/EPOCH|template]]; escopo com o harness na própria raiz: sem o prefixo `pop/`). Nunca detalhe aqui.

| # | Epoch | Descrição (≤1 linha) | Status |
|---|-------|----------------------|--------|
| 1 | [[pop/roadmap/1-nucleo-backup\|Núcleo: instalação, configuração e motor de backup]] | Script de instalação, configuração rclone/Dropbox, arquivos .env/JSON e motor de agendamento em Go. | concluída |
| 2 | [[pop/roadmap/2-interface-e-servico\|Interface e serviço]] | TUI de gestão do JSON e serviço systemd de execução periódica. | concluída |
| 3 | [[pop/roadmap/3-documentacao-e-validacao\|Documentação e validação]] | README completo e validação em servidor remoto. | em andamento |
| 4 | [[pop/roadmap/4-robustez-e-full-folder\|Robustez do compacted e tipo full-folder]] | Compacted sem esgotar memória/disco e novo tipo full-folder com rotação por max_backups. | em andamento |

**Status de epoch/phase:** pendente | em andamento | concluída

## Ideias futuras (sem epoch)

- Suporte a outros backends rclone (S3, Google Drive, OneDrive).
- Notificações de sucesso/falha (e-mail, webhook).
- Métricas de backup (tamanho, duração, histórico).
