#!/bin/bash
# Uninstall ccsl
# Usage: curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/uninstall.sh | bash

set -euo pipefail

BINARY_PATH="$HOME/.local/bin/ccsl"
SETTINGS_FILE="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/settings.json"

info()    { printf '\033[0;34m%s\033[0m\n' "$1"; }
success() { printf '\033[0;32m%s\033[0m\n' "$1"; }

info "Uninstalling ccsl..."

[ -f "$BINARY_PATH" ] && rm "$BINARY_PATH" && success "Removed $BINARY_PATH"

# Remove the statusLine block only if it still points at ccsl.
if [ -f "$SETTINGS_FILE" ] && command -v python3 &>/dev/null; then
    python3 - "$SETTINGS_FILE" <<'PY' && success "Cleaned statusLine from Claude settings"
import json, sys

path = sys.argv[1]
with open(path) as f:
    settings = json.load(f)
status_line = settings.get("statusLine")
if isinstance(status_line, dict) and "ccsl" in str(status_line.get("command", "")):
    settings.pop("statusLine")
    with open(path, "w") as f:
        json.dump(settings, f, indent=2)
PY
fi

success "Uninstalled! Restart Claude Code."
