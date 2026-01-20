---
title: AI Features
description: Optional Claude CLI integration for classification and summarization
---

# AI Features

Bearing can optionally use Claude CLI for smart classification and summarization. **Disabled by default** - must opt-in.

## What It Does

- **Auto-generate purpose**: When creating worktrees, generate a description from the branch name
- **Classify priority**: P0/P1/P2/P3 based on plan content
- **Summarize health**: Prioritize which issues to fix first
- **Suggest fixes**: Generate commands to resolve workflow issues

## Why Opt-in Only

- Calls external API (costs money, even if cheap)
- Requires Claude CLI installed and authenticated
- Not everyone wants AI features
- Tools work fine without it (just less smart)

## Enable

Three ways to opt-in:

**1. Environment variable (session)**
```bash
export BEARING_AI_ENABLED=1
```

**2. User config (persistent)**
```bash
echo "ai_enabled: true" >> ~/.bearing
```

**3. Workspace config**
Add to `.bearing.yaml`:
```yaml
ai:
  enabled: true
```

## Auto-Generated Purpose

When you create a worktree without `--purpose`, Bearing can generate one:

```bash
worktree-new myrepo feature-add-dark-mode
# workflow.jsonl now has: "purpose": "Add dark mode toggle"
```

The AI infers the purpose from the branch name. Override with:
```bash
worktree-new myrepo feature-x --purpose "Explicit description"
```

## Cost

Uses Claude haiku model (~$0.0001 per call). A typical session generates maybe 5-10 calls = ~$0.001.

## Requirements

1. Claude CLI installed: `npm install -g @anthropic-ai/claude-cli`
2. Authenticated: `claude auth login`
3. Opt-in enabled via one of the methods above

## Graceful Degradation

If Claude CLI isn't available or AI isn't enabled:
- Tools work normally
- Purpose field stays empty (or use `--purpose`)
- No errors, just silent fallback
