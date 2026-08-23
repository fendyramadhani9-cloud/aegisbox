#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# AegisBox — Minimal Python 3 RootFS Builder
# Creates a minimal, isolated root filesystem template for Python sandboxing.
# ==============================================================================

ROOTFS_DEST="${1:-/opt/aegisbox/rootfs/python}"

echo "==> Preparing AegisBox Python RootFS at: ${ROOTFS_DEST}"

if [ "$(id -u)" -ne 0 ]; then
    echo "Warning: Running without root. Some file permission adjustments may be restricted."
fi

# Ensure destination directory structure
mkdir -p "${ROOTFS_DEST}"/{bin,usr/bin,usr/lib,usr/lib64,lib,lib64,etc,dev,proc,sys,tmp,workspace}
chmod 1777 "${ROOTFS_DEST}/tmp"
chmod 0755 "${ROOTFS_DEST}/workspace"

PYTHON_BIN="$(which python3 || echo "")"
if [ -z "${PYTHON_BIN}" ]; then
    echo "Error: python3 not found on host system."
    exit 1
fi

echo "==> Found Python binary: ${PYTHON_BIN}"
cp -p "${PYTHON_BIN}" "${ROOTFS_DEST}/usr/bin/python3"
ln -sf /usr/bin/python3 "${ROOTFS_DEST}/bin/python3"
ln -sf /usr/bin/python3 "${ROOTFS_DEST}/bin/python"

# Copy dynamic shared libraries required by Python using ldd
echo "==> Resolving dynamic library dependencies..."
for lib in $(ldd "${PYTHON_BIN}" | grep -o '/lib[^ ]*' || true); do
    if [ -f "$lib" ]; then
        dest_dir="${ROOTFS_DEST}$(dirname "$lib")"
        mkdir -p "$dest_dir"
        cp -p "$lib" "$dest_dir/"
    fi
done

# Copy Python standard library modules
PY_VER="$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')"
PY_LIB_DIR="/usr/lib/python${PY_VER}"
if [ -d "${PY_LIB_DIR}" ]; then
    echo "==> Copying Python standard library from ${PY_LIB_DIR}..."
    mkdir -p "${ROOTFS_DEST}/usr/lib/python${PY_VER}"
    # Copy essential standard library files
    cp -rp "${PY_LIB_DIR}"/* "${ROOTFS_DEST}/usr/lib/python${PY_VER}/" 2>/dev/null || true
fi

# Copy dynamic linker
if [ -f "/lib64/ld-linux-x86-64.so.2" ]; then
    mkdir -p "${ROOTFS_DEST}/lib64"
    cp -p "/lib64/ld-linux-x86-64.so.2" "${ROOTFS_DEST}/lib64/"
fi

# Setup minimal /etc
cat << 'EOF' > "${ROOTFS_DEST}/etc/passwd"
root:x:0:0:root:/root:/bin/sh
aegisbox:x:1000:1000:aegisbox:/workspace:/bin/sh
nobody:x:65534:65534:nobody:/nonexistent:/bin/false
EOF

cat << 'EOF' > "${ROOTFS_DEST}/etc/group"
root:x:0:
aegisbox:x:1000:
nogroup:x:65534:
EOF

echo "==> RootFS generation complete."
echo "==> Destination: ${ROOTFS_DEST}"
