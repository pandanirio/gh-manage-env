#!/usr/bin/env bash
set -euo pipefail

REPO="pandanirio/gh-manage-env"
BIN="gh-manage-env"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

VERSION="${1:-latest}"

if [ "$VERSION" = "latest" ]; then
  URL="https://github.com/${REPO}/releases/latest/download/${BIN}_${OS}_${ARCH}"
else
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${BIN}_${OS}_${ARCH}"
fi

DEST="${DEST:-/usr/local/bin/${BIN}}"
TMP="$(mktemp)"

echo "Downloading $URL"
curl -fsSL "$URL" -o "$TMP"
chmod +x "$TMP"

echo "Installing to $DEST"
sudo mv "$TMP" "$DEST"

echo "Installed: $DEST"
"$BIN" --help >/dev/null || true
