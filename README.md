# mcp-jenkins-adapter

A containerized MCP client-server stack that bridges Jenkins with AI LLMs (OpenAI / Claude). Trigger jobs, check build status, fetch logs, and get AI-powered build analysis — all from a conversational interface.

> **PoC project.** Built for local dev and experimentation, not production use.

---

## How it works

```
You → MCP Client → MCP Server → Jenkins
              ↕
         OpenAI / Claude
```

- **Jenkins** runs your CI jobs and exposes a REST API
- **MCP Server** wraps Jenkins into AI-callable tools via SSE
- **MCP Client** connects to an LLM and routes your prompts to the right tool

---

## Quick start

### 1. Configure environment

```bash
cp .env.example .env
# Fill in Jenkins credentials and your LLM API key
```

### 2. Start Jenkins and extract its API token

```bash
podman-compose up --build -d jenkins

# Wait ~30s for Jenkins to initialize, then:
mkdir -p ./jenkins-secrets
touch ./jenkins-secrets/mcp-user.token
podman-compose exec jenkins cat /var/jenkins_home/secrets/mcp-user.token > ./jenkins-secrets/mcp-user.token
```

### 3. Start everything

```bash
podman-compose up --build -d
```

### 4. Chat with Jenkins

```bash
podman-compose exec mcp-client mcp-client
```

```
> run demo-job
> check status of demo-job build 1
> show me logs for demo-job build 1
> why did demo-job build 1 fail?
```

Or use `make quickstart` to do steps 1–3 in one shot.

---

## Available tools

| Tool | What it does |
|---|---|
| `trigger_job` | Trigger a Jenkins job by name |
| `get_build_status` | Get the status of a build |
| `get_console_log` | Fetch console output for a build |
| `analyze_logs` | Fetch logs + send to LLM for diagnosis |

---

## LLM providers

Set `LLM_PROVIDER` in your `.env`:

| Value | Key needed |
|---|---|
| `openai` | `OPENAI_API_KEY` |
| `claude` | `CLAUDE_API_KEY` |

---

## Requirements

- Podman or Docker + Compose
- An OpenAI or Claude API key

---

## Useful commands

```bash
make up           # Start all services
make down         # Stop all services
make logs         # Follow logs
make token        # Re-extract Jenkins API token
make run-client   # Open interactive MCP client session
make clean        # Stop and wipe volumes
```

---

## License

GNU GPL v3.0 — see [LICENSE](LICENSE).
