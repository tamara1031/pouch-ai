# Pouch-AI System Architecture (DDD)

## 1. Project Overview
**pouch-ai** is a self-hosted LLM proxy gateway designed for financial safety and extensibility. It uses a Domain-Driven Design (DDD) layered architecture to ensure clean separation of concerns and easy integration of new LLM providers.

## 2. Layered Architecture

### 2.1 Domain Layer (`internal/domain`)
The heart of the application, containing business logic and interfaces.
- **Key Domain**: Manages API keys, budgets, and rate limits.
- **Provider Domain**: Defines the abstraction for LLM backends (e.g., OpenAI, Anthropic).
- **Proxy Domain**: Defines the request/response flow using the **Chain of Responsibility** pattern.

### 2.2 Application Layer (`internal/app`)
Orchestrates domain entities to perform application-specific tasks.
- **KeyService**: Handles creation, validation, and usage tracking of API keys.
- **ProxyService**: Orchestrates the proxy flow through a middleware chain.

### 2.3 Infrastructure Layer (`internal/infra`)
Concrete implementations of domain interfaces and external system interactions.
- **db**: SQLite implementation of the Key Repository.
- **provider/openai**: Implementation of the OpenAI provider, including token counting and pricing.
- **proxy**: The final execution handler that performs the actual HTTP requests to LLM backends.

### 2.4 API Layer (`internal/api`)
Handles HTTP communication and presentation.
- **http/handler**: Echo handlers for keys and proxying.
- **http/middleware**: API authentication and common web middlewares.

## 3. Key Design Patterns

### Chain of Responsibility
The proxy flow is implemented as a chain of middlewares:
`RateLimit -> Mocking -> UsageTracking -> Execution`
This allows for easy injection of new features (e.g., logging, caching) without modifying existing logic.

### Provider Abstraction
By implementing the `Provider` interface, new LLM backends can be added without changing the core proxy logic.

## 4. Directory Structure

```text
pouch-ai/
├── cmd/pouch/                # Entry point
├── internal/
│   ├── api/                  # API Layer (Handlers, Middleware)
│   ├── app/                  # Application Layer (Services)
│   ├── domain/               # Domain Layer (Entities, Interfaces)
│   ├── infra/                # Infrastructure Layer (DB, Providers)
│   └── database/             # DB Connection Singleton
├── ui/                       # Astro Frontend
└── data/                     # SQLite Database
```

## 5. Technology Stack
- **Backend**: Go (Echo Framework)
- **Frontend**: Astro + TailwindCSS + DaisyUI
- **Database**: SQLite (CGO-free)
- **Token Counting**: tiktoken-go
