#!/usr/bin/env bash
set -euo pipefail

# test_install.sh — suíte de testes da Phase 1.1 para install.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
INSTALL_SCRIPT="$ROOT_DIR/install.sh"

PASS=0
FAIL=0

pass() {
    printf 'PASS: %s\n' "$1"
    PASS=$((PASS+1))
}

fail() {
    printf 'FAIL: %s\n' "$1"
    FAIL=$((FAIL+1))
}

# --- Teste 1: sintaxe ---------------------------------------------------------

test_sintaxe() {
    if bash -n "$INSTALL_SCRIPT" 2>/dev/null; then
        pass "sintaxe do install.sh é válida"
    else
        fail "sintaxe do install.sh é inválida"
    fi
}

# --- Teste 2: detecção de distro ----------------------------------------------

test_deteccao_debian() {
    local tmpdir
    tmpdir=$(mktemp -d)
    cat > "$tmpdir/os-release" <<'EOF'
PRETTY_NAME="Ubuntu 24.04 LTS"
ID=ubuntu
ID_LIKE=debian
EOF
    if grep -q 'is_debian_like' "$INSTALL_SCRIPT" && \
       grep -q 'debian|ubuntu|linuxmint' "$INSTALL_SCRIPT"; then
        pass "detecção Debian/Ubuntu/Mint presente"
    else
        fail "detecção Debian/Ubuntu/Mint ausente"
    fi
    rm -rf "$tmpdir"
}

test_deteccao_arch() {
    if grep -q 'is_arch_like' "$INSTALL_SCRIPT" && \
       grep -q 'arch|manjaro|endeavouros' "$INSTALL_SCRIPT"; then
        pass "detecção Arch presente"
    else
        fail "detecção Arch ausente"
    fi
}

test_deteccao_fedora() {
    if grep -q 'is_fedora_like' "$INSTALL_SCRIPT" && \
       grep -q 'fedora|rhel|centos' "$INSTALL_SCRIPT"; then
        pass "detecção Fedora presente"
    else
        fail "detecção Fedora ausente"
    fi
}

# --- Teste 3: pré-requisitos --------------------------------------------------

test_prerequisitos() {
    if grep -q 'command -v rclone' "$INSTALL_SCRIPT" && \
       grep -q 'command -v go' "$INSTALL_SCRIPT" && \
       grep -q 'need_rclone=false' "$INSTALL_SCRIPT" && \
       grep -q 'need_go=false' "$INSTALL_SCRIPT"; then
        pass "verificação de pré-requisitos presente"
    else
        fail "verificação de pré-requisitos ausente"
    fi
}

test_package_managers() {
    local ok=true
    grep -q 'apt-get install' "$INSTALL_SCRIPT" || ok=false
    grep -q 'pacman -S' "$INSTALL_SCRIPT" || ok=false
    grep -q 'dnf install' "$INSTALL_SCRIPT" || ok=false
    if [[ "$ok" == true ]]; then
        pass "package managers apt/pacman/dnf presentes"
    else
        fail "package managers ausentes"
    fi
}

# --- Execução -----------------------------------------------------------------

main() {
    if [[ ! -f "$INSTALL_SCRIPT" ]]; then
        printf 'ERRO: install.sh não encontrado em %s\n' "$INSTALL_SCRIPT" >&2
        exit 1
    fi

    test_sintaxe
    test_deteccao_debian
    test_deteccao_arch
    test_deteccao_fedora
    test_prerequisitos
    test_package_managers

    printf '\nResultado: %d pass, %d fail\n' "$PASS" "$FAIL"
    [[ "$FAIL" -eq 0 ]]
}

main "$@"
