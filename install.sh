#!/bin/bash
set -e

SAILKIT_DIR="$(cd "$(dirname "$0")" && pwd)"
PARENT_DIR="$(dirname "$SAILKIT_DIR")"

echo "Sailkit Installer"
echo "================="
echo ""

# Build Go binaries
echo "Building Go binaries..."
cd "$SAILKIT_DIR"
mkdir -p bin

for cmd in cmd/*/; do
    name=$(basename "$cmd")
    go build -o "bin/$name" "./$cmd"
    echo "  Built: $name"
done

# Install skills
echo ""
echo "Where should Sailkit skills be installed?"
echo "  1) Project-level: $PARENT_DIR/.claude/skills (recommended)"
echo "  2) Global: ~/.claude/skills"
echo ""
read -p "Choice [1]: " SCOPE_CHOICE
SCOPE_CHOICE="${SCOPE_CHOICE:-1}"

if [ "$SCOPE_CHOICE" = "2" ]; then
    TARGET_DIR="$HOME/.claude/skills"
else
    TARGET_DIR="$PARENT_DIR/.claude/skills"
fi

echo ""
read -p "Install skills to $TARGET_DIR? [Y/n]: " CONFIRM
CONFIRM="${CONFIRM:-Y}"

if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo "Skipped skill installation."
else
    mkdir -p "$TARGET_DIR"
    for skill_dir in "$SAILKIT_DIR/skills"/*/; do
        skill_name=$(basename "$skill_dir")
        SKILL_LINK="$TARGET_DIR/$skill_name"
        [ -L "$SKILL_LINK" ] && rm "$SKILL_LINK"
        [ -e "$SKILL_LINK" ] && { echo "Error: $SKILL_LINK exists"; exit 1; }
        ln -s "$skill_dir" "$SKILL_LINK"
        echo "  Linked: $skill_name"
    done
fi

echo ""
echo "Done! Add to PATH:"
echo "  export PATH=\"\$PATH:$SAILKIT_DIR/bin\""
