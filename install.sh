#!/usr/bin/env bash
set -euo pipefail

# install.sh — instala rclone e Go em servidores Debian/Ubuntu/Mint, Arch ou Fedora-based.
# Em Debian/Ubuntu/Mint o Go vem do PPA longsleep/golang-backports — o golang-go
# do apt é antigo demais para o go.mod (ex.: Ubuntu 22.04 entrega Go 1.18).
# Uso: ./install.sh

log() {
    printf '[install] %s\n' "$1"
}

error() {
    printf '[install] ERRO: %s\n' "$1" >&2
    exit 1
}

# --- Detecção de distro -------------------------------------------------------

if [[ ! -f /etc/os-release ]]; then
    error "Não foi possível detectar a distro: /etc/os-release não encontrado."
fi

# shellcheck source=/dev/null
source /etc/os-release

DISTRO_ID="${ID:-}"
DISTRO_LIKE="${ID_LIKE:-}"

log "Distro detectada: ${PRETTY_NAME:-$DISTRO_ID}"

is_debian_like() {
    [[ "$DISTRO_ID" =~ ^(debian|ubuntu|linuxmint)$ ]] || \
    [[ "$DISTRO_LIKE" =~ debian|ubuntu ]]
}

is_arch_like() {
    [[ "$DISTRO_ID" =~ ^(arch|manjaro|endeavouros)$ ]] || \
    [[ "$DISTRO_LIKE" =~ arch ]]
}

is_fedora_like() {
    [[ "$DISTRO_ID" =~ ^(fedora|rhel|centos|rocky|alma)$ ]] || \
    [[ "$DISTRO_LIKE" =~ fedora|rhel ]]
}

# --- Verificação de pré-requisitos -------------------------------------------

need_rclone=true
need_go=true

if command -v rclone &>/dev/null; then
    log "rclone já instalado: $(rclone version | head -n1)"
    need_rclone=false
fi

if command -v go &>/dev/null; then
    log "Go já instalado: $(go version)"
    need_go=false
fi

if [[ "$need_rclone" == false && "$need_go" == false ]]; then
    log "Nada a fazer — rclone e Go já estão instalados."
    exit 0
fi

# --- Instalação por família de distro -----------------------------------------

install_debian() {
    log "Usando apt (Debian/Ubuntu/Mint)..."
    if [[ "$need_go" == true ]]; then
        # Go sempre via PPA longsleep/golang-backports: o golang-go do apt das
        # distros LTS é antigo demais para o go.mod (ex.: Go 1.18 no Ubuntu 22.04).
        log "Adicionando PPA longsleep/golang-backports para o Go..."
        sudo apt-get update -qq
        sudo apt-get install -y software-properties-common
        sudo add-apt-repository -y ppa:longsleep/golang-backports
    fi
    sudo apt-get update -qq
    local pkgs=()
    [[ "$need_rclone" == true ]] && pkgs+=(rclone)
    [[ "$need_go" == true ]] && pkgs+=(golang-go)
    sudo apt-get install -y "${pkgs[@]}"
}

install_arch() {
    log "Usando pacman (Arch)..."
    local pkgs=()
    [[ "$need_rclone" == true ]] && pkgs+=(rclone)
    [[ "$need_go" == true ]] && pkgs+=(go)
    sudo pacman -S --needed --noconfirm "${pkgs[@]}"
}

install_fedora() {
    log "Usando dnf (Fedora)..."
    local pkgs=()
    [[ "$need_rclone" == true ]] && pkgs+=(rclone)
    [[ "$need_go" == true ]] && pkgs+=(golang)
    sudo dnf install -y "${pkgs[@]}"
}

if is_debian_like; then
    install_debian
elif is_arch_like; then
    install_arch
elif is_fedora_like; then
    install_fedora
else
    error "Distro não suportada: ${PRETTY_NAME:-$DISTRO_ID}. Suportadas: Debian/Ubuntu/Mint, Arch, Fedora-based."
fi

# --- Verificação pós-instalação ----------------------------------------------

log "Verificando instalação..."

if [[ "$need_rclone" == true ]]; then
    command -v rclone &>/dev/null || error "rclone não encontrado após instalação."
    log "rclone: $(rclone version | head -n1)"
fi

if [[ "$need_go" == true ]]; then
    command -v go &>/dev/null || error "Go não encontrado após instalação."
    log "Go: $(go version)"
fi

log "Instalação concluída com sucesso."
