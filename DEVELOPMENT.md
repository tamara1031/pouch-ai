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

Pouch AI's plugin system is decentralized, making it easy to add new LLM providers or middlewares.

### Adding a Provider
1. Create a new package in `backend/plugins/providers/` (e.g., `anthropic`).
2. Implement the `domain.Provider` and `domain.ProviderBuilder` interfaces.
3. Add an `init()` function with a call to `Register()` to decentralized registration.
4. Export the builder via a `registry.go` in your package.

### Adding a Middleware
1. Create a new package in `backend/plugins/middlewares/`.
2. Implement the `domain.Middleware` interface.
3. Define the middleware metadata (ID and schema).
4. Register the middleware in the decentralized registry.

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

The frontend currently relies on manual verification. When making changes to shared logic or UI, verify:
1. **API Client**: Ensure `src/api/api.ts` correctly maps backend responses.
2. **Form Consistency**: Verify that `KeyForm.tsx` handles both creation and editing correctly.
3. **Build Status**: Always run `npm run build` in the `frontend` directory to catch syntax or type errors.
