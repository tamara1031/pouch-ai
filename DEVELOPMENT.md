# Development Guide

This guide covers how to set up the development environment for **pouch-ai** and contribute to its codebase.

## Prerequisites

- **Go**: Version 1.25 or higher.
- **Node.js**: Version 24 or higher.
- **Package Manager**: `npm` or `pnpm`.

## Project Structure

- `cmd/pouch/`: Entry point for the Go backend.
- `backend/`: Private application code (DDD structure).
- `frontend/`: Astro frontend application.

## Running Locally

### 1. Backend (Go)

The backend serves the API and the static frontend files (from `frontend/dist`).

```bash
# Run the backend
go run cmd/pouch/main.go
```

By default, it listens on `http://localhost:8080`.

### 2. Frontend (Astro)

For frontend development with hot-reload:

```bash
cd frontend
npm install
npm run dev
```

The frontend dev server runs on `http://localhost:4321`.

### 3. Integrated Development

For a full integration test:

```bash
# Build frontend
cd frontend && npm run build && cd ..

# Run backend (serves the built frontend)
go run cmd/pouch/main.go
```

## Logging System

Pouch AI uses a structured logging system based on Go's `log/slog`.

### Guidelines
- Always use the centralized logger from `backend/util/logger`.
- Avoid using `fmt.Printf` or `log.Printf`.
- Use contextual data for better observability.

```go
import "pouch-ai/backend/util/logger"

// Standard logging
logger.L.Info("operation successful", "key_id", id)

// Error logging
logger.L.Error("failed to process request", "error", err, "plugin", pluginID)
```

## Adding New Plugins

Pouch AI's plugin system is decentralized, supporting both built-in components and external plugins.

### 1. Providers (Built-in Only)
Currently, new LLM providers must be compiled into the binary.

1. Create a new package in `backend/plugins/providers/` (e.g., `anthropic`).
2. Implement the `domain.Provider` and `domain.ProviderBuilder` interfaces.
3. Export your builder via the `GetBuilders()` function in `backend/plugins/providers/registry.go`.

### 2. Middlewares (Built-in & External)

#### Option A: Built-in Middleware
1. Create a new package in `backend/plugins/middlewares/`.
2. Implement the `domain.Middleware` interface.
3. Define the middleware metadata (`domain.PluginInfo`) and a factory function.
4. Register the middleware in the `GetBuiltins()` function in `backend/plugins/middlewares/registry.go`.

#### Option B: External Plugin (`.so`)
You can load middlewares dynamically at runtime using Go plugins.

1. Create a main package for your plugin.
2. Implement the `domain.Middleware` interface.
3. Export a function `GetPlugin() domain.MiddlewareEntry`.
4. Build the plugin: `go build -buildmode=plugin -o myplugin.so main.go`.
5. Place the `.so` file in the configured plugins directory (default: `backend/plugins/middlewares`).

**Example External Plugin:**
```go
package main

import "pouch-ai/backend/domain"

func GetPlugin() domain.MiddlewareEntry {
    return domain.MiddlewareEntry{
        Info:    domain.PluginInfo{ID: "my-plugin", ...},
        Factory: NewMyMiddleware,
    }
}
```

## Testing

### Backend Tests

Run all Go tests:

```bash
go test ./...
```

Run tests for a specific package:

```bash
go test ./backend/service/...
```

### Frontend Verification

The frontend follows a modular architecture. When making changes, verify:
1. **API Client & Types**: Ensure `src/api/api.ts` and `src/types.ts` are updated together for any API change.
2. **Component Reusability**: Use generic components from `src/components/ui` for common UI patterns.
3. **State Management**: Use Preact Signals in `src/hooks/` for shared state (like modal visibility) to avoid event-based complexity.
4. **Build Status**: Always run `npm run build` in the `frontend` directory to catch syntax or type errors.
