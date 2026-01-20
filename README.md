<p align="center">
  <img src="https://img.shields.io/badge/⚓-Bearing-blue?style=for-the-badge&logoColor=white" alt="Bearing" />
</p>

<h1 align="center">⚓ Bearing</h1>

<p align="center">
  <strong>Worktree-based workflow for parallel AI-assisted development</strong>
</p>

<p align="center">
  <a href="https://bearing.dev"><img src="https://img.shields.io/badge/📖_Docs-bearing.dev-blue?style=flat-square" alt="Documentation" /></a>
  <a href="https://github.com/joshribakoff/bearing/actions"><img src="https://img.shields.io/github/actions/workflow/status/joshribakoff/bearing/go.yml?style=flat-square&label=build" alt="Build Status" /></a>
  <a href="https://goreportcard.com/report/github.com/joshribakoff/bearing"><img src="https://goreportcard.com/badge/github.com/joshribakoff/bearing?style=flat-square" alt="Go Report Card" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="License" /></a>
</p>

<p align="center">
  <a href="https://www.joshribakoff.com/blog/deliberate-ai-use/">📝 Read the blog post</a> •
  <a href="https://bearing.dev">📖 Documentation</a> •
  <a href="#-quick-start">🚀 Quick Start</a>
</p>

---

## ✨ Why Bearing?

When multiple AI agents work on the same codebase, **they step on each other**. Branch switching in shared folders causes conflicts, lost work, and confusion.

**Bearing enforces a worktree-per-task pattern** that keeps every agent isolated:

- 🔒 **Isolation** — Each task gets its own directory. No conflicts.
- 🚀 **Parallelism** — Run 10 Claude sessions on 10 features simultaneously
- 📊 **Visibility** — See all active work at a glance
- 🔄 **Workflow** — Track purpose, status, and relationships
- ⚡ **Scale** — Thousands of worktrees across hundreds of repos

---

## 🖥️ Beautiful Terminal UI

Inspired by lazygit, the Bearing TUI gives you full visibility into your workspace:

![Bearing TUI](docs/public/images/tui-screenshot.svg)

**Features:**
- 📁 Browse all projects and worktrees
- 🎯 Vim-style navigation (`j/k`, `h/l`)
- 📋 Numbered panel switching (like lazygit)
- 🔍 Health status at a glance (dirty, unpushed, PR state)
- 🌙 Darcula-inspired dark theme

```bash
# Install TUI (Python 3.10+)
pip install bearing-tui

# Run
bearing-tui
```

---

## 🚀 Quick Start

### Install CLI

```bash
# Clone and build
git clone https://github.com/joshribakoff/bearing ~/Projects/bearing
cd ~/Projects/bearing
go build -o bearing ./cmd/bearing
sudo mv bearing /usr/local/bin/

# Initialize your workspace
cd ~/Projects
bearing init
```

### Create Your First Worktree

```bash
# Create a worktree for a new feature
bearing worktree new myapp feature-auth

# List all worktrees
bearing worktree list

# Clean up after merging
bearing worktree cleanup myapp feature-auth
```

---

## 📁 Workspace Layout

Bearing uses a flat workspace structure for maximum visibility:

```
~/Projects/
├── 📦 bearing/                 # Bearing itself
├── 📦 myapp/                   # Base folder (stays on main)
├── 🔀 myapp-feature-auth/      # Worktree for auth feature
├── 🔀 myapp-fix-bug-123/       # Worktree for bug fix
├── 📦 other-project/           # Another base folder
├── 🔀 other-project-refactor/  # Its worktree
├── 📄 workflow.jsonl           # Workflow state (committable)
└── 📄 local.jsonl              # Local worktree paths
```

**Base folders stay on `main`**. Worktrees are created for each task. This scales to **thousands of worktrees**.

---

## 🛠️ Commands

| Command | Description |
|---------|-------------|
| `bearing worktree new <repo> <branch>` | 🆕 Create worktree for branch |
| `bearing worktree cleanup <repo> <branch>` | 🧹 Remove worktree after merge |
| `bearing worktree sync` | 🔄 Rebuild manifest from git |
| `bearing worktree list` | 📋 Display ASCII table |
| `bearing worktree status` | 📊 Show health (dirty, PR) |
| `bearing worktree check` | ✅ Validate invariants |
| `bearing daemon start` | 👻 Start health monitor |

---

## 🤖 Claude Code Integration

Bearing integrates with Claude Code's hook system:

```json
{
  "hooks": {
    "UserPromptSubmit": [{
      "hooks": [{
        "type": "command",
        "command": "bearing worktree check --json"
      }]
    }]
  }
}
```

**What it does:**
- ✅ Checks invariants before every Claude action
- ⚠️ Warns when base folders drift from main
- 🔧 Claude can auto-fix violations

---

## 🏗️ Architecture

| Layer | Responsibility |
|-------|----------------|
| **Git** | Source of truth (submodules, worktrees) |
| **Manifest** | Workflow metadata (`workflow.jsonl`) |
| **CLI** | Orchestration & guardrails |
| **Daemon** | Background health monitoring |
| **TUI** | Visual workspace browser |

---

## 📊 State Files

**workflow.jsonl** (committable):
```jsonl
{"repo":"myapp","branch":"feature","purpose":"Add auth","status":"in_progress"}
```

**local.jsonl** (local only):
```jsonl
{"folder":"myapp-feature","repo":"myapp","branch":"feature","base":false}
```

---

## 🧪 Testing

```bash
# Go tests
go test ./...

# TUI tests
cd tui && make test
```

---

## 📚 Learn More

- 📖 [Full Documentation](https://bearing.dev)
- 📝 [Blog Post: Deliberate AI Use](https://www.joshribakoff.com/blog/deliberate-ai-use/)
- 🐛 [Report Issues](https://github.com/joshribakoff/bearing/issues)

---

<p align="center">
  Made with ⚓ for the AI-assisted development era
</p>
