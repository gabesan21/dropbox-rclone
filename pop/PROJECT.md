# dropbox-rclone

- **Categoria:** work
- **Status:** planejando
- **Prioridade:** alta
- **Criado em:** 2026-08-07
- **Roadmap:** [[pop/ROADMAP|Roadmap]]

## Objetivo

Ferramenta de backup automatizado para servidores remotos: instala rclone e Go, configura Dropbox, agenda backups via JSON com três tipos (`compacted`, `folder-backup`, `folder-sync`) e executa como serviço periódico, com TUI de gestão em Go.

## Contexto

Backup manual de servidores é propenso a esquecimento e erro. Este projeto entrega um processo automatizado e configurável: o administrador instala, configura as pastas no JSON e o serviço cuida do resto. Foco em servidores remotos acessados via SSH — a execução real da instalação e do primeiro teste é gate humano.

## Estrutura de pastas

Anatomia padrão (ver AGENTS.md da raiz): `AGENTS.md` do projeto + `.agents/skills/` na raiz; **todo o harness em `pop/`** — `pop/PROJECT.md` + `pop/ROADMAP.md` + `pop/roadmap/` (epochs), `pop/researches/` (pesquisas por assunto), `pop/skills/`, `pop/specs/`, `pop/notes/` (learnings/decisions/ideas/references), `pop/memory/` (resumos de tasks concluídas), `pop/worktrees/` (gitignorada), `pop/kanban/` (estágios 001–005_closing do [[WORKFLOW|WORKFLOW]]); o **conteúdo do projeto** (código Go, shell scripts, configs) vive direto na raiz.

## Harness do agente

- **Type e repositórios:** type `included` — a raiz do projeto é o próprio repo `https://github.com/gabesan21/dropbox-rclone.git`, branch de PR `main`.
- **Worktree por task:** sim (padrão).
- **Ferramentas e restrições:** Go, rclone, CHARM (Bubble Tea/Lip Gloss/Bubbles/Huh), systemd, shell script. Distros alvo: Debian/Ubuntu/Mint, Arch, Fedora-based. Execução em servidor remoto via SSH exige gate humano para instalação, configuração e teste real.
- **Tom/estilo:** código limpo, comentários em pt-BR, logs claros para diagnóstico remoto.
- **Tasks críticas por padrão?** sim — tasks que tocam instalação, serviço systemd ou configuração de rclone em produção exigem gate humano extra em 005.
- **Skills:** nenhuma skill de domínio ainda.

## Projetos relacionados

_Nenhum._

## Decisões

- **2026-08-07:** Projeto criado como type `included` com repo externo. Instalação, configuração e teste em servidor remoto ficam fora do kanban — são gate humano, não tasks de código.
