#!/bin/bash
# Claude Code hook: Validate worktree safety
# Add to .claude/settings.json:
# "hooks": { "preToolCall": [{ "command": "sailkit-dev/hooks/validate-worktree.sh" }] }

# This hook checks if an agent is about to switch branches in a base folder

# Read tool call from stdin (Claude Code passes JSON)
TOOL_CALL=$(cat)

# Extract tool name and command
TOOL_NAME=$(echo "$TOOL_CALL" | grep -o '"tool":"[^"]*"' | cut -d'"' -f4)

if [ "$TOOL_NAME" = "Bash" ]; then
    COMMAND=$(echo "$TOOL_CALL" | grep -o '"command":"[^"]*"' | cut -d'"' -f4)

    # Check for dangerous git checkout/switch in base folders
    if echo "$COMMAND" | grep -qE '(git checkout|git switch)' && \
       ! echo "$COMMAND" | grep -qE '\-[bB]'; then
        # Trying to switch branches (not create new)
        echo "WARNING: Detected branch switch. Ensure you're in a worktree, not a base folder."
        echo "Base folders must stay on main. Use wt-new to create worktrees."
    fi
fi

# Always allow (hook is advisory for now)
exit 0
