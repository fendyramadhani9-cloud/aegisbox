#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# AegisBox — Host Installation & Environment Provisioning Script
# ==============================================================================

INSTALL_DIR="/opt/aegisbox"
BIN_DIR="/usr/local/bin"

echo "=================================================="
echo "  AegisBox System Installation"
echo "=================================================="

if [ "$(id -u)" -ne 0 ]; then
    echo "Error: Root privileges required for system installation. Please run with sudo."
    exit 1
fi

# 1. Create directory structure
mkdir -p "${INSTALL_DIR}"/{bin,config,releases,rootfs,workspaces}
mkdir -p /sys/fs/cgroup/aegisbox

# 2. Build or copy current release
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"

# 3. Setup Python RootFS if not already present
if [ ! -d "${INSTALL_DIR}/rootfs/python" ] && [ -f "${PROJECT_ROOT}/scripts/setup-rootfs.sh" ]; then
    echo "==> Building Python RootFS template..."
    "${PROJECT_ROOT}/scripts/setup-rootfs.sh" "${INSTALL_DIR}/rootfs/python"
fi

# 4. Install systemd service unit
if [ -f "${PROJECT_ROOT}/deploy/aegisbox.service" ]; then
    cp -p "${PROJECT_ROOT}/deploy/aegisbox.service" /etc/systemd/system/aegisbox.service
    systemctl daemon-reload
    systemctl enable aegisbox.service
    echo "==> Systemd service registered and enabled (/etc/systemd/system/aegisbox.service)"
fi

# 5. Link global CLI executable
ln -sf "${INSTALL_DIR}/bin/aegisbox" "${BIN_DIR}/aegisbox"

# 6. Execute deployment of the initial release
echo "==> Performing initial release deployment..."
cd "${PROJECT_ROOT}"
"${PROJECT_ROOT}/scripts/deploy.sh"

echo "=================================================="
echo "  AegisBox Host Installation Completed!"
echo "=================================================="
