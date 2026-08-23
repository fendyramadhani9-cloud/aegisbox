#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# AegisBox — Host Installation Script
# ==============================================================================

INSTALL_DIR="/opt/aegisbox"
BIN_DIR="/usr/local/bin"

echo "==> Installing AegisBox to ${INSTALL_DIR}..."

if [ "$(id -u)" -ne 0 ]; then
    echo "Error: Root privileges required for system installation. Please run with sudo."
    exit 1
fi

mkdir -p "${INSTALL_DIR}"/{bin,config,rootfs,workspaces}
mkdir -p /sys/fs/cgroup/aegisbox

# Copy binary
if [ -f "./bin/aegisbox" ]; then
    cp -p "./bin/aegisbox" "${INSTALL_DIR}/bin/aegisbox"
    ln -sf "${INSTALL_DIR}/bin/aegisbox" "${BIN_DIR}/aegisbox"
    echo "==> AegisBox binary installed to ${BIN_DIR}/aegisbox"
else
    echo "Error: ./bin/aegisbox not found. Run 'go build -o bin/aegisbox ./cmd/aegisbox' first."
    exit 1
fi

# Copy default config
if [ -f "./configs/config.yaml" ]; then
    cp -n "./configs/config.yaml" "${INSTALL_DIR}/config/config.yaml" || true
fi

# Generate Python RootFS
if [ -x "./scripts/setup-rootfs.sh" ]; then
    ./scripts/setup-rootfs.sh "${INSTALL_DIR}/rootfs/python"
fi

# Install systemd service if present
if [ -f "./deploy/aegisbox.service" ]; then
    cp -p "./deploy/aegisbox.service" /etc/systemd/system/aegisbox.service
    systemctl daemon-reload
    echo "==> Systemd service installed (/etc/systemd/system/aegisbox.service)"
fi

echo "==> AegisBox installation completed successfully."
