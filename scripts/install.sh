#!/bin/bash
# Install ccsl (Claude Code StatusLine)
# Usage: curl -fsSL https://raw.githubusercontent.com/hergert/ccsl/main/scripts/install.sh | bash
#        ... | bash -s -- --build    # build from source instead of fetching a release binary

set -euo pipefail

REPO="hergert/ccsl"
INSTALL_DIR="$HOME/.local/bin"
BINARY_PATH="$INSTALL_DIR/ccsl"
SETTINGS_FILE="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/settings.json"

info()    { printf '\033[0;34m%s\033[0m\n' "$1"; }
success() { printf '\033[0;32m%s\033[0m\n' "$1"; }
warn()    { printf '\033[0;33m%s\033[0m\n' "$1"; }
error()   { printf '\033[0;31m%s\033[0m\n' "$1"; exit 1; }

MODE="release"
[ "${1:-}" = "--build" ] && MODE="build"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

sha256_check() {
    # args: <checksums-file> <filename>; run from the file's directory
    if command -v shasum &>/dev/null; then
        grep " $2\$" "$1" | shasum -a 256 -c - >/dev/null
    elif command -v sha256sum &>/dev/null; then
        grep " $2\$" "$1" | sha256sum -c - >/dev/null
    else
        return 1
    fi
}

install_release() {
    local os arch tag asset base
    case "$(uname -s)" in
        Darwin) os="darwin" ;;
        Linux)  os="linux" ;;
        *) return 1 ;;
    esac
    case "$(uname -m)" in
        arm64|aarch64) arch="arm64" ;;
        x86_64|amd64)  arch="amd64" ;;
        *) return 1 ;;
    esac

    tag=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null \
        | grep -m1 '"tag_name"' | cut -d'"' -f4) || return 1
    [ -n "$tag" ] || return 1

    asset="ccsl_${os}_${arch}.tar.gz"
    base="https://github.com/$REPO/releases/download/$tag"
    info "Downloading ccsl $tag ($os/$arch)..."
    curl -fsSL -o "$TMP_DIR/$asset" "$base/$asset" || return 1
    curl -fsSL -o "$TMP_DIR/checksums.txt" "$base/checksums.txt" || return 1
    (cd "$TMP_DIR" && sha256_check checksums.txt "$asset") || { warn "Checksum mismatch for $asset"; return 1; }

    tar -xzf "$TMP_DIR/$asset" -C "$TMP_DIR" ccsl
    mkdir -p "$INSTALL_DIR"
    install -m 0755 "$TMP_DIR/ccsl" "$BINARY_PATH"
    success "Installed ccsl $tag to $BINARY_PATH"
}

install_build() {
    command -v go &>/dev/null || error "Go required for --build. Install from https://go.dev/dl/"
    command -v git &>/dev/null || error "Git required for --build."
    info "Building from source..."
    git clone --depth 1 --quiet "https://github.com/$REPO.git" "$TMP_DIR/src"
    mkdir -p "$INSTALL_DIR"
    (cd "$TMP_DIR/src" && go build -o "$BINARY_PATH" ./cmd/ccsl)
    success "Built and installed $BINARY_PATH"
}

info "Installing ccsl..."
if [ "$MODE" = "release" ]; then
    install_release || { warn "No release binary available; building from source instead."; install_build; }
else
    install_build
fi

[[ ":$PATH:" != *":$INSTALL_DIR:"* ]] && warn "Add to PATH: export PATH=\"\$HOME/.local/bin:\$PATH\""

# Point Claude Code at the binary. Merges into any existing statusLine block:
# type/command are set, everything else the user configured is kept.
if command -v python3 &>/dev/null; then
    python3 - "$SETTINGS_FILE" "$BINARY_PATH" <<'PY' && success "Configured statusLine in $SETTINGS_FILE"
import json, os, sys

path, binary = sys.argv[1], sys.argv[2]
settings = {}
if os.path.exists(path):
    with open(path) as f:
        settings = json.load(f)
status_line = settings.get("statusLine")
if not isinstance(status_line, dict):
    status_line = {}
status_line["type"] = "command"
status_line["command"] = binary
status_line.setdefault("padding", 0)
status_line.setdefault("refreshInterval", 60)
settings["statusLine"] = status_line
os.makedirs(os.path.dirname(path), exist_ok=True)
with open(path, "w") as f:
    json.dump(settings, f, indent=2)
PY
else
    warn "python3 not found. Add to $SETTINGS_FILE yourself:"
    warn "  \"statusLine\": {\"type\": \"command\", \"command\": \"$BINARY_PATH\", \"padding\": 0, \"refreshInterval\": 60}"
fi

"$BINARY_PATH" --version && success "Done! Restart Claude Code to activate."
