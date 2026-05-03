#!/usr/bin/env sh
set -e

REPO="nullapt/nullapt"
BINARY="nullapt"
INSTALL_DIR="/usr/local/bin"

# ── Detect OS ─────────────────────────────────────────────────────────────────
case "$(uname -s)" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)
    echo "Unsupported OS: $(uname -s)"
    echo "Download manually from https://github.com/${REPO}/releases"
    exit 1
    ;;
esac

# ── Detect architecture ───────────────────────────────────────────────────────
case "$(uname -m)" in
  x86_64 | amd64)  ARCH="amd64" ;;
  arm64  | aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $(uname -m)"
    echo "Download manually from https://github.com/${REPO}/releases"
    exit 1
    ;;
esac

# ── Resolve latest version ────────────────────────────────────────────────────
echo "Fetching latest release..."
TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$TAG" ]; then
  echo "Could not resolve latest release. Check https://github.com/${REPO}/releases"
  exit 1
fi

# GoReleaser strips the leading 'v' from the archive filename
VERSION="${TAG#v}"
ARCHIVE="${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ARCHIVE}"

# ── Download & extract ────────────────────────────────────────────────────────
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Downloading ${BINARY} ${TAG} (${OS}/${ARCH})..."
curl -fsSL "$URL" -o "${TMP}/${ARCHIVE}"

echo "Extracting..."
tar -xzf "${TMP}/${ARCHIVE}" -C "$TMP"

# ── Install ───────────────────────────────────────────────────────────────────
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

chmod +x "${INSTALL_DIR}/${BINARY}"

echo ""
echo "✓ ${BINARY} ${TAG} installed to ${INSTALL_DIR}/${BINARY}"
echo ""
echo "Get started:"
echo "  ${BINARY} get web-search      # install a skill"
echo "  ${BINARY} list                # list installed skills"
echo "  ${BINARY} --help              # all commands"
