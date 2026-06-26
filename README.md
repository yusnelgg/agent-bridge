# Agent Bridge

[![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen)](https://github.com/z4d3s/agent-bridge/pulls)

**Agent Bridge** connects two AI coding assistants (opencode, Claude Code, Cursor) running on different PCs so they can collaborate autonomously.

```
┌─────────────────────┐          NATS           ┌─────────────────────┐
│  FRONTEND (Uruguay) │◄───────────────────────►│  BACKEND (EEUU)     │
│                     │     (Tailscale/VPS)     │                     │
│  opencode / Claude  │                          │  opencode / Claude  │
│  "agent ask --wait" │                          │  "agent listen"     │
└─────────────────────┘                          └─────────────────────┘
```

---

## Features

- **Zero dependencies** — single binary, no Docker, no Python, no Node.
- **AI-agnostic** — works with opencode, Claude Code, Cursor, or any MCP-capable AI.
- **Multi-PC** — connect AIs across the internet via Tailscale, ZeroTier, or a VPS.
- **Embedded NATS** — no external message broker to install or configure.
- **MCP server** — exposes tools so the AI can send/receive messages natively.
- **SQLite persistence** — messages and tasks survive restarts.

## Install

### Linux / macOS

```bash
curl -L https://github.com/z4d3s/agent-bridge/releases/latest/download/agent-bridge-linux.zip -o agent-bridge.zip
unzip agent-bridge.zip
cd agent-bridge-dist
./install.sh
```

### Windows

Download `agent-bridge-windows.zip` from [Releases](https://github.com/z4d3s/agent-bridge/releases/latest), unzip, and double-click `install.bat`.

### From source

```bash
git clone https://github.com/z4d3s/agent-bridge.git
cd agent-bridge
make build
sudo cp agent agent-bridge /usr/local/bin/
```

## Quick Start

### 1. Start the bridges

Open **two terminals**.

**Terminal 1 — Frontend** (hosts NATS):
```bash
agent-bridge -config configs/frontend.yaml
```

**Terminal 2 — Backend**:
```bash
export AGENT_BRIDGE=http://localhost:9091
agent-bridge -config configs/backend.yaml
```

### 2. The AI flow

**Frontend AI** asks for something:
```bash
agent ask --wait backend "Create a REST API with CRUD for users"
```

**Backend AI** listens, codes, and responds:
```bash
export AGENT_BRIDGE=http://localhost:9091
agent listen           # waits for a message
# ... reads the request, writes the code ...
agent respond frontend "Done. Endpoints: POST/GET/PUT/DELETE /api/users"
```

### 3. Connect remotely

Set `nats_url` in both configs to the IP of the machine hosting NATS:

```yaml
nats_url: "nats://100.x.x.x:4222"    # Tailscale IP
```

No other changes needed.

## Commands

| Command | Description |
|---|---|
| `agent ask --wait <agent> <message>` | Send a message and block until reply |
| `agent listen` | Block until a new message arrives |
| `agent respond <agent> <message>` | Reply to an agent |
| `agent check` | Check for new messages (non-blocking) |
| `agent delegate --wait <agent> <task>` | Delegate a task and wait for result |
| `agent tasks` | List pending tasks |

## Architecture

```
┌──────────────┐     MCP/HTTP      ┌──────────────┐     NATS      ┌──────────────┐
│   AI Tool    │◄─────────────────►│  agent-bridge │◄────────────►│  agent-bridge │
│  (opencode)  │   ask/respond     │   (frontend)  │   pub/sub     │   (backend)   │
└──────────────┘                   └──────┬───────┘               └──────┬───────┘
                                          │                              │
                                    ┌─────▼─────┐                  ┌─────▼─────┐
                                    │   SQLite   │                  │   SQLite   │
                                    └───────────┘                  └───────────┘
```

## Configuration

See [`configs/frontend.yaml`](configs/frontend.yaml) and [`configs/backend.yaml`](configs/backend.yaml).

```yaml
identity: "frontend"        # or "backend"
listen_addr: "127.0.0.1:9090"
nats_url: "nats://127.0.0.1:4222"
server_mode: true           # true if this instance hosts NATS
mcp_server: false           # true to expose MCP tools (Claude Code)
```

## Documentation

Full docs at **[z4d3s.github.io/agent-bridge](https://z4d3s.github.io/agent-bridge)**

## License

[MIT](LICENSE)
