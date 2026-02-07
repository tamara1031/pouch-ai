# pouch-ai

**pouch-ai** is a self-hosted, single-binary LLM proxy gateway designed for personal use and homelabs.
Its primary goal is **Financial Safety** when using pay-per-token APIs (like OpenAI) by enforcing a strict budget even with streaming responses.

## Features

- **Financial Safety**: Enforces a strict budget by reserving estimated costs before requests and refunding unused amounts after completion.
- **Single Binary**: The Go backend embeds the compiled Astro frontend and manages the SQLite database, resulting in a single executable key for easy distribution.
- **CGO-Free**: Uses `modernc.org/sqlite` to ensure the binary is static and cross-compatible (e.g., Linux amd64/arm64) without libc dependencies.
- **OpenAI Compatible**: Works with existing OpenAI clients by changing the base URL.
- **Provider Support**: Currently supports **OpenAI** and a **Mock** provider for testing/development.
- **Secure**: API keys are stored encrypted (AES-256-GCM).

## Getting Started

### Prerequisites

- **Go 1.25+** (Required for the latest Go features)
- **Node.js 24+** (For building the frontend)

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/tamara1031/pouch-ai.git
   cd pouch-ai
   ```

2. **Build the Frontend**:
   ```bash
   cd ui
   npm install
   npm run build
   cd ..
   ```

3. **Build the Backend**:
   ```bash
   go build -o pouch cmd/pouch/main.go
   ```

### Usage

Run the server with default settings:

```bash
./pouch -port 8080 -data ./data
```

- **Dashboard**: Visit `http://localhost:8080` to configure settings and monitor usage.
- **API Endpoint**: `http://localhost:8080/v1/chat/completions`

### Configuration

#### CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-port` | Port to run the server on | `8080` |
| `-target` | Target OpenAI API Base URL | `https://api.openai.com` |
| `-data` | Directory to store the SQLite database | `./data` |
| `-cors-origins` | Comma-separated list of allowed CORS origins | `*` |

#### Environment Variables

The application relies on environment variables for provider authentication:

- `OPENAI_API_KEY`: Required when using the OpenAI provider.

## Architecture

For a deep dive into the system design, see [ARCHITECTURE.md](ARCHITECTURE.md).

## Development

For detailed development instructions, including how to run the project locally and run tests, see [DEVELOPMENT.md](DEVELOPMENT.md).

The project consists of:
- **Backend**: Go (Echo, SQLite, DDD Architecture)
- **Frontend**: Astro + DaisyUI + TailwindCSS

See `ARCHITECTURE.md` for the directory structure and module details.
