#!/usr/bin/env bash
set -euo pipefail

# setup-rclone.sh — verifica instalação do rclone, orienta configuração do Dropbox e valida conexão.
# Uso: ./setup-rclone.sh [nome-do-remote]

REMOTE_NAME="${1:-dropbox}"

log() {
    printf '[setup-rclone] %s\n' "$1"
}

error() {
    printf '[setup-rclone] ERRO: %s\n' "$1" >&2
    exit 1
}

# --- Verificação de instalação ------------------------------------------------

if ! command -v rclone &>/dev/null; then
    error "rclone não encontrado. Execute ./install.sh primeiro."
fi

log "rclone encontrado: $(rclone version | head -n1)"

# --- Listagem de remotes ------------------------------------------------------

REMOTES=$(rclone listremotes 2>/dev/null || true)

if [[ -z "$REMOTES" ]]; then
    log "Nenhum remote configurado ainda."
    log ""
    log "Para configurar o Dropbox, execute:"
    log "  rclone config"
    log ""
    log "Passos no rclone config:"
    log "  1. n  (novo remote)"
    log "  2. nome: ${REMOTE_NAME}"
    log "  3. tipo: dropbox"
    log "  4. siga as instruções de OAuth (abrirá browser ou peça token)"
    log ""
    log "Em servidor headless (sem browser), configure localmente e copie:"
    log "  ~/.config/rclone/rclone.conf"
    log ""
    exit 0
fi

log "Remotes configurados:"
printf '%s\n' "$REMOTES" | sed 's/^/  - /'

# --- Busca por remote Dropbox -------------------------------------------------

DROPBOX_REMOTE=""
while IFS= read -r line; do
    remote="${line%:}"
    if rclone config show "$remote" 2>/dev/null | grep -q 'type = dropbox'; then
        DROPBOX_REMOTE="$remote"
        break
    fi
done <<< "$REMOTES"

if [[ -z "$DROPBOX_REMOTE" ]]; then
    log ""
    log "Nenhum remote do tipo Dropbox encontrado."
    log "Execute 'rclone config' e crie um remote do tipo 'dropbox'."
    exit 0
fi

log ""
log "Remote Dropbox encontrado: ${DROPBOX_REMOTE}"

# --- Validação de conexão -----------------------------------------------------

log "Validando conexão com ${DROPBOX_REMOTE}..."

if rclone lsd "${DROPBOX_REMOTE}:" --max-depth 1 &>/dev/null; then
    log "Conexão OK — remote ${DROPBOX_REMOTE} acessível."
else
    error "Falha ao acessar ${DROPBOX_REMOTE}:. Verifique o token em ~/.config/rclone/rclone.conf"
fi
