#!/usr/bin/env bash
# service.sh — gerencia o serviço de backup como unit + timer de systemd de usuário.
#
# Uso: ./service.sh <install|enable|disable|status|remove>
#
# install  — gera os units em ~/.config/systemd/user/ e recarrega o systemd
# enable   — install + habilita e inicia o timer
# disable  — para e desabilita o timer
# status   — mostra o estado do timer e a próxima execução
# remove   — disable + apaga os units
#
# Variáveis de ambiente:
#   SYSTEMD_USER_DIR — sobrescreve o diretório de units (padrão ~/.config/systemd/user)
#   DRY_RUN=1        — imprime os comandos systemctl em vez de executá-los
set -euo pipefail

APP_NAME="dropbox-rclone"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UNIT_DIR="${SYSTEMD_USER_DIR:-$HOME/.config/systemd/user}"
SERVICE_FILE="$UNIT_DIR/$APP_NAME.service"
TIMER_FILE="$UNIT_DIR/$APP_NAME.timer"
ENV_FILE="$PROJECT_DIR/.env"
BINARY="$PROJECT_DIR/$APP_NAME"

# read_interval lê BACKUP_INTERVAL_MINUTES do .env (default 30) e valida.
read_interval() {
    local interval="30"
    if [[ -f "$ENV_FILE" ]]; then
        local value
        value="$(grep -E '^BACKUP_INTERVAL_MINUTES=' "$ENV_FILE" | tail -1 | cut -d= -f2- | tr -d '[:space:]' || true)"
        if [[ -n "$value" ]]; then
            interval="$value"
        fi
    fi
    if ! [[ "$interval" =~ ^[0-9]+$ ]] || [[ "$interval" -lt 1 ]]; then
        echo "ERRO: BACKUP_INTERVAL_MINUTES inválido: '$interval' (use um inteiro >= 1)" >&2
        return 1
    fi
    echo "$interval"
}

# generate_service escreve a unit de serviço (oneshot que roda o binário).
generate_service() {
    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Backup Dropbox via rclone ($APP_NAME)

[Service]
Type=oneshot
WorkingDirectory=$PROJECT_DIR
EnvironmentFile=-$ENV_FILE
ExecStart=$BINARY
EOF
}

# generate_timer escreve o timer com o intervalo do .env.
generate_timer() {
    local interval="$1"
    cat > "$TIMER_FILE" <<EOF
[Unit]
Description=Timer do backup Dropbox via rclone ($APP_NAME)

[Timer]
OnBootSec=2min
OnUnitActiveSec=${interval}min
Unit=$APP_NAME.service

[Install]
WantedBy=timers.target
EOF
}

# run_systemctl executa (ou ecoa, em DRY_RUN) um comando systemctl --user.
run_systemctl() {
    if [[ "${DRY_RUN:-0}" == "1" ]]; then
        echo "[dry-run] systemctl --user $*"
        return 0
    fi
    systemctl --user "$@"
}

cmd_install() {
    local interval
    interval="$(read_interval)"

    mkdir -p "$UNIT_DIR"
    generate_service
    generate_timer "$interval"
    echo "Units geradas em $UNIT_DIR (intervalo: ${interval}min)."

    if [[ ! -x "$BINARY" ]]; then
        echo "AVISO: binário $BINARY não encontrado — compile com: go build -o $APP_NAME ." >&2
    fi

    run_systemctl daemon-reload
    echo "Install concluído. Habilite com: $0 enable"
}

cmd_enable() {
    cmd_install
    run_systemctl enable --now "$APP_NAME.timer"
    echo "Timer habilitado e iniciado."
}

cmd_disable() {
    run_systemctl disable --now "$APP_NAME.timer"
    echo "Timer desabilitado."
}

cmd_status() {
    run_systemctl status "$APP_NAME.timer" || true
    run_systemctl list-timers "$APP_NAME.timer" || true
}

cmd_remove() {
    if [[ -f "$TIMER_FILE" ]]; then
        run_systemctl disable --now "$APP_NAME.timer" || true
    fi
    rm -f "$SERVICE_FILE" "$TIMER_FILE"
    run_systemctl daemon-reload
    echo "Units removidas de $UNIT_DIR."
}

main() {
    local cmd="${1:-}"
    case "$cmd" in
        install) cmd_install ;;
        enable) cmd_enable ;;
        disable) cmd_disable ;;
        status) cmd_status ;;
        remove) cmd_remove ;;
        *)
            echo "uso: $0 <install|enable|disable|status|remove>" >&2
            exit 2
            ;;
    esac
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
