# API & Interface Definition

This document defines the contract between the **Frontend (UI)** and **Backend (Go)** to enable independent, parallel development by multiple agents.

## 1. REST API Contract
Base URL: `/v1` (Relative to the server root)

### 1.1 Budget Management
**Get Current Budget**
- **Endpoint**: `GET /v1/stats/budget`
- **Response**: `200 OK`
    ```json
    {
      "budget": 12.50  // Float: Current balance in USD
    }
    ```
- **Error**: `500 Internal Server Error`

**Update Budget (Top-up/Set)**
- **Endpoint**: `POST /v1/stats/budget`
- **Headers**: `Content-Type: application/json`
- **Request**:
    ```json
    {
      "amount": 20.00  // Float: New balance to set (USD)
    }
    ```
- **Response**: `200 OK`
    ```json
    "updated"
    ```
- **Error**: `400 Bad Request` (Invalid JSON), `500 Internal Server Error`

### 1.2 OpenAI Proxy
**Chat Completions**
- **Endpoint**: `POST /v1/chat/completions`
- **Behavior**: Standard OpenAI API Proxy.
- **Constraints**:
    - If `Budget < Estimated Cost`, returns `402 Payment Required`.
    - If `Rate Limit Exceeded`, returns `429 Too Many Requests`.
- **Response**: Standard OpenAI Stream or JSON response.

---

## 2. Build & Asset Interface

To ensure the Frontend and Backend agents can work without stepping on each other's build processes:

### 2.1 Frontend Output
- **Responsibility**: Frontend Agent
- **Output Directory**: `ui/dist/`
- **Requirement**: The frontend build process (e.g., `npm run build` in `ui/`) **MUST** generate static HTML/JS/CSS files into the `ui/dist/` directory.
- **Entry Point**: `ui/dist/index.html` must be the main entry point.

### 2.2 Backend Embedding
- **Responsibility**: Backend Agent
- **Mechanism**: The Go executable expects to find assets in `ui/dist` (during development) or embedded in the binary (production).
- **Contract**: The backend will serve any file found in `ui/dist` at the root path `/`.

---

## 3. Development Workflow

### Frontend Agent
1. Work inside `ui/` directory.
2. Mock API calls to `/v1/*` or use the running backend.
3. run `npm run build` to update artifacts for the backend to serve.

### Backend Agent
1. Work inside `cmd/`, `internal/`.
2. Do **NOT** modify files in `ui/` (except `embed.go`).
3. Assume `ui/dist` contains valid static files.
