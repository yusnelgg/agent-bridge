# Agent Bridge

[![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen)](https://github.com/yusnelgg/agent-bridge/pulls)

**Agent Bridge** connects two AI coding assistants (opencode, Claude Code, Cursor) running on different PCs so they can collaborate autonomously.

```
┌─────────────────────┐          NATS           ┌─────────────────────┐
│  FRONTEND (Uruguay) │◄───────────────────────►│  BACKEND (EEUU)     │
│                     │   WebSocket push        │                     │
│  opencode / Claude  │   (real-time)            │  opencode / Claude  │
│  "agent-bridge ask" │                          │  "agent-bridge listen"│
└─────────────────────┘                          └─────────────────────┘
```

---

## Features

- **Zero dependencies** — single binary (`agent-bridge` does everything), no Docker, no Python, no Node.
- **Real-time** — WebSocket push (fallback a polling si no disponible).
- **AI-agnostic** — works with opencode, Claude Code, Cursor, or any MCP-capable AI.
- **Multi-PC** — connect AIs across the internet via Tailscale, ZeroTier, or a VPS.
- **Embedded NATS** — no external message broker to install or configure.
- **MCP server** — exposes tools so the AI can send/receive messages natively.
- **SQLite persistence** — messages and tasks survive restarts.

## One-command install

```bash
curl -fsSL https://raw.githubusercontent.com/yusnelgg/agent-bridge/master/scripts/install.sh | sh
```

Auto-detects OS, downloads latest release from GitHub, and installs.

### Manual install

**Linux / macOS:**
```bash
curl -L https://github.com/yusnelgg/agent-bridge/releases/latest/download/agent-bridge-linux.zip -o agent-bridge.zip
unzip agent-bridge.zip
cd agent-bridge-dist
./install.sh
```

**Windows:** Download `agent-bridge-windows.zip` from [Releases](https://github.com/yusnelgg/agent-bridge/releases/latest), unzip, and double-click `install.bat`.

**From source:**
```bash
git clone https://github.com/yusnelgg/agent-bridge.git
cd agent-bridge
make build
sudo cp agent-bridge /usr/local/bin/
```

## How it works

### Roles

**FRONTEND** — SOLO consume. Pide endpoints, recibe respuesta, implementa la UI.

**BACKEND** — SOLO programa. Recibe pedidos, escribe TODO el backend, responde con instrucciones de consumo.

Ninguno programa lo que le corresponde al otro. ([Ver REGLA DE ORO completa](AGENTS.md))

### Quick start

Open **two terminals**.

**Terminal 1 — Frontend** (hosts NATS):
```bash
agent-bridge serve -config ~/.agent-bridge/frontend.yaml
```

**Terminal 2 — Backend**:
```bash
export AGENT_BRIDGE=http://localhost:9091
agent-bridge serve -config ~/.agent-bridge/backend.yaml
```

### AI flow (no changes needed)

**Frontend AI** asks:
```bash
agent-bridge ask --wait backend "Create a REST API with CRUD for users"
```

**Backend AI** listens (real-time via WebSocket):
```bash
export AGENT_BRIDGE=http://localhost:9091
agent-bridge listen    # waits for a message (push, no polling)
# ... reads the request, writes the code ...
agent-bridge respond frontend "Done. Endpoints: POST/GET/PUT/DELETE /api/users"
```

### Connect remotely

Set `nats_url` in both configs to the IP of the machine hosting NATS:

```yaml
nats_url: "nats://100.x.x.x:4222"    # Tailscale IP
```

No other changes needed.

## Commands

| Command | Description |
|---|---|---|
| `agent-bridge serve -config <file>` | Start the bridge daemon |
| `agent-bridge ask --wait <agent> <message>` | Send a message and block until reply |
| `agent-bridge listen` | Block until a new message arrives (WebSocket push) |
| `agent-bridge respond <agent> <message>` | Reply to an agent |
| `agent-bridge check` | Check for new messages (non-blocking) |
| `agent-bridge delegate --wait <agent> <task>` | Delegate a task and wait for result |
| `agent-bridge tasks` | List pending tasks |
| `agent-bridge init` | Generate a config file |
| `agent-bridge watch` | Live message monitor |

## Architecture

```
┌──────────────┐     MCP/HTTP      ┌──────────────┐     NATS      ┌──────────────┐
│   AI Tool    │◄─────────────────►│  agent-bridge │◄────────────►│  agent-bridge │
│  (opencode)  │   ask/respond     │   (frontend)  │   pub/sub     │   (backend)   │
└──────────────┘                   └──┬───────┬───┘               └──┬───────┬───┘
                                      │       │                      │       │
                                 ┌────▼───┐ ┌▼────────┐        ┌────▼───┐ ┌▼────────┐
                                 │ SQLite │ │ WS Hub  │        │ SQLite │ │ WS Hub  │
                                 └────────┘ └─────────┘        └────────┘ └─────────┘
```

Messages arrive instantly via **WebSocket push**. If WebSocket is unavailable, `agent listen` falls back to polling (2s interval).

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

Full docs at **[yusnelgg.github.io/agent-bridge](https://yusnelgg.github.io/agent-bridge)**

## License

[MIT](LICENSE)
