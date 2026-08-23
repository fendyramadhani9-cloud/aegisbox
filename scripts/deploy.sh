#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# AegisBox — Zero-Downtime Release Deployment & Health Validation Script
# ==============================================================================
# Usage:
#   sudo ./scripts/deploy.sh [PATH_TO_TAR_GZ] [PORT]
# Example:
#   sudo ./scripts/deploy.sh /tmp/aegisbox-linux-amd64.tar.gz 8080
# ==============================================================================

TARBALL="${1:-}"
PORT="${2:-8080}"
BASE_DIR="/opt/aegisbox"
RELEASES_DIR="${BASE_DIR}/releases"
CURRENT_LINK="${RELEASES_DIR}/current"
PREVIOUS_LINK="${RELEASES_DIR}/previous"
BIN_LINK="${BASE_DIR}/bin/aegisbox"
CONFIG_DIR="${BASE_DIR}/config"

echo "=================================================="
echo "  AegisBox Continuous Delivery Deployment"
echo "=================================================="

# 1. Root verification
if [ "$(id -u)" -ne 0 ]; then
    echo "Error: Root privileges required for deployment. Please run with sudo."
    exit 1
fi

mkdir -p "${RELEASES_DIR}" "${BASE_DIR}/bin" "${CONFIG_DIR}" "${BASE_DIR}/workspaces"

# 2. Determine Release ID
TIMESTAMP="$(date -u +'%Y%m%d%H%M%S')"
TMP_EXTRACT="/tmp/aegisbox-extract-${TIMESTAMP}"

if [ -n "${TARBALL}" ] && [ -f "${TARBALL}" ]; then
    echo "==> Extracting release archive: ${TARBALL}"
    mkdir -p "${TMP_EXTRACT}"
    tar -xzf "${TARBALL}" -C "${TMP_EXTRACT}"
