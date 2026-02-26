#!/usr/bin/env sh
# install.sh — Install unispeedtest from GitHub Releases
set -eu

REPO="hsblabs/universal-speedtest-cli"
BINARY="unispeedtest"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# ---------------------------------------------------------------------------
# Detect OS and architecture
# ---------------------------------------------------------------------------
detect_os() {
  case "$(uname -s)" in
    Linux)  echo "linux" ;;
    Darwin) echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) echo "Unsupported OS: $(uname -s)" >&2; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
  esac
}

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Required command not found: $1" >&2
    exit 1
  fi
}

download() {
  local url="$1"
  local dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$dest"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$dest" "$url"
  else
    echo "Neither curl nor wget found. Please install one of them." >&2
    exit 1
  fi
}

sha256_file() {
  local file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{print $1}'
  else
    echo "Neither sha256sum nor shasum found. Cannot verify checksums." >&2
    exit 1
  fi
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
OS=$(detect_os)
ARCH=$(detect_arch)
need_cmd awk

# Resolve latest version via GitHub API if VERSION is not set
if [ -z "${VERSION:-}" ]; then
  need_cmd grep
  need_cmd sed
  API_URL="https://api.github.com/repos/${REPO}/releases/latest"
  TMP_JSON=$(mktemp)
  download "$API_URL" "$TMP_JSON"
  VERSION=$(grep '"tag_name"' "$TMP_JSON" | sed -E 's/.*"([^"]+)".*/\1/')
  rm -f "$TMP_JSON"
fi

if [ -z "$VERSION" ]; then
  echo "Could not determine the latest version. Set VERSION env var manually." >&2
  exit 1
fi

# Strip leading 'v' for archive name
VERSION_NUM="${VERSION#v}"

if [ "$OS" = "windows" ]; then
  ARCHIVE="${BINARY}_${VERSION_NUM}_${OS}_${ARCH}.zip"
else
  ARCHIVE="${BINARY}_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

echo "Installing ${BINARY} ${VERSION} (${OS}/${ARCH})..."
echo "Downloading: ${DOWNLOAD_URL}"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

download "$DOWNLOAD_URL" "${TMP_DIR}/${ARCHIVE}"
download "$CHECKSUMS_URL" "${TMP_DIR}/checksums.txt"

EXPECTED_SHA=$(grep "  ${ARCHIVE}\$" "${TMP_DIR}/checksums.txt" | awk '{print $1}' | head -n1)
if [ -z "${EXPECTED_SHA}" ]; then
  echo "Could not find checksum for ${ARCHIVE} in checksums.txt" >&2
  exit 1
fi

ACTUAL_SHA=$(sha256_file "${TMP_DIR}/${ARCHIVE}")
if [ "${ACTUAL_SHA}" != "${EXPECTED_SHA}" ]; then
  echo "Checksum mismatch for ${ARCHIVE}" >&2
  echo "Expected: ${EXPECTED_SHA}" >&2
  echo "Actual:   ${ACTUAL_SHA}" >&2
  exit 1
fi

# Extract archive
cd "$TMP_DIR"
if [ "$OS" = "windows" ]; then
  need_cmd unzip
  unzip -q "${ARCHIVE}"
else
  need_cmd tar
  tar xzf "${ARCHIVE}"
fi

# Install binary
BINARY_SRC="${TMP_DIR}/${BINARY}"
if [ "$OS" = "windows" ]; then
  BINARY_SRC="${TMP_DIR}/${BINARY}.exe"
fi

if [ ! -f "$BINARY_SRC" ]; then
  echo "Binary not found in archive." >&2
  exit 1
fi

chmod +x "$BINARY_SRC"

if [ -w "$INSTALL_DIR" ]; then
  mv "$BINARY_SRC" "${INSTALL_DIR}/${BINARY}"
else
  echo "Installing to ${INSTALL_DIR} requires sudo..."
  sudo mv "$BINARY_SRC" "${INSTALL_DIR}/${BINARY}"
fi

echo ""
echo "${BINARY} ${VERSION} installed to ${INSTALL_DIR}/${BINARY}"
echo "Run '${BINARY} --help' to get started."
