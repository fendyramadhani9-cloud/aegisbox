#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# AegisBox — Automated CD Deployment Script
# ==============================================================================

SERVER_HOST="${1:-127.0.0.1}"
SERVER_PORT="${2:-8080}"

echo "==> Deploying AegisBox Service..."

# Ensure systemd service is reloaded and running
systemctl daemon-reload
systemctl restart aegisbox.service

echo "==> Waiting for service to become healthy..."
MAX_RETRIES=10
RETRY_COUNT=0

until curl -s "http://${SERVER_HOST}:${SERVER_PORT}/health" | grep -q '"status": "ok"'; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ "${RETRY_COUNT}" -ge "${MAX_RETRIES}" ]; then
        echo "Error: AegisBox health check failed after ${MAX_RETRIES} attempts."
        journalctl -u aegisbox.service -n 20 --no-pager
        exit 1
    fi
    sleep 1
done

echo "==> AegisBox service is healthy and operational on port ${SERVER_PORT}!"
