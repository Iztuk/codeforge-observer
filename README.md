# CodeForge Observer

An OpenAPI contract auditing proxy for real-time API validation and observability.

CodeForge Observer sits between your client and API, intercepting every request and response. It validates your traffic against OpenAPI contracts and automatically captures any contract violations—giving you immediate insight into API behavior without modifying your code.

Perfect for:

- **Development** — Catch API contract violations early
- **Observability** — Track real API behavior vs documented contracts
- **Feedback** — Identify where your API deviates from spec

## Installation

**Linux x86_64:**

```bash
curl -L https://github.com/Iztuk/codeforge-observer/releases/latest/download/cf-observer-linux-amd64 -o cf-observer
chmod +x cf-observer
sudo mv cf-observer /usr/local/bin/
```

**macOS Intel:**

```bash
curl -L https://github.com/Iztuk/codeforge-observer/releases/latest/download/cf-observer-darwin-amd64 -o cf-observer
chmod +x cf-observer
sudo mv cf-observer /usr/local/bin/
```

**macOS Apple Silicon:**

```bash
curl -L https://github.com/Iztuk/codeforge-observer/releases/latest/download/cf-observer-darwin-arm64 -o cf-observer
chmod +x cf-observer
sudo mv cf-observer /usr/local/bin/
```

**Windows:**

```powershell
Invoke-WebRequest -Uri "https://github.com/Iztuk/codeforge-observer/releases/latest/download/cf-observer-windows-amd64.exe" -OutFile cf-observer.exe
```

## Usage

**Start the daemon:**

```bash
cf-observer start
```

**Add an API to monitor:**

```bash
cf-observer host -action add \
  -name my-api \
  -upstream http://localhost:3000 \
  -contract ./openapi.json
```

**List configured hosts:**

```bash
cf-observer host -action list
```

**Remove a host:**

```bash
cf-observer host -action remove -name my-api
```

**Stop the daemon:**

```bash
cf-observer stop
```

## How It Works

1. The daemon listens on `http://localhost:8080`
2. Configure hosts with their upstream URL and OpenAPI contract
3. Route requests through the proxy
4. Observer validates requests/responses against the contract
5. Findings are stored in SQLite at `~/.config/cf-observer/observer.db`

## Documentation

Full documentation coming soon

## Status

v0.1 (Early Release) — Feedback welcome
