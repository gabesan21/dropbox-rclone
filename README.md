# dropbox-rclone

Backup automatizado de pastas para o **Dropbox** usando [rclone](https://rclone.org), com motor de agendamento escrito em **Go**. Pensado para **servidores remotos acessados via SSH** (sem browser): um timer systemd roda o binário em intervalos fixos e ele executa os backups agendados para aquela janela.

## Como funciona

1. Um **timer systemd de usuário** dispara o binário a cada `BACKUP_INTERVAL_MINUTES` minutos (padrão: 30).
2. O binário lê o `backups.json` e seleciona as entradas cujo `backup_time` cai dentro da janela atual — exemplo: rodando às 1h15 com intervalo de 30 min, executa os agendamentos entre 1h00 e 1h30.
3. Cada entrada é executada conforme seu **tipo** (compactado, espelho ou sincronização bidirecional — ver abaixo).

## Pré-requisitos

- Servidor Linux **Debian/Ubuntu/Mint**, **Arch** ou **Fedora-based**, com `sudo`.
- Uma conta Dropbox.
- Acesso SSH ao servidor.

## Instalação

```bash
git clone https://github.com/gabesan21/dropbox-rclone.git
cd dropbox-rclone

# Instala rclone e Go (detecta a distro e usa apt, pacman ou dnf).
# Se já estiverem instalados, o script apenas confirma e sai.
./install.sh

# Compila o binário
go build -o dropbox-rclone .
```

## Configurando o Dropbox (servidor sem browser)

O rclone precisa de um token OAuth do Dropbox. Em servidor headless há duas rotas:

**Rota A — port-forward via SSH (recomendada):** na sua máquina local, abra um túnel para a porta de callback do rclone e rode a configuração no servidor:

```bash
# na máquina local
ssh -L 53682:localhost:53682 usuario@servidor

# no servidor, dentro da sessão SSH
rclone config
```

**Rota B — autorizar na máquina local e levar o token:** numa máquina com browser, rode `rclone authorize "dropbox"`, copie o token gerado e cole quando o `rclone config` do servidor pedir (responda `n` para "auto config").

No `rclone config`:

1. `n` — novo remote
2. nome: `dropbox` (é o `rclone_account` padrão usado no `backups.json`)
3. tipo: `dropbox`
4. conclua o OAuth por uma das rotas acima

Depois valide a conexão:

```bash
./setup-rclone.sh            # ou ./setup-rclone.sh <nome-do-remote>
```

O script lista os remotes, encontra o do tipo Dropbox e testa o acesso.

## Configuração dos backups

### `.env`

Crie o arquivo `.env` na raiz do projeto (está no `.gitignore`):

```bash
BACKUP_INTERVAL_MINUTES=30
```

Define o tamanho da janela de agendamento e o intervalo do timer systemd.

### `backups.json`

Também gitignorado. Array de objetos, um por backup:

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `path` | string | Pasta local de origem (caminho absoluto). |
| `rclone_account` | string | Nome do remote rclone (ex.: `dropbox`). |
| `remote_path` | string | Pasta de destino dentro do remote (ex.: `backups/servidor/dados`). |
| `backup_time` | string | Horário do agendamento, formato `HH:MM`. |
| `max_backups` | int | Máximo de arquivos mantidos no remoto (rotação; usado pelo tipo `compacted`). |
| `type` | string | `compacted`, `folder-backup` ou `folder-sync`. |

### Tipos de backup

- **`compacted`** — compacta a pasta local num `.tar.gz` datado (`<pasta>-AAAAMMDD-HHMMSS.tar.gz`) e sobe para o `remote_path`. Ao ultrapassar `max_backups` no remoto, remove os arquivos mais antigos.
- **`folder-backup`** — espelha o conteúdo local no remoto com `rclone sync`: a pasta remota fica uma **cópia idêntica** da local, que é a referência (o que não existe mais localmente é removido do remoto).
- **`folder-sync`** — sincronização **bidirecional** com `rclone bisync`, respeitando as duas fontes. **Na primeira execução**, inicialize manualmente:

  ```bash
  rclone bisync /caminho/local dropbox:backups/servidor/dados --resync
  ```

Exemplo de `backups.json`:

```json
[
  {
    "path": "/home/usuario/dados",
    "rclone_account": "dropbox",
    "remote_path": "backups/servidor/dados",
    "backup_time": "02:00",
    "max_backups": 7,
    "type": "compacted"
  },
  {
    "path": "/etc/nginx",
    "rclone_account": "dropbox",
    "remote_path": "backups/servidor/nginx",
    "backup_time": "03:00",
    "max_backups": 3,
    "type": "folder-backup"
  }
]
```

### Gerenciando pelo TUI

Em vez de editar o JSON na mão, use a interface de terminal:

```bash
./dropbox-rclone manage
```

Permite **listar, adicionar, editar e remover** entradas do `backups.json`, com formulários validados (horário `HH:MM`, tipo entre os três suportados etc.).

## Serviço (timer systemd)

O `service.sh` cria e gerencia uma **unit + timer de systemd de usuário**, com o intervalo lido do `BACKUP_INTERVAL_MINUTES` do `.env`:

```bash
./service.sh install   # gera os units em ~/.config/systemd/user/
./service.sh enable    # install + habilita e inicia o timer
./service.sh status    # estado do timer e próxima execução
./service.sh disable   # para e desabilita o timer
./service.sh remove    # disable + apaga os units
```

Para o timer rodar mesmo **sem sessão SSH aberta**, habilite o linger (uma vez, com sudo):

```bash
sudo loginctl enable-linger "$USER"
```

Acompanhe as execuções pelos logs:

```bash
journalctl --user -u dropbox-rclone.service
```

## Teste manual

Para rodar uma verificação imediatamente (sem esperar o timer):

```bash
./dropbox-rclone
```

Ele imprime os backups da janela atual e executa cada um, reportando `OK` ou o erro de cada entrada.
