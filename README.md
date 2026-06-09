# Go Code Execution Engine

A robust, distributed, and secure code execution engine built with **Go**, **Docker**, and **Redis**. This system mimics the core backend of platforms like LeetCode or Codeforces, capable of running untrusted user code in isolated environments with strict resource constraints.

## 🚀 Key Features

*   **Dual Execution Modes:**
    *   **Local (Docker):** Runs jobs in transient Docker containers with strict CGroups resource limits.
    *   **Remote (GitHub Actions):** Dispatches execution to GitHub Actions for serverless, zero-maintenance execution (useful for scaling or bypassing local Docker requirements).
*   **Flexible Storage:** Supports both **Redis** (for persistence and distribution) and **In-Memory** (for simple local development) storage modes.
*   **Integrated Frontend:** Includes a modern React playground with Monaco Editor support.
*   **Resource Hardening (Local Mode):**
    *   **Time Limit:** Enforced via `context.WithTimeout` and Linux `timeout`.
    *   **Memory Limit:** Docker `Memory` and `MemorySwap` constraints (128MB-256MB).
    *   **Process Limit:** `PidsLimit` (20-32) to prevent Fork Bombs.
    *   **CPU Throttling:** `NanoCPUs` (0.5 cores) to prevent CPU hogging.
    *   **Network Isolation:** `NetworkMode: "none"` ensures no internet access.
*   **Observability:** Built-in Prometheus metrics (`/metrics`) for tracking execution counts, latency, and status codes.

## 🛠️ Tech Stack

### Backend
*   **Language:** Go (Golang) 1.25.0
*   **Orchestration:** Docker SDK & GitHub Actions API
*   **Queue/Broker:** Redis
*   **Observability:** Prometheus

### Frontend
*   **Framework:** React 19 (TypeScript)
*   **Build Tool:** Vite 8
*   **Styling:** TailwindCSS 4
*   **Editor:** Monaco Editor

## 📁 Project Structure

```text
.
├── cmd/
│   ├── server/          # API Server (Ingestion & Dispatcher)
│   └── worker/          # Local Execution Worker (Docker)
├── internal/
│   ├── api/             # HTTP Handlers & Middleware
│   ├── domain/          # Core Interfaces & Types
│   ├── infrastructure/  # Redis, Memory, & GHA implementations
│   ├── observability/   # Prometheus metrics
│   └── runner/          # Docker execution logic & Language configs
├── frontend/            # React-based playground
└── .github/workflows/   # Workflow for GHA execution mode
```

## ⚙️ Configuration

The application is configured via Environment Variables (or a `.env` file).

| Variable | Description |
| :--- | :--- |
| `EXECUTION_MODE` | `local` (Docker Worker) or `gha` (GitHub Actions) |
| `STORE_MODE` | `redis` or `memory` |
| `REDIS_ADDR` | Address of the Redis instance (e.g., `localhost:6379`) |
| `HTTP_PORT` | Port for the API server (default `5005`) |
| `RESULT_TTL` | How long to keep results (e.g., `10m`, `1h`) |
| `FRONTEND_SHARED_SECRET` | Secret for `X-Frontend-Secret` header |
| `FETCH_SECRET` | Secret for GHA to fetch job details |
| `CALLBACK_SECRET` | Secret for GHA to post results back |
| `GITHUB_TOKEN` | PAT with repo scope (required for `gha` mode) |
| `GITHUB_OWNER/REPO` | Target repository for GHA execution |

## 🔌 API Documentation

### 1. Execute Code
**POST** `/api/execute`  
*Requires `X-Frontend-Secret` header.*

**Request Body:**
```json
{
  "language": "python",
  "code": "print(int(input()) * 2)",
  "input": "21"
}
```

**Response:**
```json
{
  "job_id": "uuid-v4-string",
  "status": "Queued"
}
```

### 2. Get Result
**GET** `/api/result/{id}`

**Response (Completed):**
```json
{
  "id": "uuid-v4-string",
  "stdout": "42\n",
  "stderr": "",
  "status": "Successfully Executed",
  "time_ms": 25,
  "memory_kb": 3400
}
```

### Supported Languages
*   `python` (Python 3.12+)
*   `c++` (G++ 13+, C++17)
*   `java` (OpenJDK 17)
*   `javascript` (Node.js 20)

## 🛡️ Security Mechanisms

| Attack Vector | Defense Mechanism |
| :--- | :--- |
| **Infinite Loop** | Execution timeouts (5s default) |
| **Fork Bomb** | Process ID limits (PidsLimit=32) |
| **Memory Bomb** | Hard memory limits + OOM Killer awareness |
| **Log Flooding** | Output truncation (>10KB) |
| **Network Access** | Full network isolation in Docker containers |

## 🏃 Getting Started

### Using Docker Compose
```bash
docker-compose up --build
```

### Manual Development
1. Start Redis.
2. Configure `.env` file based on the [Configuration](#️-configuration) section.
3. Start Server: `go run cmd/server/main.go`
4. Start Worker (if in `local` mode): `go run cmd/worker/main.go`
5. Start Frontend: `cd frontend/code-execution-engine && npm install && npm run dev`
