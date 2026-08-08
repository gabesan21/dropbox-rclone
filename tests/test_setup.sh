#!/usr/bin/env bash
set -euo pipefail

# test_setup.sh — suíte de testes da Phase 1.2 para setup-rclone.sh, .env e backups.json

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
SETUP_SCRIPT="$ROOT_DIR/setup-rclone.sh"
ENV_FILE="$ROOT_DIR/.env"
JSON_FILE="$ROOT_DIR/backups.json"

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

# --- Teste 1: sintaxe do setup-rclone.sh --------------------------------------

test_sintaxe_setup() {
    if bash -n "$SETUP_SCRIPT" 2>/dev/null; then
        pass "sintaxe do setup-rclone.sh é válida"
    else
        fail "sintaxe do setup-rclone.sh é inválida"
    fi
}

# --- Teste 2: .env ------------------------------------------------------------

test_env_existe() {
    if [[ -f "$ENV_FILE" ]]; then
        pass ".env existe"
    else
        fail ".env não existe"
    fi
}

test_env_intervalo() {
    if grep -q 'BACKUP_INTERVAL_MINUTES' "$ENV_FILE"; then
        pass ".env contém BACKUP_INTERVAL_MINUTES"
    else
        fail ".env não contém BACKUP_INTERVAL_MINUTES"
    fi
}

# --- Teste 3: backups.json -----------------------------------------------------

test_json_existe() {
    if [[ -f "$JSON_FILE" ]]; then
        pass "backups.json existe"
    else
        fail "backups.json não existe"
    fi
}

test_json_valido() {
    if python3 -m json.tool "$JSON_FILE" >/dev/null 2>&1; then
        pass "backups.json é JSON válido"
    else
        fail "backups.json não é JSON válido"
    fi
}

test_json_campos() {
    local campos=("path" "rclone_account" "remote_path" "backup_time" "max_backups" "type")
    local ok=true
    for campo in "${campos[@]}"; do
        if ! grep -q "\"$campo\"" "$JSON_FILE"; then
            ok=false
            break
        fi
    done
    if [[ "$ok" == true ]]; then
        pass "backups.json tem todos os campos obrigatórios"
    else
        fail "backups.json está faltando campo obrigatório"
    fi
}

test_json_tipos() {
    if grep -q '"compacted"' "$JSON_FILE" || \
       grep -q '"folder-backup"' "$JSON_FILE" || \
       grep -q '"folder-sync"' "$JSON_FILE"; then
        pass "backups.json contém tipo de backup válido"
    else
        fail "backups.json não contém tipo de backup válido"
    fi
}

# --- Execução -----------------------------------------------------------------

main() {
    if [[ ! -f "$SETUP_SCRIPT" ]]; then
        printf 'ERRO: setup-rclone.sh não encontrado em %s\n' "$SETUP_SCRIPT" >&2
        exit 1
    fi

    test_sintaxe_setup
    test_env_existe
    test_env_intervalo
    test_json_existe
    test_json_valido
    test_json_campos
    test_json_tipos

    printf '\nResultado: %d pass, %d fail\n' "$PASS" "$FAIL"
    [[ "$FAIL" -eq 0 ]]
}

main "$@"
