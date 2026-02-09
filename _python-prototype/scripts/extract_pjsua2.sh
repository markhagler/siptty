#!/usr/bin/env bash
# scripts/extract_pjsua2.sh
# Build the pjsua2 Python bindings in Docker and copy them into the local .venv.
#
# Usage:  ./scripts/extract_pjsua2.sh [path/to/.venv]
#         Default venv path: .venv (relative to repo root)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
IMAGE_TAG="siptty-pjsua2-builder"
DOCKERFILE="${REPO_ROOT}/docker/Dockerfile.pjsua2-builder"
VENV_DIR="${1:-${REPO_ROOT}/.venv}"

# Resolve the site-packages directory inside the venv
SITE_PACKAGES="$(
    "${VENV_DIR}/bin/python" -c \
        'import sysconfig; print(sysconfig.get_path("purelib"))'
)"

echo "==> Repository root : ${REPO_ROOT}"
echo "==> Dockerfile       : ${DOCKERFILE}"
echo "==> Target venv      : ${VENV_DIR}"
echo "==> site-packages    : ${SITE_PACKAGES}"
echo

# ── 1. Build the Docker image ────────────────────────────────────────────────
echo "==> Building Docker image '${IMAGE_TAG}' …"
docker build \
    -f "${DOCKERFILE}" \
    -t "${IMAGE_TAG}" \
    "${REPO_ROOT}"
echo

# ── 2. Extract artifacts from the image ──────────────────────────────────────
CONTAINER_ID=$(docker create "${IMAGE_TAG}")
trap 'docker rm "${CONTAINER_ID}" >/dev/null 2>&1' EXIT

echo "==> Copying _pjsua2.so and pjsua2.py from container …"

# Copy the binding files
docker cp "${CONTAINER_ID}:/output/_pjsua2.so" "${SITE_PACKAGES}/_pjsua2.so"
docker cp "${CONTAINER_ID}:/output/pjsua2.py"  "${SITE_PACKAGES}/pjsua2.py"

# Copy the pjproject shared libraries that _pjsua2.so depends on.
# docker cp doesn't support globs, so we copy the whole directory and filter.
mkdir -p "${VENV_DIR}/lib/pjproject"
TMP_LIBS=$(mktemp -d)
docker cp "${CONTAINER_ID}:/usr/local/lib/" "${TMP_LIBS}/"
cp "${TMP_LIBS}"/lib/lib*.so* "${VENV_DIR}/lib/pjproject/" 2>/dev/null || true
rm -rf "${TMP_LIBS}"

echo
echo "==> Artifacts installed:"
ls -lh "${SITE_PACKAGES}/_pjsua2.so" "${SITE_PACKAGES}/pjsua2.py"
echo
echo "==> PJ shared libs:"
ls -lh "${VENV_DIR}/lib/pjproject/" 2>/dev/null || echo "  (none copied – will rely on system libs)"
echo

# ── 3. Verify the import works ───────────────────────────────────────────────
echo "==> Verifying import …"
export LD_LIBRARY_PATH="${VENV_DIR}/lib/pjproject:${LD_LIBRARY_PATH:-}"
if "${VENV_DIR}/bin/python" -c "import pjsua2; print('SUCCESS: pjsua2 module imported')"; then
    echo
    echo "✔ pjsua2 is ready to use in ${VENV_DIR}"
else
    echo
    echo "✘ Import failed. You may need to install runtime dependencies:"
    echo "    sudo apt-get install libasound2 libopus0 libssl3"
    echo "  and ensure LD_LIBRARY_PATH includes ${VENV_DIR}/lib/pjproject"
    exit 1
fi
