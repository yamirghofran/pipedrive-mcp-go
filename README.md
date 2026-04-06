# Pipedrive MCP Server

A Go-based [Model Context Protocol](https://modelcontextprotocol.io) server for [Pipedrive CRM](https://www.pipedrive.com/). It exposes your Pipedrive data — deals, persons, organizations, pipelines, stages, leads, notes, and users — as tools and prompts that AI clients like Claude Desktop, Cursor, Windsurf, and other MCP-compatible applications can use directly.

Single static binary, zero runtime dependencies, ~10 MB Docker image.

---

## Table of Contents

- [What It Does](#what-it-does)
- [Quick Start](#quick-start)
  - [Prerequisites](#prerequisites)
  - [Option A — stdio (Local, Recommended)](#option-a--stdio-local-recommended)
  - [Option B — HTTP (Remote / Docker)](#option-b--http-remote--docker)
- [MCP Tools](#mcp-tools)
- [MCP Prompts](#mcp-prompts)
- [Configuration Reference](#configuration-reference)
- [Client Setup Guides](#client-setup-guides)
  - [Claude Desktop](#claude-desktop)
  - [Cursor](#cursor)
  - [Windsurf](#windsurf)
  - [Generic HTTP Client](#generic-http-client)
- [Docker Deployment](#docker-deployment)
- [JWT Authentication](#jwt-authentication)
- [Security](#security)
- [Building from Source](#building-from-source)
- [Architecture](#architecture)
- [License](#license)

---

## What It Does

This server acts as a bridge between MCP-compatible AI applications and your Pipedrive CRM account. Once configured, your AI assistant can:

- **List and search** deals, persons, organizations, leads, and pipelines
- **Filter deals** by status, owner, stage, pipeline, value range, and date range
- **Read deal notes** and custom booking details
- **Search across all entity types** at once
- **Analyze your pipeline** using built-in prompts (e.g., "group deals by stage with totals")

All 15 tools and 8 prompts are available immediately — no extra configuration needed beyond your Pipedrive API credentials.

---

## Quick Start

### Prerequisites

1. A [Pipedrive](https://www.pipedrive.com/) account
2. Your Pipedrive API token — find it at **Settings > Personal > API** or `https://<your-company>.pipedrive.com/settings/api`
3. Your Pipedrive company domain (e.g., `mycompany.pipedrive.com`)
4. An MCP-compatible client (Claude Desktop, Cursor, etc.)

### Option A — stdio (Local, Recommended)

Best for single-user desktop setups. The AI client launches the server as a subprocess and communicates over stdin/stdout. No network exposure.

1. **Download** the latest binary for your platform from [Releases](../../releases), or [build from source](#building-from-source).

2. **Make it available on your PATH** (or note the full path to the binary).

3. **Configure your MCP client** to launch the server. The exact config file location depends on your client:

   ```json
   {
     "mcpServers": {
       "pipedrive": {
         "command": "pipedrive-mcp-server",
         "args": [],
         "env": {
           "PIPEDRIVE_API_TOKEN": "<your_api_token>",
           "PIPEDRIVE_DOMAIN": "your-company.pipedrive.com"
         }
       }
     }
   }
   ```

4. **Restart your client.** The server starts automatically when the client needs it.

That's it. Ask your AI assistant something like *"List all open deals in my Pipedrive account"* and it will use the tools.

### Option B — HTTP (Remote / Docker)

Best for shared/team setups, Docker deployments, or when you want to run the server as a persistent service.

1. **Run the server** (via Docker or directly):

   ```bash
   # Using Docker Compose (recommended)
   cp .env.example .env
   # Edit .env with your PIPEDRIVE_API_TOKEN and PIPEDRIVE_DOMAIN
   docker compose up -d

   # Or run directly
   PIPEDRIVE_API_TOKEN=your_token \
   PIPEDRIVE_DOMAIN=your-company.pipedrive.com \
   MCP_TRANSPORT=http \
   MCP_HOST=0.0.0.0 \
   MCP_PORT=3000 \
   pipedrive-mcp-server
   ```

2. **Verify it's running:**

   ```bash
   curl http://localhost:3000/health
   # {"status":"ok","transport":"http"}
   ```

3. **Configure your MCP client** to connect to the HTTP endpoint:

   ```json
   {
     "mcpServers": {
       "pipedrive": {
         "type": "http",
         "url": "http://localhost:3000/mcp"
       }
     }
   }
   ```

For production deployments, add JWT authentication — see [JWT Authentication](#jwt-authentication).

---

## MCP Tools

### Deal Tools

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `get-deals` | List deals with flexible filters: search, status, owner, stage, pipeline, value range, date range | `searchTitle`, `status`, `ownerId`, `stageId`, `pipelineId`, `minValue`, `maxValue`, `daysBack`, `limit` |
| `get-deal` | Get a single deal by ID, including all custom fields | `dealId` (required) |
| `search-deals` | Quick search deals by term | `term` (required) |
| `get-deal-notes` | Get notes and booking details for a specific deal | `dealId` (required), `limit` |

### Contact Tools

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `get-persons` | List all persons (contacts) | — |
| `get-person` | Get a single person by ID, including custom fields | `personId` (required) |
| `search-persons` | Search persons by name or email | `term` (required) |

### Organization Tools

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `get-organizations` | List all organizations | — |
| `get-organization` | Get a single organization by ID, including custom fields | `organizationId` (required) |
| `search-organizations` | Search organizations by name | `term` (required) |

### Pipeline & Stage Tools

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `get-pipelines` | List all pipelines | — |
| `get-stages` | List all stages across all pipelines, enriched with `pipeline_name` | — |

### Other Tools

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `get-users` | List all users/owners (useful for filtering deals by `ownerId`) | — |
| `search-leads` | Search leads by term | `term` (required) |
| `search-all` | Cross-entity search across deals, persons, organizations, leads, and more | `term` (required), `itemTypes` |

### The `get-deals` Tool in Detail

This is the most powerful tool. Here's how the filtering works:

- **With `searchTitle`**: Uses Pipedrive's search endpoint, then applies client-side filters (owner, status, stage, pipeline, value range).
- **Without `searchTitle`**: Lists deals sorted by last activity date, filtered by `status` (default: `open`), then applies date range (`daysBack`) and value range filters client-side.
- Results are capped at 30 summarized deals per response to keep payloads manageable.

---

## MCP Prompts

Static prompts that guide the AI assistant through common analysis tasks. Available in any MCP client that supports prompts:

| Prompt | What It Asks the AI |
|--------|---------------------|
| `list-all-deals` | List all deals with title, value, status, and stage |
| `list-all-persons` | List all contacts with name, email, phone, and org |
| `list-all-pipelines` | List all pipelines with their stages |
| `analyze-deals` | Group deals by stage with total value per stage |
| `analyze-contacts` | Group contacts by organization with counts |
| `analyze-leads` | Group leads by status |
| `compare-pipelines` | Compare pipelines side by side |
| `find-high-value-deals` | Identify top-value deals with stage and association details |

---

## Configuration Reference

All configuration is via environment variables. Only two are required.

### Required

| Variable | Description | Example |
|----------|-------------|---------|
| `PIPEDRIVE_API_TOKEN` | Your Pipedrive API token | `abc123def456...` |
| `PIPEDRIVE_DOMAIN` | Your Pipedrive company domain | `mycompany.pipedrive.com` |

### Transport

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_TRANSPORT` | `stdio` | Transport mode: `stdio` or `http` |
| `MCP_PORT` | `3000` | HTTP listen port (http mode only) |
| `MCP_HOST` | `localhost` | HTTP listen host. Use `0.0.0.0` for Docker |

### JWT Authentication (http mode only)

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_JWT_SECRET` | _(none)_ | HMAC secret or PEM-encoded RSA public key. When set, all HTTP requests require a valid JWT. |
| `MCP_JWT_ALGORITHM` | `HS256` | Algorithm: `HS256`, `HS384`, `HS512`, `RS256`, `RS384`, `RS512` |
| `MCP_JWT_AUDIENCE` | _(none)_ | Expected `aud` claim |
| `MCP_JWT_ISSUER` | _(none)_ | Expected `iss` claim |

### Rate Limiting

| Variable | Default | Description |
|----------|---------|-------------|
| `PIPEDRIVE_RATE_LIMIT_MIN_INTERVAL` | `250ms` | Minimum time between Pipedrive API calls |
| `PIPEDRIVE_RATE_LIMIT_MAX_CONCURRENT` | `2` | Maximum concurrent Pipedrive API calls |

### Session Management (http mode)

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_SESSION_MAX_AGE` | `1h` | Maximum SSE session lifetime |
| `MCP_SESSION_CLEANUP_INTERVAL` | `5m` | How often to clean up expired sessions |
| `MCP_SESSION_MAX_COUNT` | `100` | Maximum concurrent sessions before rejecting new ones |

### Custom Fields

| Variable | Default | Description |
|----------|---------|-------------|
| `PIPEDRIVE_BOOKING_FIELD_KEY` | `8f4b27...7324e` | Pipedrive custom field hash key for booking details. Override this to match your instance's custom field. |

### TLS (http mode)

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_TLS_CERT` | _(none)_ | Path to TLS certificate PEM file |
| `MCP_TLS_KEY` | _(none)_ | Path to TLS private key PEM file |

---

## Client Setup Guides

### Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

**stdio mode:**
```json
{
  "mcpServers": {
    "pipedrive": {
      "command": "pipedrive-mcp-server",
      "env": {
        "PIPEDRIVE_API_TOKEN": "your_token_here",
        "PIPEDRIVE_DOMAIN": "your-company.pipedrive.com"
      }
    }
  }
}
```

If the binary isn't on your PATH, use the full path:
```json
"command": "/usr/local/bin/pipedrive-mcp-server"
```

**http mode:**
```json
{
  "mcpServers": {
    "pipedrive": {
      "type": "http",
      "url": "http://localhost:3000/mcp"
    }
  }
}
```

Restart Claude Desktop after editing. You should see a 🔨 hammer icon indicating MCP tools are available.

### Cursor

Add to your Cursor MCP settings (`Settings > MCP` or `.cursor/mcp.json` in your project):

**stdio mode:**
```json
{
  "mcpServers": {
    "pipedrive": {
      "command": "pipedrive-mcp-server",
      "env": {
        "PIPEDRIVE_API_TOKEN": "your_token_here",
        "PIPEDRIVE_DOMAIN": "your-company.pipedrive.com"
      }
    }
  }
}
```

**http mode:**
```json
{
  "mcpServers": {
    "pipedrive": {
      "url": "http://localhost:3000/mcp"
    }
  }
}
```

### Windsurf

Open `Settings > MCP Servers` and add:

**stdio mode:**
```json
{
  "mcpServers": {
    "pipedrive": {
      "command": "pipedrive-mcp-server",
      "env": {
        "PIPEDRIVE_API_TOKEN": "your_token_here",
        "PIPEDRIVE_DOMAIN": "your-company.pipedrive.com"
      }
    }
  }
}
```

### Generic HTTP Client

Any MCP-compatible client that supports the streamable HTTP transport can connect to:

```
POST http://localhost:3000/mcp    — JSON-RPC messages
GET  http://localhost:3000/mcp    — SSE stream establishment
GET  http://localhost:3000/health — Health check
```

With JWT enabled, include the token in the `Authorization` header:
```
Authorization: Bearer <jwt_token>
```

---

## Docker Deployment

### Using Docker Compose

The easiest way to run the server for a team or in production:

```bash
# 1. Copy the example env file
cp .env.example .env

# 2. Edit .env — at minimum set these two values:
#    PIPEDRIVE_API_TOKEN=your_token_here
#    PIPEDRIVE_DOMAIN=your-company.pipedrive.com

# 3. Start the server
docker compose up -d

# 4. Check it's healthy
curl http://localhost:3000/health
```

### Using Docker Directly

```bash
docker build -t pipedrive-mcp-server .

docker run -d \
  --name pipedrive-mcp \
  -e PIPEDRIVE_API_TOKEN=your_token \
  -e PIPEDRIVE_DOMAIN=your-company.pipedrive.com \
  -e MCP_TRANSPORT=http \
  -e MCP_HOST=0.0.0.0 \
  -e MCP_PORT=3000 \
  -p 3000:3000 \
  --restart unless-stopped \
  pipedrive-mcp-server
```

The resulting image is built `FROM scratch` — no shell, no package manager, ~10 MB. Runs as non-root (UID 65534). The only thing inside is the static binary and CA certificates for HTTPS calls to the Pipedrive API.

---

## JWT Authentication

When running in HTTP mode, anyone who can reach the server can use it. To restrict access, set `MCP_JWT_SECRET` to require JWT authentication on every request.

### HMAC (Symmetric) — Simplest

```bash
# Generate a random secret
SECRET=$(openssl rand -base64 32)

# Start the server with JWT enabled
docker run -d \
  -e PIPEDRIVE_API_TOKEN=your_token \
  -e PIPEDRIVE_DOMAIN=your-company.pipedrive.com \
  -e MCP_TRANSPORT=http \
  -e MCP_HOST=0.0.0.0 \
  -e MCP_JWT_SECRET="$SECRET" \
  -p 3000:3000 \
  pipedrive-mcp-server
```

Clients must include a valid JWT:
```json
{
  "mcpServers": {
    "pipedrive": {
      "type": "http",
      "url": "https://your-server.example.com/mcp",
      "headers": {
        "Authorization": "Bearer <jwt_signed_with_same_secret>"
      }
    }
  }
}
```

### RSA (Asymmetric) — For Multi-Service Setups

Set `MCP_JWT_SECRET` to a PEM-encoded RSA public key. The server will verify tokens signed with the corresponding private key. This lets you issue tokens from a separate auth service without sharing the signing key with the MCP server.

```bash
MCP_JWT_SECRET="$(cat public_key.pem)" \
MCP_JWT_ALGORITHM=RS256 \
pipedrive-mcp-server
```

### Additional JWT Validation

```bash
# Require specific audience
MCP_JWT_AUDIENCE=mcp-server

# Require specific issuer
MCP_JWT_ISSUER=my-auth-service
```

---

## Security

This server is designed to be secure by default:

| Measure | Details |
|---------|---------|
| **Error sanitization** | API tokens are stripped from all error messages before returning to clients |
| **No wildcard CORS** | Defaults to `localhost` only; no `Access-Control-Allow-Origin: *` |
| **Request body limit** | Capped at 1 MB per request |
| **Rate limiting** | Token-bucket limiter on outbound Pipedrive API calls (250 ms min interval, 2 concurrent) |
| **Session management** | TTL-based session expiry, configurable max sessions (default: 100), periodic cleanup |
| **Minimal attack surface** | 3 third-party dependencies, all well-maintained; Docker image built `FROM scratch` |
| **No token in logs** | Structured JSON logging to stderr; API tokens never appear in log output |
| **Non-root container** | Docker image runs as UID 65534 (`nobody`) |
| **No CGO** | Pure Go, statically compiled binary |

---

## Building from Source

Requirements: [Go 1.24+](https://go.dev/dl/)

```bash
# Clone
git clone https://github.com/yamirghofran/pipedrive-mcp-go.git
cd pipedrive-mcp-go

# Build
go build -o pipedrive-mcp-server ./cmd/server

# Move to PATH
sudo mv pipedrive-mcp-server /usr/local/bin/

# Verify
pipedrive-mcp-server --help
```

### Development

```bash
# Run tests
go test ./...

# Static analysis
go vet ./...

# Build optimized binary
go build -ldflags="-s -w" -o pipedrive-mcp-server ./cmd/server
```

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   MCP Client                        │
│         (Claude Desktop, Cursor, etc.)              │
└──────────────┬──────────────────┬───────────────────┘
               │                  │
         stdio │           HTTP   │  /mcp
               │                  │
┌──────────────▼──────────────────▼───────────────────┐
│              MCP Server (Go)                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │
│  │ 15 Tools │  │ 8 Prompts│  │   Middleware      │  │
│  └────┬─────┘  └──────────┘  │ • JWT Auth        │  │
│       │                      │ • CORS            │  │
│       │                      │ • Rate Limiting   │  │
│       │                      │ • Body Size Limit │  │
│       │                      └──────────────────┘  │
│       │                                              │
│  ┌────▼─────────────────────────────────────────┐   │
│  │         Pipedrive API Client                  │   │
│  │  • Token-bucket rate limiter                  │   │
│  │  • 30s request timeout                        │   │
│  │  • Error sanitization                         │   │
│  │  • Concurrent sub-requests (stages, notes)    │   │
│  └────┬─────────────────────────────────────────┘   │
└───────┼─────────────────────────────────────────────┘
        │ HTTPS
┌───────▼─────────────┐
│   Pipedrive API v1  │
│  api.pipedrive.com  │
└─────────────────────┘
```

**Project structure:**

```
├── cmd/server/main.go              # Entrypoint
├── internal/
│   ├── config/config.go            # Environment variable parsing
│   ├── pipedrive/                  # Pipedrive API client
│   │   ├── client.go               # HTTP client, rate limiter, error handling
│   │   ├── deals.go                # Deal filtering and summarization
│   │   ├── persons.go              # Person endpoints
│   │   ├── organizations.go        # Organization endpoints
│   │   ├── pipelines.go            # Pipeline + stage endpoints
│   │   ├── leads.go                # Lead search
│   │   ├── notes.go                # Note + booking details
│   │   ├── users.go                # User endpoints
│   │   └── search.go               # Cross-entity search
│   ├── server/
│   │   ├── server.go               # MCP server wiring
│   │   ├── tools.go                # 15 tool handlers
│   │   └── prompts.go              # 8 prompt handlers
│   ├── transport/
│   │   ├── stdio.go                # stdio transport
│   │   └── http.go                 # Streamable HTTP transport + sessions
│   └── middleware/
│       ├── auth.go                 # JWT authentication
│       ├── cors.go                 # CORS + body size limit
│       └── ratelimit.go            # Per-client rate limiting
├── Dockerfile                      # Multi-stage scratch build
├── docker-compose.yml
└── .env.example
```

**Dependencies (3 total):**

| Package | Purpose |
|---------|---------|
| `github.com/modelcontextprotocol/go-sdk` | Official MCP SDK — server, tools, prompts, both transports |
| `github.com/golang-jwt/jwt/v5` | JWT verification for HTTP transport authentication |
| `golang.org/x/time` | Token-bucket rate limiter |

Everything else is the Go standard library.

---

## License

MIT
