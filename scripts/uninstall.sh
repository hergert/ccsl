#!/bin/bash
# Uninstall ccsl
# Usage: curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/uninstall.sh | bash

set -e

BINARY_PATH="$HOME/.local/bin/ccsl"
SETTINGS_FILE="$HOME/.claude/settings.json"

info()    { printf '\033[0;34m%s\033[0m\n' "$1"; }
success() { printf '\033[0;32m%s\033[0m\n' "$1"; }

info "Uninstalling ccsl..."

# Remove binary
[ -f "$BINARY_PATH" ] && rm "$BINARY_PATH" && success "Removed $BINARY_PATH"

# Remove from Claude settings
if [ -f "$SETTINGS_FILE" ]; then
    python3 -c "
import json
with open('$SETTINGS_FILE', 'r') as f: s = json.load(f)
s.pop('statusLine', None)
with open('$SETTINGS_FILE', 'w') as f: json.dump(s, f, indent=2)
" 2>/dev/null && success "Removed statusLine from Claude settings"
fi

success "Uninstalled! Restart Claude Code."
