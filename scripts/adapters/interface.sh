#!/bin/bash
# Adapter Interface for Issue Trackers
#
# Each adapter must implement these functions:
#   adapter_init          - Setup/validation
#   adapter_create_issue  - Create new issue, output issue ID
#   adapter_get_issue     - Get issue data as JSON
#   adapter_update_issue  - Update existing issue
#   adapter_list_issues   - List issues with label, output JSON array
#   adapter_close_issue   - Close/complete issue
#   adapter_reopen_issue  - Reopen issue
#
# All functions receive repo identifier as first arg.
# JSON output format for get/list:
# {
#   "id": "42",
#   "title": "Issue title",
#   "body": "Issue body markdown",
#   "state": "open|closed",
#   "labels": ["plan", "enhancement"],
#   "updated_at": "2026-01-19T10:00:00Z"
# }

# Load adapter based on config or env
load_adapter() {
    local repo="$1"
    local adapter="${PLAN_ADAPTER:-github}"

    # Check for repo-specific adapter in config
    local config_file="$WORKSPACE_DIR/plans/.config"
    if [ -f "$config_file" ]; then
        local repo_adapter=$(grep "^$repo:" "$config_file" | cut -d: -f2)
        [ -n "$repo_adapter" ] && adapter="$repo_adapter"
    fi

    local adapter_file="$SCRIPT_DIR/adapters/${adapter}.sh"
    if [ ! -f "$adapter_file" ]; then
        echo "Error: Adapter '$adapter' not found at $adapter_file" >&2
        exit 1
    fi

    source "$adapter_file"

    # Validate adapter implements required functions
    for fn in adapter_init adapter_create_issue adapter_get_issue adapter_update_issue adapter_list_issues; do
        if ! type "$fn" &>/dev/null; then
            echo "Error: Adapter '$adapter' missing function: $fn" >&2
            exit 1
        fi
    done
}

# Parse frontmatter from markdown file
# Output: key=value lines
parse_frontmatter() {
    local file="$1"

    # Check for frontmatter delimiter
    if ! head -1 "$file" | grep -q '^---$'; then
        return 1
    fi

    # Extract between first and second ---
    sed -n '2,/^---$/p' "$file" | sed '/^---$/d' | while read -r line; do
        key=$(echo "$line" | cut -d: -f1 | tr -d ' ')
        value=$(echo "$line" | cut -d: -f2- | sed 's/^ *//')
        echo "$key=$value"
    done
}

# Get frontmatter value
get_frontmatter() {
    local file="$1"
    local key="$2"
    parse_frontmatter "$file" | grep "^$key=" | cut -d= -f2-
}

# Extract body (everything after frontmatter)
get_body() {
    local file="$1"

    if ! head -1 "$file" | grep -q '^---$'; then
        cat "$file"
        return
    fi

    # Skip first --- and content until second ---
    awk 'BEGIN{skip=1} /^---$/{if(skip){skip=0;next}else{found=1;next}} found{print}' "$file"
}

# Update or add frontmatter field
set_frontmatter() {
    local file="$1"
    local key="$2"
    local value="$3"

    if grep -q "^$key:" "$file"; then
        # Update existing
        sed -i '' "s/^$key:.*/$key: $value/" "$file" 2>/dev/null || \
        sed -i "s/^$key:.*/$key: $value/" "$file"
    else
        # Add after first ---
        sed -i '' "s/^---$/---\n$key: $value/" "$file" 2>/dev/null || \
        sed -i "s/^---$/---\n$key: $value/" "$file"
    fi
}
