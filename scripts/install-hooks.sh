#!/usr/bin/env sh
# Install git hooks from scripts/ into .git/hooks/
set -e

REPO_ROOT="$(git rev-parse --show-toplevel)"
HOOKS_DIR="$REPO_ROOT/.git/hooks"
SCRIPTS_DIR="$REPO_ROOT/scripts"

install_hook() {
    name="$1"
    src="$SCRIPTS_DIR/$name"
    dst="$HOOKS_DIR/$name"
    if [ -f "$src" ]; then
        cp "$src" "$dst"
        chmod +x "$dst"
        echo "  installed $name"
    fi
}

install_hook "pre-commit"
install_hook "commit-msg"

echo "Git hooks installed. Run 'git commit' to verify."
