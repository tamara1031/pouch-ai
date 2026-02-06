# pouch-ai System Architecture

## 1. Project Overview
**pouch-ai** is a self-hosted, single-binary LLM proxy gateway designed for personal use and homelabs.
Its primary goal is **Financial Safety** when using pay-per-token APIs (like OpenAI) by enforcing a strict budget even with streaming responses.

### 2. Core Concepts
- **Deposit / Refund Model**:
    - **Reserve (Deposit)**: Before a request is proxied, specificed strict budget is "reserved" (deducted) based on `max_tokens`.
    - **Refund**: After the request completes, the actual cost is calculated, and the difference (unused budget) is "refunded" to the user's balance.
- **Single Binary**: The Go backend embeds the compiled Astro frontend and manages the SQLite database, resulting in a single executable key for easy distribution.
- **CGO-Free**: Uses `modernc.org/sqlite` to ensure the binary is static and cross-compatible (e.g., Linux amd64/arm64) without libc dependencies.

## 3. Technology Stack
- **Language**: Go 1.22+
- **HTTP Framework**: [Echo v4](https://echo.labstack.com/)
- **Database**: SQLite (via [modernc.org/sqlite](https://gitlab.com/cznic/sqlite))
- **Frontend**: [Astro](https://astro.build/) + [DaisyUI](https://daisyui.com/) + TailwindCSS
- **Token Counting**: [tiktoken-go](https://github.com/pkoukk/tiktoken-go)

## 4. Directory Structure

```text
pouch-ai/
├── cmd/
│   └── pouch/
│       └── main.go           # Application Entry Point. Embeds UI assets.
├── internal/
│   ├── budget/               # Budget Manager. Handles Reserve/Refund logic and DB transactions.
│   ├── database/             # SQLite initialization and schema migrations.
│   ├── proxy/                # Reverse proxy logic, cost estimation, and response interception.
│   ├── security/             # Crypto utilities (AES-256-GCM, PBKDF2).
│   ├── server/               # Echo server setup, middleware, and route definitions.
│   └── token/                # Token counting utilities (tiktoken wrapper).
├── ui/
│   ├── src/                  # Astro frontend source code.
│   ├── dist/                 # Compiled frontend assets (generated).
│   ├── embed.go              # Go package to embed the `dist/` directory.
│   └── package.json          # Node.js dependencies.
├── assets/
│   └── pricing.json          # Embedded model pricing definitions.
└── data/                     # (Runtime) Directory for SQLite DB (pouch.db).
```

## 5. Module Details

### 5.1 Budget (`internal/budget`)
- **Manager**: Controls the `system_config` table in SQLite to track `budget_usd`.
- **Concurrency**: Uses `sync.Mutex` alongside SQLite transactions (`BEGIN IMMEDIATE` implied by locking strategy or driver) to prevent race conditions during concurrent requests.

### 5.2 Proxy (`internal/proxy`)
- **Handler**:
    1. **Estimation**: Calculates `max_cost = (input_tokens + max_tokens) * price`.
    2. **Reservation**: Calls `budget.Reserve(max_cost)`. Fails with 402 if insufficient.
    3. **Proxying**: Uses `httputil.ReverseProxy` to forward to OpenAI.
    4. **Interception**: Wraps `http.ResponseWriter` to capture response body (even for streams).
    5. **Refund**: Calculates `actual_cost` and calls `budget.Refund(max_cost - actual_cost)`.
- **Pricing**: Loaded from `internal/proxy/pricing.json`.

### 5.3 Security (`internal/security`)
- **Design**: "Zero-Dependency Secret Management".
- **Master Password**: Hashed with PBKDF2/Argon2.
- **API Keys**: Stored encrypted (AES-256-GCM) in `credentials` table. Decrypted only in memory.

### 5.4 Token (`internal/token`)
- Wraps `tiktoken-go` to provide accurate token counts for standard OpenAI models.

## 6. Build & Development

### Prerequisites
- Go 1.22+
- Node.js (v20+)

### Build Workflow
The project is built in two stages:
1. **Frontend**:
    ```bash
    cd ui
    npm install
    npm run build  # Output to ui/dist
    ```
2. **Backend**:
    ```bash
    # Root of repo
    go build -o pouch cmd/pouch/main.go
    ```

### Running Locally
```bash
./pouch -port 8080 -data ./data
```
Visit `http://localhost:8080` for the Dashboard.
API Endpoint: `http://localhost:8080/v1/chat/completions`.

## 7. Future Considerations for Agents
- **Testing**: When adding features, ensure `internal/budget` logic remains invariant. Money must not disappear.
- **Migration**: UI is currently embedded via `ui/embed.go`. If moving files, ensure `//go:embed` directives are updated.
- **Extensibility**: To add new providers (Anthropic, Gemini), implement a standardized `Provider` interface in `internal/proxy`.