else
    echo "==> Deploying from working directory..."
    mkdir -p "${TMP_EXTRACT}"
    cp -r ./* "${TMP_EXTRACT}/"
fi

# Detect Commit SHA or Version from metadata if present
if [ -f "${TMP_EXTRACT}/RELEASE_METADATA" ]; then
    RELEASE_ID="$(grep '^COMMIT_SHA=' "${TMP_EXTRACT}/RELEASE_METADATA" | cut -d'=' -f2 || echo "")"
    if [ -z "${RELEASE_ID}" ]; then
        RELEASE_ID="${TIMESTAMP}"
    fi
else
    RELEASE_ID="${TIMESTAMP}"
fi

TARGET_RELEASE_DIR="${RELEASES_DIR}/${RELEASE_ID}"
echo "==> Target Release ID: ${RELEASE_ID}"
echo "==> Target Directory: ${TARGET_RELEASE_DIR}"

# 3. Create target release directory and copy files
mkdir -p "${TARGET_RELEASE_DIR}"
cp -r "${TMP_EXTRACT}"/* "${TARGET_RELEASE_DIR}/"
rm -rf "${TMP_EXTRACT}"
chmod +x "${TARGET_RELEASE_DIR}/bin/aegisbox" || true
chmod +x "${TARGET_RELEASE_DIR}/scripts/"*.sh 2>/dev/null || true

# 4. Pre-deployment Smoke Test on new binary
echo "==> Validating candidate binary integrity..."
if ! "${TARGET_RELEASE_DIR}/bin/aegisbox" version; then
    echo "Error: Candidate binary failed version check. Aborting deployment without modifying active release."
    rm -rf "${TARGET_RELEASE_DIR}"
    exit 1
fi

# 5. Capture previous active release for rollback capability
PREVIOUS_RELEASE=""
if [ -L "${CURRENT_LINK}" ]; then
    PREVIOUS_RELEASE="$(readlink -f "${CURRENT_LINK}")"
    echo "==> Existing active release identified: ${PREVIOUS_RELEASE}"
fi

# 6. Atomic Symlink Switch
echo "==> Activating new release symlink..."
ln -sfn "${TARGET_RELEASE_DIR}" "${CURRENT_LINK}"
ln -sfn "${CURRENT_LINK}/bin/aegisbox" "${BIN_LINK}"

# Ensure default configuration exists
if [ ! -f "${CONFIG_DIR}/config.yaml" ] && [ -f "${CURRENT_LINK}/configs/config.yaml" ]; then
    cp "${CURRENT_LINK}/configs/config.yaml" "${CONFIG_DIR}/config.yaml"
fi

# 7. Ensure RootFS template exists
if [ ! -d "${BASE_DIR}/rootfs/python" ] && [ -x "${CURRENT_LINK}/scripts/setup-rootfs.sh" ]; then
    echo "==> Initializing Python RootFS..."
    "${CURRENT_LINK}/scripts/setup-rootfs.sh" "${BASE_DIR}/rootfs/python"
fi

# 8. Update Systemd unit if included
if [ -f "${CURRENT_LINK}/deploy/aegisbox.service" ]; then
    cp -p "${CURRENT_LINK}/deploy/aegisbox.service" /etc/systemd/system/aegisbox.service
    systemctl daemon-reload
fi

# 9. Restart AegisBox Service
echo "==> Restarting aegisbox.service..."
systemctl restart aegisbox.service || true

# 10. Health Check Polling & Automated Rollback
echo "==> Polling health check endpoint (http://127.0.0.1:${PORT}/health)..."
MAX_ATTEMPTS=15
ATTEMPT=0
HEALTHY=false

while [ "${ATTEMPT}" -lt "${MAX_ATTEMPTS}" ]; do
    ATTEMPT=$((ATTEMPT + 1))
    if curl -s "http://127.0.0.1:${PORT}/health" | grep -q '"status": "ok"'; then
        HEALTHY=true
        break
    fi
    echo "    [Attempt ${ATTEMPT}/${MAX_ATTEMPTS}] Waiting for service to respond..."
    sleep 1
done

if [ "${HEALTHY}" = "true" ]; then
    echo "=================================================="
    echo "  [SUCCESS] Deployment Verified Healthy!"
    echo "=================================================="
    echo "Active Release: ${TARGET_RELEASE_DIR}"
    if [ -n "${PREVIOUS_RELEASE}" ] && [ "${PREVIOUS_RELEASE}" != "${TARGET_RELEASE_DIR}" ]; then
        ln -sfn "${PREVIOUS_RELEASE}" "${PREVIOUS_LINK}"
    fi

    # Prune old releases keeping last 5
    echo "==> Pruning old releases (retaining 5 most recent)..."
    ls -dt "${RELEASES_DIR}"/*/ 2>/dev/null | tail -n +6 | xargs -r rm -rf

    # Final service version check
    curl -s "http://127.0.0.1:${PORT}/health" || true
    echo ""
    exit 0
else
    echo "=================================================="
    echo "  [FAILURE] Health Check Failed! Initiating Rollback..."
    echo "=================================================="
    journalctl -u aegisbox.service -n 25 --no-pager || true

    if [ -n "${PREVIOUS_RELEASE}" ] && [ -d "${PREVIOUS_RELEASE}" ]; then
        echo "==> Rolling back to previous release: ${PREVIOUS_RELEASE}"
        ln -sfn "${PREVIOUS_RELEASE}" "${CURRENT_LINK}"
        ln -sfn "${CURRENT_LINK}/bin/aegisbox" "${BIN_LINK}"
        systemctl restart aegisbox.service
        sleep 2
        if curl -s "http://127.0.0.1:${PORT}/health" | grep -q '"status": "ok"'; then
            echo "==> [ROLLBACK OK] Service restored to previous release."
        else
            echo "==> [CRITICAL] Previous release also failed health check."
        fi
    else
        echo "==> No prior release available to roll back to."
    fi

    exit 1
fi
