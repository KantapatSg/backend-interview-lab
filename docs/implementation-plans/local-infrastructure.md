# Local Infrastructure Implementation Plan

This is the executable contract for the first vertical slice. It uses
PostgreSQL as durable source of truth, Kafka for asynchronous delivery, and
Redis only as a disposable read accelerator.

## Scope

```text
POST /jobs -> PostgreSQL transaction (jobs + outbox_events)
  -> outbox publisher -> Kafka jobs.events.v1
  -> consumer group job-worker
  -> processed_events receipt + job state update
GET /jobs/{id} -> Redis cache-aside -> PostgreSQL fallback
```

The MVP runs the outbox publisher and Kafka consumer in one `worker` binary so
the lifecycle is easy to demonstrate. The master plan keeps them as separable
service responsibilities for a later scaling lab.

The API returns `202 Accepted`: the command is durably accepted, while the
eventual worker state is observed through the query endpoint.

## Run locally

```powershell
docker compose up -d postgres migrate
docker compose --profile eventing up -d
docker compose --profile cache up -d redis
docker compose --profile full up --build -d
.\scripts\infra-status.ps1
```

Normal shutdown keeps data:

```powershell
docker compose -p backend-interview-lab down --remove-orphans
```

Data reset is explicit and project-scoped:

```powershell
.\scripts\reset-local-data.ps1 -ConfirmDataLoss
```

## Service contracts

| Service | Contract | Failure behavior |
|---|---|---|
| PostgreSQL | durable jobs, outbox and processed-event receipts | no accepted command without DB commit |
| Kafka | `jobs.events.v1`, `jobs.retry.v1`, `jobs.dlq.v1` | publisher retries from outbox |
| Redis | `job-view:v1:{tenant}:{id}`, short TTL | query falls back to PostgreSQL |

Kafka uses a single-node KRaft topology only for local learning. Containers use
`kafka:9092`; host tools use `localhost:29092`.

## Correctness invariants

1. Job insert and outbox insert commit in the same PostgreSQL transaction.
2. An outbox event is marked published only after Kafka acknowledges it. A
   crash may duplicate an event, but cannot lose it.
3. The consumer writes `processed_events` and applies state in one transaction
   before committing the Kafka offset. Redelivery is a no-op.
4. The Kafka message key is `job_id`, preserving per-job ordering within a
   partition; it is not global ordering.
5. Cache data can be stale or deleted without changing business correctness.

## Verification matrix

| Check | Command/experiment | Expected proof |
|---|---|---|
| Go unit | `go test ./...` | domain and cache fallback tests pass |
| Compose syntax | `docker compose config` | services resolve |
| PostgreSQL | create job, inspect `jobs` + `outbox_events` | rows appear atomically |
| Kafka | restart worker and redeliver | receipt PK prevents duplicate side effect |
| Redis | stop Redis, query job | PostgreSQL still serves response |
| E2E | POST then bounded-poll GET | eventual worker state is observable |

## Graphify/MyMap nodes

```text
Compose.Postgres -> Postgres.Migrations -> Postgres.Jobs
Postgres.Jobs -> Postgres.OutboxEvents -> Kafka.JobsEventsV1
Kafka.JobsEventsV1 -> Kafka.JobConsumer -> Postgres.ProcessedEvents
Kafka.JobConsumer -> Postgres.Jobs
HTTP.GetJob -> Cache.JobView -> Postgres.Jobs
```

Implementation files are under `platform/cmd`, `platform/internal`,
`platform/migrations/001_init.sql` and `platform/web`. The `migrate` Compose
service runs the versioned SQL explicitly; PostgreSQL initialization hooks are
not used as the migration mechanism.
