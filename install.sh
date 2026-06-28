#!/bin/bash
set -e
DIR="$(cd "$(dirname "$0")" && pwd)"
BIN_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.agent-bridge"

echo "Installing Agent Bridge..."
sudo cp "$DIR/bin/agent-bridge" "$BIN_DIR/agent-bridge"
sudo chmod +x "$BIN_DIR/agent-bridge"
mkdir -p "$CONFIG_DIR"
cp "$DIR/configs/frontend.yaml" "$CONFIG_DIR/frontend.yaml"
cp "$DIR/configs/backend.yaml" "$CONFIG_DIR/backend.yaml"
cp "$DIR/AGENTS.md" "$CONFIG_DIR/AGENTS.md"
cp -r "$DIR/prompts" "$CONFIG_DIR/"
echo "Done. Binaries in $BIN_DIR, configs in $CONFIG_DIR."
