#!/usr/bin/env bash
set -euo pipefail

# test_service.sh — suíte de testes da Phase 2.2 para service.sh (sem systemd real)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
SERVICE_SCRIPT="$ROOT_DIR/service.sh"

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
    if bash -n "$SERVICE_SCRIPT" 2>/dev/null; then
        pass "sintaxe do service.sh é válida"
    else
        fail "sintaxe do service.sh é inválida"
    fi
}

# --- Teste 2: intervalo default sem .env --------------------------------------

test_intervalo_default() {
    local got
    got=$(ENV_FILE=/nonexistent/.env bash -c 'source "$1"; read_interval' _ "$SERVICE_SCRIPT")
    if [[ "$got" == "30" ]]; then
        pass "intervalo default é 30 sem .env"
    else
        fail "intervalo default: esperado 30, veio '$got'"
    fi
}

# --- Teste 3: intervalo lido do .env ------------------------------------------

test_intervalo_do_env() {
    local tmpdir got
    tmpdir=$(mktemp -d)
    printf 'BACKUP_INTERVAL_MINUTES=45\n' > "$tmpdir/.env"
    got=$(ENV_FILE="$tmpdir/.env" bash -c 'source "$1"; read_interval' _ "$SERVICE_SCRIPT")
    rm -rf "$tmpdir"
    if [[ "$got" == "45" ]]; then
        pass "intervalo lido do .env (45)"
    else
        fail "intervalo do .env: esperado 45, veio '$got'"
    fi
}

# --- Teste 4: intervalo inválido falha ----------------------------------------

test_intervalo_invalido() {
    local tmpdir
    tmpdir=$(mktemp -d)
    printf 'BACKUP_INTERVAL_MINUTES=abc\n' > "$tmpdir/.env"
    if ENV_FILE="$tmpdir/.env" bash -c 'source "$1"; read_interval' _ "$SERVICE_SCRIPT" 2>/dev/null; then
        fail "intervalo inválido deveria falhar"
    else
        pass "intervalo inválido retorna erro"
    fi
    rm -rf "$tmpdir"
}

# --- Teste 5: timer com OnUnitActiveSec do intervalo --------------------------

test_timer_intervalo() {
    local tmpdir
    tmpdir=$(mktemp -d)
    TIMER_FILE="$tmpdir/dropbox-rclone.timer" bash -c 'source "$1"; generate_timer 15' _ "$SERVICE_SCRIPT"
    if grep -q '^OnUnitActiveSec=15min$' "$tmpdir/dropbox-rclone.timer" \
        && grep -q '^OnBootSec=2min$' "$tmpdir/dropbox-rclone.timer" \
        && grep -q '^WantedBy=timers.target$' "$tmpdir/dropbox-rclone.timer"; then
        pass "timer gerado com OnUnitActiveSec=15min, OnBootSec e WantedBy"
    else
        fail "timer gerado sem os campos esperados"
    fi
    rm -rf "$tmpdir"
}

# --- Teste 6: service com WorkingDirectory/ExecStart/EnvironmentFile ----------

test_service_campos() {
    local tmpdir
    tmpdir=$(mktemp -d)
    SERVICE_FILE="$tmpdir/dropbox-rclone.service" bash -c 'source "$1"; generate_service' _ "$SERVICE_SCRIPT"
    if grep -q "^WorkingDirectory=$ROOT_DIR$" "$tmpdir/dropbox-rclone.service" \
        && grep -q "^ExecStart=$ROOT_DIR/dropbox-rclone$" "$tmpdir/dropbox-rclone.service" \
        && grep -q "^EnvironmentFile=-$ROOT_DIR/.env$" "$tmpdir/dropbox-rclone.service" \
        && grep -q '^Type=oneshot$' "$tmpdir/dropbox-rclone.service"; then
        pass "service gerado com WorkingDirectory, ExecStart, EnvironmentFile e oneshot"
    else
        fail "service gerado sem os campos esperados"
    fi
    rm -rf "$tmpdir"
}

# --- Teste 7: install em DRY_RUN gera os 2 units sem systemctl real -----------

test_install_dry_run() {
    local tmpdir out
    tmpdir=$(mktemp -d)
    out=$(SYSTEMD_USER_DIR="$tmpdir" DRY_RUN=1 "$SERVICE_SCRIPT" install)
    if [[ -f "$tmpdir/dropbox-rclone.service" && -f "$tmpdir/dropbox-rclone.timer" ]] \
        && grep -q '\[dry-run\] systemctl --user daemon-reload' <<< "$out"; then
        pass "install DRY_RUN gera service+timer e não executa systemctl"
    else
        fail "install DRY_RUN não gerou os units ou chamou systemctl"
    fi
    rm -rf "$tmpdir"
}

# --- Teste 8: uso sem argumento retorna exit 2 --------------------------------

test_uso_sem_argumento() {
    local code=0
    "$SERVICE_SCRIPT" > /dev/null 2>&1 || code=$?
    if [[ "$code" -eq 2 ]]; then
        pass "sem argumento sai com exit 2"
    else
        fail "sem argumento: esperado exit 2, veio $code"
    fi
}

# --- Execução -----------------------------------------------------------------

test_sintaxe
test_intervalo_default
test_intervalo_do_env
test_intervalo_invalido
test_timer_intervalo
test_service_campos
test_install_dry_run
test_uso_sem_argumento

printf '\nResultado: %d pass, %d fail\n' "$PASS" "$FAIL"
[[ "$FAIL" -eq 0 ]]
