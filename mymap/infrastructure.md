# Infrastructure MyMap

## Definition

Docker Compose gives the labs reproducible PostgreSQL, Kafka and Redis
dependencies; the application still owns retry, readiness and correctness.

## Why it exists

- PostgreSQL is the durable source of truth and transaction boundary.
- Kafka is the durable asynchronous event stream and consumer-group delivery.
- Redis is a disposable cache that reduces read pressure and can be rebuilt.

## Core flow

```text
HTTP.CreateJob -> Postgres.Transaction(jobs + outbox)
Postgres.Outbox -> Kafka.Producer -> Kafka.jobs.events.v1
Kafka.Consumer -> Postgres.ProcessedEvents + Postgres.JobState
HTTP.GetJob -> Redis.CacheAside -> Postgres.Fallback
Next.JobDashboard -> HTTP.CreateJob -> HTTP.GetJob
```

## Interview answers

**30 seconds:** PostgreSQL commits the job and outbox together, Kafka decouples
processing, and Redis accelerates reads. The consumer records an event receipt
in the same transaction as its state update, so duplicate delivery is safe.

**2 minutes:** The publisher marks an outbox row only after Kafka acknowledges
it. A crash can duplicate but cannot lose the event. Kafka offsets are committed
after the database transaction. Redis has a TTL and fallback, so it is never
used for progress, idempotency or business truth.

**Deep dive:** Local Kafka is single-node KRaft with separate internal and
external listeners. `depends_on` only orders startup; bounded application
retries and health checks are still needed. Production would add replication,
schema compatibility, retry/DLQ policy and lag observability.

## Important failure cases

- PostgreSQL down: command is rejected; no false `202` response.
- Kafka down: committed outbox remains pending and is retried.
- Worker crash after DB commit: offset is not committed, redelivery is deduped.
- Redis down: query reads PostgreSQL and remains correct.
- Poison event: the MVP reserves `jobs.dlq.v1`; DLQ policy is a later lab.

## Code and graph paths

- [Compose](../compose.yaml)
- [Migration](../platform/migrations/001_init.sql)
- [API command](../platform/internal/app/jobs.go)
- [PostgreSQL transaction](../platform/internal/adapter/postgres/repository.go)
- [Kafka transport](../platform/internal/adapter/kafka/transport.go)
- [Redis cache](../platform/internal/adapter/redis/cache.go)
- [Next.js dashboard](../platform/web/app/page.tsx)
- `HTTP.CreateJobHandler -> Postgres.OutboxEvents`
- `Kafka.JobConsumer -> Postgres.ProcessedEvents`

## Related topics

- [master implementation plan](../docs/11-master-curriculum-implementation-plan.md)
- [local infrastructure plan](../docs/implementation-plans/local-infrastructure.md)
