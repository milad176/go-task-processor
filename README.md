# 🚀 go-task-processor

A lightweight concurrent task processing system built in Go.  
Implements a simplified distributed job queue with worker pools, atomic job claiming, recovery mechanisms, and observability.

---

# 🎯 Why this project exists

This project was built to demonstrate core backend and distributed systems concepts using only Go and SQLite, without external infrastructure like Redis or RabbitMQ.

It focuses on simulating how real-world background job systems behave internally:

- concurrent workers safely processing jobs
- atomic scheduling to avoid duplicate work
- failure recovery for crashed workers
- retry and backoff strategies
- system observability and metrics

---

# ⚙️ Features

- 🧵 Worker pool using goroutines
- ⚡ Atomic job claiming (prevents double processing)
- 📊 Priority-based job scheduling
- 🔁 Retry mechanism with exponential backoff
- 💥 Stuck job recovery (crash-safe processing)
- 🛑 Graceful shutdown with WaitGroup coordination
- 📈 Basic metrics (processed / failed / recovered jobs)
- 🗄 SQLite persistent storage
- 🧾 Structured logging with execution timing

---

# 🧠 Architecture Overview

```text
Client
  │
  ▼
HTTP API Server
  │
  ▼
SQLite Job Table
  │
  ▼
┌───────────────────────────┐
│      Worker Pool          │
│  (Goroutines, N workers)  │
└───────────────────────────┘
  │
  ▼
Job Execution Engine
  │
  ├── SUCCESS → DONE
  ├── FAILURE → RETRY (backoff)
  └── STUCK → RECOVERY → PENDING
  │
  ▼
Metrics + Logging System
```

---

🔄 Job Lifecycle

Each job moves through a controlled lifecycle:

PENDING
   ↓ (atomic claim)
PROCESSING
   ↓
DONE / FAILED


If a worker crashes:

PROCESSING (stuck)
      ↓
Recovery Service
      ↓
PENDING (re-queued)
      ↓
Reprocessed by another worker

---

⚙️ Core Design Decisions

1. SQLite instead of external queue

This project intentionally avoids Redis/RabbitMQ to demonstrate:
	•	concurrency control at application level
	•	transactional job claiming
	•	locking behavior in real systems


2. Atomic job claiming

Jobs are claimed using a transaction-based approach:
	•	prevents multiple workers processing the same job
	•	ensures safe concurrency without distributed locks


3. Stuck job recovery

A background recovery service periodically:
	•	finds jobs in processing state
	•	checks claimed_at timestamp
	•	resets stale jobs back to pending

This simulates crash recovery in production systems.


4. Graceful shutdown

System shutdown ensures:
	•	no new jobs are claimed
	•	active jobs are allowed to finish
	•	workers exit cleanly using WaitGroup synchronization


5. Metrics tracking

The system tracks:
	•	processed jobs
	•	failed jobs
	•	recovered jobs
  
---

📡 API Endpoints

Create Job:
POST /jobs

Example payload:
{
  "type": "print",
  "priority": 1,
  "payload": {
    "message": "hello world"
  }
}

Metrics:
GET /metrics

Example response:
{
  "Processed": 12,
  "Failed": 1,
  "Recovered": 2
}

---

▶️ How to Run

go run cmd/server/main.go

Server will start at:
http://localhost:8080

---

🧪 Example Output

[WORKER-1] processing job=123 type=payment priority=3
[WORKER-1] processing payment for order=ORD-1001
Payment completed
[WORKER-1] finished job=123 status=done duration=7.2s

---

🧩 What this project demonstrates

This project shows understanding of:
	•	concurrency in Go (goroutines + worker pools)
	•	race condition prevention strategies
	•	transactional job processing
	•	failure recovery patterns
	•	backend system design principles
	•	observability basics

---

🚀 Future Improvements

Then stop after:

	•	Redis or RabbitMQ integration (true distributed queue)
	•	Prometheus metrics exporter
	•	OpenTelemetry tracing
	•	Web dashboard for job monitoring
	•	Job prioritization improvements (heap-based scheduler)
	•	Horizontal scaling (multi-node workers)

---

🏁 Summary
  
This project simulates a production-style background job system with a focus on correctness, concurrency safety, and system resilience.

It demonstrates how reliable job processing systems behave under failures, concurrency pressure, and shutdown scenarios.
