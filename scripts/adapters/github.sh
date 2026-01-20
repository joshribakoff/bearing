#!/bin/bash
# GitHub Issues Adapter
# Uses gh CLI - requires `gh auth login` first
# No API keys stored - uses gh's auth mechanism

adapter_init() {
    if ! command -v gh &>/dev/null; then
        echo "Error: gh CLI not found. Install: brew install gh" >&2
        return 1
    fi

    if ! gh auth status &>/dev/null; then
        echo "Error: Not authenticated. Run: gh auth login" >&2
        return 1
    fi

    return 0
}

# Create issue, output issue number
# Args: repo title body [labels]
adapter_create_issue() {
    local repo="$1"
    local title="$2"
    local body="$3"
    local labels="${4:-plan}"

    # gh outputs the URL, extract issue number
    local url
    url=$(gh issue create \
        --repo "$repo" \
        --title "$title" \
        --body "$body" \
        --label "$labels" \
        2>/dev/null)

    if [ $? -ne 0 ]; then
        echo "Error: Failed to create issue" >&2
        return 1
    fi

    # Extract issue number from URL
    echo "$url" | grep -oE '[0-9]+$'
}

# Get issue as compact JSON
# Args: repo issue_number
adapter_get_issue() {
    local repo="$1"
    local issue="$2"

    # Request only needed fields
    gh issue view "$issue" \
        --repo "$repo" \
        --json number,title,body,state,labels,updatedAt \
        2>/dev/null | \
    jq -c '{
        id: .number|tostring,
        title: .title,
        body: .body,
        state: (.state|ascii_downcase),
        labels: [.labels[].name],
        updated_at: .updatedAt
    }'
}

# Update issue
# Args: repo issue_number title body
adapter_update_issue() {
    local repo="$1"
    local issue="$2"
    local title="$3"
    local body="$4"

    gh issue edit "$issue" \
        --repo "$repo" \
        --title "$title" \
        --body "$body" \
        &>/dev/null
}

# List issues with plan label, compact JSON array
# Args: repo [state: open|closed|all]
adapter_list_issues() {
    local repo="$1"
    local state="${2:-all}"

    gh issue list \
        --repo "$repo" \
        --label "plan" \
        --state "$state" \
        --limit 100 \
        --json number,title,state,updatedAt \
        2>/dev/null | \
    jq -c '[.[] | {
        id: .number|tostring,
        title: .title,
        state: (.state|ascii_downcase),
        updated_at: .updatedAt
    }]'
}

# Close issue
# Args: repo issue_number
adapter_close_issue() {
    local repo="$1"
    local issue="$2"

    gh issue close "$issue" --repo "$repo" &>/dev/null
}

# Reopen issue
# Args: repo issue_number
adapter_reopen_issue() {
    local repo="$1"
    local issue="$2"

    gh issue reopen "$issue" --repo "$repo" &>/dev/null
}

# Add labels to issue
# Args: repo issue_number labels (comma-separated)
adapter_add_labels() {
    local repo="$1"
    local issue="$2"
    local labels="$3"

    gh issue edit "$issue" --repo "$repo" --add-label "$labels" &>/dev/null
}
