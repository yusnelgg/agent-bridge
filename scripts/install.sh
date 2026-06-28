#!/bin/sh
set -eu

REPO="yusnelgg/agent-bridge"
VERSION="${AGENT_VERSION:-latest}"

# ── Detect platform ──
detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$OS" in
    linux)   OS="linux" ;;
    darwin)  OS="darwin" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *) echo "❌ OS no soportado: $OS"; exit 1 ;;
  esac

  case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "❌ Arquitectura no soportada: $ARCH"; exit 1 ;;
  esac

  if [ "$OS" = "darwin" ] && [ "$ARCH" = "amd64" ]; then
    echo "⚠️  macOS Intel detectado. Usando amd64."
  fi
}

# ── Get latest version from GitHub ──
get_latest_version() {
  if [ "$VERSION" = "latest" ]; then
    echo "🔍 Detectando última versión..." >&2
    VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | \
      grep '"tag_name"' | cut -d'"' -f4)
    if [ -z "$VERSION" ]; then
      echo "❌ No se pudo detectar la última versión. Usá AGENT_VERSION=vX.Y.Z" >&2
      exit 1
    fi
    echo "   Última versión: $VERSION" >&2
  fi
}

# ── Download and install ──
download_and_install() {
  BASENAME="agent-bridge-$OS"
  if [ "$OS" = "windows" ]; then
    BASENAME="agent-bridge-windows"
  fi
  URL="https://github.com/$REPO/releases/download/$VERSION/$BASENAME.zip"

  TMP_DIR=$(mktemp -d)
  ZIP_FILE="$TMP_DIR/agent-bridge.zip"

  echo "📥 Descargando $URL ..." >&2
  curl -fsSL "$URL" -o "$ZIP_FILE"

  echo "📦 Extrayendo..." >&2
  unzip -q "$ZIP_FILE" -d "$TMP_DIR/extracted"

  if [ "$OS" = "windows" ]; then
    echo "📋 Ejecutando install.bat..." >&2
    cmd.exe /c "$TMP_DIR/extracted/install.bat" 2>/dev/null || \
      powershell -Command "Start-Process -Wait -FilePath '$TMP_DIR/extracted/install.bat'"
  else
    echo "📋 Ejecutando install.sh..." >&2
    chmod +x "$TMP_DIR/extracted/install.sh"
    sh "$TMP_DIR/extracted/install.sh"
  fi

  rm -rf "$TMP_DIR"
}

# ── Main ──
echo "=============================="
echo "  Agent Bridge — Instalador"
echo "=============================="
echo ""

detect_platform
get_latest_version
download_and_install

echo ""
echo "✅ Instalación completada."
echo ""
echo "   Frontend:  agent-bridge -config ~/.agent-bridge/frontend.yaml"
echo "   Backend:   AGENT_BRIDGE=http://localhost:9091 agent-bridge -config ~/.agent-bridge/backend.yaml"
echo ""
