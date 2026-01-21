#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
print_info() { echo -e "${BLUE}ℹ ${1}${NC}"; }
print_success() { echo -e "${GREEN}✅ ${1}${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  ${1}${NC}"; }
print_error() { echo -e "${RED}❌ ${1}${NC}"; }

# Configuration
CLAUDE_DIR="$HOME/.claude"
CCSL_CONFIG_DIR="$HOME/.config/ccsl"
SETTINGS_FILE="$CLAUDE_DIR/settings.json"
BINARY_NAME="ccsl"

print_info "Installing ccsl (Claude Code StatusLine)..."

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64) ARCH="arm64" ;;
    *) print_error "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case $OS in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    *) print_error "Unsupported operating system: $OS"; exit 1 ;;
esac

# Install location
INSTALL_DIR="$HOME/.local/bin"
BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"

# Create install directory
mkdir -p "$INSTALL_DIR"

# Build from source
print_info "Building ccsl from source..."

# Check if Go is available
if ! command -v go &> /dev/null; then
    print_error "Go is required to build ccsl from source"
    print_info "Install Go from https://golang.org/dl/"
    exit 1
fi

# To support `curl | bash`, clone the repo into a temp dir
if ! command -v git &> /dev/null; then
    print_error "Git is required to build ccsl from source"
    exit 1
fi

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

print_info "Cloning repository..."
if git clone --depth 1 https://github.com/hergert/ccsl.git "$TMP_DIR"; then
    cd "$TMP_DIR"
else
    print_error "Failed to clone repository."
    exit 1
fi

# Build the binary
print_info "Building..."

if go build -o "$BINARY_PATH" ./cmd/ccsl; then
    print_success "Built ccsl at $BINARY_PATH"
else
    print_error "Failed to build ccsl"
    exit 1
fi

chmod +x "$BINARY_PATH"

# Check if install directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    print_warning "$INSTALL_DIR is not in your PATH"
    print_info "Add this to your shell profile:"
    print_info "export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

# Create config directory
mkdir -p "$CCSL_CONFIG_DIR"

# Create default config if it doesn't exist
if [ ! -f "$CCSL_CONFIG_DIR/config.toml" ]; then
    cat > "$CCSL_CONFIG_DIR/config.toml" << 'EOF'
[ui]
template = "{model} {cwd}{git?prefix= }{ctx?prefix= }{cost?prefix= }"
truncate = 120

[theme]
icons = true
ansi = true

[plugin.git]
timeout_ms = 80

[limits]
per_plugin_timeout_ms = 100
total_budget_ms = 200
EOF
    print_success "Created config at $CCSL_CONFIG_DIR/config.toml"
fi

# Create Claude directories
mkdir -p "$CLAUDE_DIR"

# Backup existing settings
if [ -f "$SETTINGS_FILE" ]; then
    BACKUP_FILE="${SETTINGS_FILE}.backup.$(date +%Y%m%d_%H%M%S)"
    cp "$SETTINGS_FILE" "$BACKUP_FILE"
    print_success "Backed up settings to $(basename $BACKUP_FILE)"
fi

# Create default Claude settings if none exist
if [ ! -f "$SETTINGS_FILE" ]; then
    cat > "$SETTINGS_FILE" << 'EOF'
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json"
}
EOF
fi

# Update Claude settings with ccsl
print_info "Updating Claude Code settings..."

python3 << EOF
import json
import sys

settings_file = "$SETTINGS_FILE"

try:
    with open(settings_file, 'r') as f:
        settings = json.load(f)
except:
    settings = {"\$schema": "https://json.schemastore.org/claude-code-settings.json"}

settings["statusLine"] = {
    "type": "command",
    "command": "$BINARY_PATH",
    "padding": 0
}

try:
    with open(settings_file, 'w') as f:
        json.dump(settings, f, indent=2)
    print("Settings updated")
except Exception as e:
    print(f"Error: {e}")
    sys.exit(1)
EOF

# Test the installation
print_info "Testing..."
echo '{"model":{"display_name":"Test"},"workspace":{"current_dir":"'$(pwd)'"}}' | "$BINARY_PATH" > /dev/null

if [ $? -eq 0 ]; then
    print_success "ccsl installed successfully!"
    print_info "Config: $CCSL_CONFIG_DIR/config.toml"
    print_info "Restart Claude Code to see the new status bar."
else
    print_error "Installation test failed."
    exit 1
fi
