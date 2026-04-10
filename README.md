
# Go Code Execution Engine

A robust, distributed, and secure code execution engine built with **Go**, **Docker**, and **Redis**. This system mimics the core backend of platforms like LeetCode or Codeforces, capable of running untrusted user code in isolated environments with strict resource constraints.

##  Features

* **Distributed Architecture:** Decouples submission ingestion (API) from execution (Worker) using a Redis Job Queue.
* **Secure Isolation:** Runs every job in a transient Docker container with no network access.
* **Resource Hardening:**
* **Time Limit:** Enforced via Context Timeouts & Linux `timeout` commands.
* **Memory Limit:** Docker `Memory` and `MemorySwap` constraints to prevent OOM.
* **Process Limit:** `PidsLimit` to prevent Fork Bombs.
* **CPU Throttling:** `NanoCPUs` to prevent CPU hogging/overheating.


* **Output Management:** Handles `stdin` injection and enforces strict `stdout`/`stderr` size limits to prevent log flooding attacks.
* **Scalable:** Worker pool pattern allows processing multiple jobs concurrently.

##  Tech Stack

* **Language:** Go (Golang) 1.22+
* **Orchestration:** Docker & Docker SDK
* **Queue/Broker:** Redis
* **Database:** PostgreSQL (Planned)
* **Infrastructure:** Docker Compose

## Getting Started

### Prerequisites

* Docker & Docker Compose installed and running.
* Go 1.22+ (for local development).

### Method 1: Run via Docker Compose (Recommended)

This spins up the API, Worker, Redis, and Database instantly.

```bash
# 1. Clone the repository
git clone https://github.com/siddhesh241o/code-execution-engine.git
cd code-execution-engine

# 2. Start the services
docker-compose up --build

```

### Method 2: Manual Run (Local Dev)

If you want to debug Go code without rebuilding containers:

1. **Start Infrastructure:**
```bash
docker-compose up -d redis

```


2. **Start Worker:**
```bash
go run cmd/worker/main.go

```


3. **Start API:**
```bash
go run cmd/api/main.go

```



## 🔌 API Documentation

### 1. Submit Code

**POST** `/submit`

Enqueues a job for execution.

**Request Body:**

```json
{
  "language": "cpp",
  "code": "#include <iostream>\nint main() { std::cout << \"Hello World\"; return 0; }",
  "input": ""
}

```

**Response:**

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "Queued"
}

```

---

### 2. Get Result

**GET** `/result/{id}` (or `/result?id={id}` depending on version)

Fetches the status and output of a submission.

**Response (Pending):**

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "Processing"
}

```

**Response (Completed):**

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "Accepted",
  "stdout": "Hello World",
  "stderr": "",
  "execution_time": "0.002s"
}

```

## ⚙️ Configuration

The application is configured via Environment Variables. See `internal/config/config.go`.

| Variable | Default | Description |
| --- | --- | --- |
| `REDIS_ADDR` | `localhost:6379` | Address of the Redis instance |
| `WORKER_COUNT` | `4` | Number of concurrent workers |
| `DOCKER_MEM_LIMIT` | `256MB` | Max RAM per user container |
| `DOCKER_CPU_LIMIT` | `0.5` | Max CPU cores per user container |

## 🛡️ Security Mechanisms

| Attack Vector | Defense Mechanism |
| --- | --- |
| **Infinite Loop** | `context.WithTimeout` (Go) + `timeout` command (Linux) |
| **Fork Bomb** | `PidsLimit` set to 20 processes |
| **Memory Bomb** | Docker CGroup memory limits (128MB) |
| **Log Flooding** | `LimitedWriter` truncates output > 10KB + Docker `max-file` logging config |
| **Network Access** | `NetworkMode: "none"` (No Internet Access) |
