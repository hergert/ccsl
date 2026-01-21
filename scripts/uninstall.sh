#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() { echo -e "${BLUE}ℹ ${1}${NC}"; }
print_success() { echo -e "${GREEN}✅ ${1}${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  ${1}${NC}"; }

BINARY_PATH="$HOME/.local/bin/ccsl"
CLAUDE_DIR="$HOME/.claude"
SETTINGS_FILE="$CLAUDE_DIR/settings.json"

print_info "Uninstalling ccsl..."

# Remove binary
if [ -f "$BINARY_PATH" ]; then
    rm "$BINARY_PATH"
    print_success "Removed ccsl binary"
fi

# Remove statusLine config from Claude settings
if [ -f "$SETTINGS_FILE" ]; then
    print_info "Removing statusLine config from Claude settings..."
    python3 -c "
import json
import sys

settings_file = '$SETTINGS_FILE'

try:
    with open(settings_file, 'r') as f:
        settings = json.load(f)

    if 'statusLine' in settings:
        del settings['statusLine']

        with open(settings_file, 'w') as f:
            json.dump(settings, f, indent=2)
        print('Removed statusLine from Claude settings')
    else:
        print('No statusLine config found in Claude settings')

except Exception as e:
    print(f'Error updating settings: {e}')
"
fi

print_warning "Config files in ~/.config/ccsl/ were left intact"
print_success "ccsl uninstalled successfully!"
print_info "Restart Claude Code to complete removal."
