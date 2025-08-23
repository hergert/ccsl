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

# Download URL (will be updated when we have releases)
BINARY_URL="https://github.com/hergert/ccsl/releases/latest/download/ccsl-${OS}-${ARCH}"

# Install location
INSTALL_DIR="$HOME/.local/bin"
BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"

# Create install directory
mkdir -p "$INSTALL_DIR"

# For now, we'll build locally since we don't have releases yet
print_info "Building ccsl from source..."

# Check if Go is available
if ! command -v go &> /dev/null; then
    print_error "Go is required to build ccsl from source"
    print_info "Install Go from https://golang.org/dl/"
    exit 1
fi

# To support `curl | bash`, we need to clone the repo into a temp dir.
if ! command -v git &> /dev/null; then
    print_error "Git is required to build ccsl from source"
    exit 1
fi

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT # Cleanup on exit

print_info "Cloning repository into a temporary directory..."
if git clone --depth 1 https://github.com/hergert/ccsl.git "$TMP_DIR"; then
    cd "$TMP_DIR"
else
    print_error "Failed to clone repository."
    exit 1
fi

# Build the binary
print_info "Building in $(pwd)..."

if go build -o "$BINARY_PATH" ./cmd/ccsl; then
    print_success "Built ccsl binary at $BINARY_PATH"
else
    print_error "Failed to build ccsl"
    exit 1
fi

# Make sure binary is executable
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
template = "{model}  {cwd}  {agent}  {git?prefix=  }{prompt?prefix= — 🗣 }"
truncate = 120
padding = 0

[theme]
mode = "auto"
icons = true
ansi = true

[plugins]
order = ["model", "cwd", "agent", "git", "prompt"]

[plugin.git]
type = "builtin"
style = "dim"
timeout_ms = 90
cache_ttl_ms = 300

[limits]
per_plugin_timeout_ms = 120
total_budget_ms = 220
EOF
    print_success "Created default config at $CCSL_CONFIG_DIR/config.toml"
fi

# Create Claude directories
mkdir -p "$CLAUDE_DIR"

# Backup existing settings
if [ -f "$SETTINGS_FILE" ]; then
    BACKUP_FILE="${SETTINGS_FILE}.backup.$(date +%Y%m%d_%H%M%S)"
    cp "$SETTINGS_FILE" "$BACKUP_FILE"
    print_success "Backed up existing settings to $(basename $BACKUP_FILE)"
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

# Load existing settings
try:
    with open(settings_file, 'r') as f:
        settings = json.load(f)
except:
    settings = {"\$schema": "https://json.schemastore.org/claude-code-settings.json"}

# Add/update statusline configuration
settings["statusLine"] = {
    "type": "command",
    "command": "$BINARY_PATH",
    "padding": 0
}

# Write updated settings
try:
    with open(settings_file, 'w') as f:
        json.dump(settings, f, indent=2)
    print("Settings updated successfully")
except Exception as e:
    print(f"Error updating settings: {e}")
    sys.exit(1)
EOF

# Test the installation
print_info "Testing ccsl installation..."
echo '{"model":{"display_name":"Test"},"workspace":{"current_dir":"'$(pwd)'"}}' | "$BINARY_PATH" > /dev/null

if [ $? -eq 0 ]; then
    print_success "ccsl installed successfully!"
    print_info "Restart Claude Code to see the new status bar."
    print_info "Config location: $CCSL_CONFIG_DIR/config.toml"
else
    print_error "Installation test failed. Check the ccsl binary."
    exit 1
fi

print_success "Installation complete!"