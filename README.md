# ccmux — Colby's Claude Multiplexer

A terminal-based orchestrator for managing multiple [Claude Code](https://claude.ai/claude-code) agents working on tasks in parallel. Provides a unified tmux-backed interface to spawn, monitor, intervene with, and manage concurrent AI agents across git projects.

Each agent gets its own git worktree, branch, and tmux window — so multiple agents can work on different tasks in the same repo without conflicts.

## Prerequisites

- [tmux](https://github.com/tmux/tmux)
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) CLI (`claude`)
- [GitHub CLI](https://cli.github.com/) (`gh`)
- Git

## Installation

Download the latest binary for your platform from [GitHub Releases](https://github.com/CDFalcon/ccmux/releases):

```bash
# macOS (Apple Silicon)
curl -L https://github.com/CDFalcon/ccmux/releases/latest/download/ccmux-darwin-arm64 -o ccmux

# macOS (Intel)
curl -L https://github.com/CDFalcon/ccmux/releases/latest/download/ccmux-darwin-amd64 -o ccmux

# Linux (x86_64)
curl -L https://github.com/CDFalcon/ccmux/releases/latest/download/ccmux-linux-amd64 -o ccmux

# Linux (ARM64)
curl -L https://github.com/CDFalcon/ccmux/releases/latest/download/ccmux-linux-arm64 -o ccmux
```

Then make it executable and move it to your PATH:

```bash
chmod +x ccmux
mv ccmux /usr/local/bin/  # or ~/bin/, ~/.local/bin/, etc.
```

## Quick Start

### 1. Start a session

```bash
ccmux              # start or attach to the "default" session
ccmux my-project   # start or attach to a named session
```

Sessions are backed by tmux. Detach with `Ctrl+C Ctrl+C` and reattach later — all agents keep running.

### 2. Register a project

Before spawning agents you need to tell ccmux where your code lives.

1. Press **`p`** to open the project manager.
2. Press **`a`** to add a new project.
3. Enter a short name (e.g. `backend`) and the absolute path to the git repo.

Projects are stored globally in `~/.ccmux/projects.json` and shared across sessions.

### 3. Create a new task

1. Press **`n`** from the dashboard.
2. Select the project the task belongs to.
3. Pick a base branch (defaults to `origin/master` — search/filter to find others).
4. Describe what you want done. Multi-line input is supported with `Shift+Enter`.

ccmux creates an isolated git worktree and branch (`ccmux/<agent-id>`), opens a new tmux window, and launches Claude Code with your task. The agent shows up on the dashboard with a **running** spinner.

### 4. Work the queue

As agents work, items land in the **quick action queue** on the dashboard. Press **`q`** to pop the top item and take action. There are three kinds of queue items:

| Icon | Type | What happened | What you do |
|------|------|---------------|-------------|
| 💤 | **Idle** | Agent's terminal has been inactive for >10 s — it may be stuck or waiting for input. | Press `Enter` to jump into the agent's tmux window and see what's going on. You can also type a message to send directly to the agent's terminal. |
| ❓ | **Question** | Agent explicitly asked for help (via `ccmux need-help`). | Read the question in the details pane, jump to the agent, and respond. |
| 🔀 | **PR Ready** | Agent finished its task, pushed a branch, and opened a pull request. | Review the PR, then: **`a`**ccept (merge + cleanup), **`c`**omment (resume the agent to address feedback), **`r`**eject (close PR + cleanup), or **`b`**rowser (open the PR URL). |

### 5. Other dashboard actions

| Key | Action |
|-----|--------|
| `n` | Spawn a new agent task |
| `q` | Quick-action the top queue item |
| `j` | Jump to a specific agent's tmux window |
| `k` | Kill (terminate) a single agent |
| `K` | Kill the entire session and all agents |
| `p` | Manage registered projects |
| `u` | Check for ccmux updates |
