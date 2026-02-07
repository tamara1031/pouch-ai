# Development Guide

This guide covers how to set up the development environment for **pouch-ai**.

## Prerequisites

- **Go**: Version 1.25 or higher.
- **Node.js**: Version 24 or higher.
- **Package Manager**: `npm` or `pnpm`.

## Project Structure

- `cmd/pouch/`: Entry point for the Go backend.
- `internal/`: Private application code (DDD structure).
- `frontend/`: Astro frontend application.

## Running Locally

### 1. Backend (Go)

The backend serves the API and the static frontend files (from `frontend/dist`).

```bash
# Run the backend (rebuilds on change if using air, otherwise manual restart)
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

**Note:** If the frontend code makes relative API calls (e.g., `/v1/...`), they will fail on port 4321 unless you configure a proxy in `astro.config.mjs` or point the API client to `http://localhost:8080`.

For a full integration test, build the frontend and run the backend:

```bash
# Build frontend
cd frontend && npm run build && cd ..

# Run backend (serves the built frontend)
go run cmd/pouch/main.go
```

Then visit `http://localhost:8080`.

## Building for Production

To create the single binary:

1. **Build Frontend**:
   ```bash
   cd frontend
   npm install
   npm run build
   ```

2. **Build Backend**:
   ```bash
   # From project root
   go build -o pouch cmd/pouch/main.go
   ```

The resulting `pouch` binary contains everything needed to run.

## Testing

### Backend Tests

Run all Go tests:

```bash
go test ./...
```

To run with race detection:

```bash
go test -race ./...
```

### Frontend Tests

Currently, the frontend does not have a dedicated test suite (e.g., Vitest/Jest). Development relies on manual verification.
