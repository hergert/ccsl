#!/bin/bash
# Install ccsl (Claude Code StatusLine)
# Usage: curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/install.sh | bash

set -e

REPO_URL="https://github.com/hergert/ccsl.git"
INSTALL_DIR="$HOME/.local/bin"
BINARY_PATH="$INSTALL_DIR/ccsl"
CLAUDE_DIR="$HOME/.claude"
SETTINGS_FILE="$CLAUDE_DIR/settings.json"

info()    { printf '\033[0;34m%s\033[0m\n' "$1"; }
success() { printf '\033[0;32m%s\033[0m\n' "$1"; }
warn()    { printf '\033[0;33m%s\033[0m\n' "$1"; }
error()   { printf '\033[0;31m%s\033[0m\n' "$1"; exit 1; }

info "Installing ccsl..."

# Check dependencies
command -v go &>/dev/null || error "Go required. Install from https://go.dev/dl/"
command -v git &>/dev/null || error "Git required."

# Clone to temp dir
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

info "Downloading..."
git clone --depth 1 --quiet "$REPO_URL" "$TMP_DIR"

# Build
info "Building..."
mkdir -p "$INSTALL_DIR"
(cd "$TMP_DIR" && go build -o "$BINARY_PATH" ./cmd/ccsl)
chmod +x "$BINARY_PATH"
success "Installed $BINARY_PATH"

# Check PATH
[[ ":$PATH:" != *":$INSTALL_DIR:"* ]] && warn "Add to PATH: export PATH=\"\$HOME/.local/bin:\$PATH\""

# Configure Claude Code
mkdir -p "$CLAUDE_DIR"

if [ -f "$SETTINGS_FILE" ]; then
    python3 -c "
import json
with open('$SETTINGS_FILE', 'r') as f: s = json.load(f)
s['statusLine'] = {'type': 'command', 'command': '$BINARY_PATH', 'padding': 0}
with open('$SETTINGS_FILE', 'w') as f: json.dump(s, f, indent=2)
" 2>/dev/null || warn "Could not update settings.json"
else
    cat > "$SETTINGS_FILE" << EOF
{
  "statusLine": {
    "type": "command",
    "command": "$BINARY_PATH",
    "padding": 0
  }
}
EOF
fi

# Verify
"$BINARY_PATH" doctor &>/dev/null && success "Verified working"

success "Done! Restart Claude Code to activate."
