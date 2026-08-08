---
author: agent
created: 2026-08-07
---

# Checklist — validação em servidor remoto (gate humano)

Validação ponta a ponta do dropbox-rclone num servidor real via SSH. Explicações e contexto estão no [[README]] — aqui vai só o roteiro executável. Marque cada caixa; se algo falhar, anote na seção **Resultado** e reporte.

## 1. Preparo

- [ ] Clonar o repo no servidor: `git clone https://github.com/gabesan21/dropbox-rclone.git && cd dropbox-rclone`
- [ ] Pass é: repo clonado e você na pasta do projeto.

## 2. Instalação

- [ ] Rodar `./install.sh`.
- [ ] Pass é: script detecta sua distro, instala (ou confirma) rclone e Go, e termina com "Instalação concluída com sucesso".
- [ ] Compilar o binário: `go build -o dropbox-rclone .`
- [ ] Pass é: `ls -l dropbox-rclone` mostra o executável.

## 3. Token Dropbox (headless)

- [ ] Escolher uma rota do [[README#configurando-o-dropbox-servidor-sem-browser|README]] e rodar `rclone config` (remote `dropbox`, tipo `dropbox`).
- [ ] Pass é: OAuth concluído sem browser no servidor.
- [ ] **Anotar qual rota funcionou:** port-forward `ssh -L 53682` ☐ · `rclone authorize "dropbox"` ☐ (fecha o RECON NEEDED da Epoch 3).

## 4. Conexão

- [ ] Rodar `./setup-rclone.sh`.
- [ ] Pass é: "Conexão OK — remote dropbox acessível".

## 5. Configuração dos backups

- [ ] Criar `.env` com `BACKUP_INTERVAL_MINUTES=30`.
- [ ] Rodar `./dropbox-rclone manage` e cadastrar **3 entradas de teste**, uma de cada tipo, em pastas descartáveis (ex.: `/tmp/bkp-teste-*`):
  - [ ] `compacted` (max_backups=2)
  - [ ] `folder-backup`
  - [ ] `folder-sync`
- [ ] Testar na TUI: adicionar, editar e remover uma entrada extra.
- [ ] Pass é: entradas salvas no `backups.json` (conferir com `cat backups.json`).

## 6. Execução manual

- [ ] Ajustar o `backup_time` das entradas para caírem na janela atual e rodar `./dropbox-rclone`.
- [ ] `compacted` — pass é: `.tar.gz` datado aparece no Dropbox; rodando de novo com max_backups=2, o mais antigo some na 3ª execução.
- [ ] `folder-backup` — pass é: remoto fica cópia idêntica da pasta local (apagar um arquivo local e rodar de novo → some do remoto).
- [ ] `folder-sync` — pass é: após o `rclone bisync <path> dropbox:<remote_path> --resync` inicial, mudanças dos dois lados convergem.

## 7. Serviço

- [ ] Rodar `./service.sh enable` e `sudo loginctl enable-linger "$USER"`.
- [ ] Pass é: `./service.sh status` mostra o timer ativo com próxima execução; `journalctl --user -u dropbox-rclone.service` mostra a 1ª corrida.
- [ ] Deslogar do SSH e voltar após um intervalo: timer disparou sem sessão aberta.

## 8. Limpeza

- [ ] `./service.sh remove` e apagar as pastas/arquivos de teste (locais e no Dropbox).

## Resultado

- Data: ____ · Servidor/distro: ____
- Rota de token que funcionou: ____
- Falhas e observações: ____
