#!/bin/bash
# Install ccsl (Claude Code StatusLine)
# Usage: curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/install.sh | bash

set -e

REPO="github.com/hergert/ccsl"
INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="ccsl"
BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"
CLAUDE_DIR="$HOME/.claude"
SETTINGS_FILE="$CLAUDE_DIR/settings.json"

# Colors
info()    { printf '\033[0;34m%s\033[0m\n' "$1"; }
success() { printf '\033[0;32m%s\033[0m\n' "$1"; }
warn()    { printf '\033[0;33m%s\033[0m\n' "$1"; }
error()   { printf '\033[0;31m%s\033[0m\n' "$1"; exit 1; }

info "Installing ccsl..."

# Check for Go
command -v go &>/dev/null || error "Go required. Install from https://go.dev/dl/"

# Install binary
mkdir -p "$INSTALL_DIR"
info "Building from source..."
GOBIN="$INSTALL_DIR" go install "$REPO/cmd/ccsl@latest"
success "Installed $BINARY_PATH"

# Check PATH
[[ ":$PATH:" != *":$INSTALL_DIR:"* ]] && warn "Add to PATH: export PATH=\"\$HOME/.local/bin:\$PATH\""

# Configure Claude Code
mkdir -p "$CLAUDE_DIR"

if [ -f "$SETTINGS_FILE" ]; then
    # Update existing settings
    python3 -c "
import json
with open('$SETTINGS_FILE', 'r') as f: s = json.load(f)
s['statusLine'] = {'type': 'command', 'command': '$BINARY_PATH', 'padding': 0}
with open('$SETTINGS_FILE', 'w') as f: json.dump(s, f, indent=2)
" 2>/dev/null || warn "Could not update settings.json"
else
    # Create new settings
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

success "Installed! Restart Claude Code to activate."
