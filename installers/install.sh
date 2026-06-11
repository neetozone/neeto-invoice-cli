#!/bin/sh
set -e

BASE_URL="https://neeto-downloads.s3.amazonaws.com/cli/NeetoInvoice/latest"
INSTALL_DIR="/usr/local/bin"

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "macos" ;;
    *)       echo "Unsupported OS: $(uname -s)" >&2; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)             echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
  esac
}

OS=$(detect_os)
ARCH=$(detect_arch)
ARCHIVE="neetoinvoice_${OS}_${ARCH}.tar.gz"
URL="${BASE_URL}/${ARCHIVE}"

echo "Downloading NeetoInvoice CLI for ${OS}/${ARCH}..."
TMPDIR=$(mktemp -d)
curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}"

echo "Extracting..."
tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

echo "Installing to ${INSTALL_DIR}..."
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMPDIR}/neetoinvoice" "${INSTALL_DIR}/neetoinvoice"
else
  sudo mv "${TMPDIR}/neetoinvoice" "${INSTALL_DIR}/neetoinvoice"
fi

rm -rf "$TMPDIR"

echo "NeetoInvoice CLI installed successfully. Run 'neetoinvoice --help' to get started."
