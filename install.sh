#!/bin/bash
set -e

SAILKIT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Sailkit Installer"
echo "================="
echo ""
echo "Sailkit enforces worktree-based workflows for parallel AI development."
echo ""

# Detect parent directory (likely Projects folder)
PARENT_DIR="$(dirname "$SAILKIT_DIR")"

echo "Where should Sailkit skills be installed?"
echo ""
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
echo "This will create symlinks in: $TARGET_DIR"
read -p "Continue? [Y/n]: " CONFIRM
CONFIRM="${CONFIRM:-Y}"

if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

# Create target directory
mkdir -p "$TARGET_DIR"

# Create symlink for worktree skill
SKILL_LINK="$TARGET_DIR/worktree"
if [ -L "$SKILL_LINK" ]; then
    echo "Removing existing symlink: $SKILL_LINK"
    rm "$SKILL_LINK"
elif [ -e "$SKILL_LINK" ]; then
    echo "Error: $SKILL_LINK exists and is not a symlink"
    exit 1
fi

ln -s "$SAILKIT_DIR/skills/worktree" "$SKILL_LINK"
echo "Created symlink: $SKILL_LINK -> $SAILKIT_DIR/skills/worktree"

# Add scripts to PATH suggestion
echo ""
echo "Done! To use worktree-* commands, add to your shell profile:"
echo ""
echo "  export PATH=\"\$PATH:$SAILKIT_DIR/scripts\""
echo ""
echo "Or invoke directly: $SAILKIT_DIR/scripts/worktree-new"
