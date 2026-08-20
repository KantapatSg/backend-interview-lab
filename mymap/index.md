# Backend Interview Lab MyMap

This map connects the learning notes to one Data Workflow Platform. Read a
topic page for the explanation, then follow its code and Graphify paths.

| Topic | Key question | Implementation |
|---|---|---|
| [Infrastructure](infrastructure.md) | Why PostgreSQL + Kafka + Redis, and what happens when each fails? | `compose.yaml`, `platform/` |
| Go | How are ownership, errors and cancellation expressed? | `labs/go-context`, `labs/go-worker-pool` |
| Microservices | Where do handler, service and adapter boundaries belong? | `platform/internal/httpapi`, `platform/internal/app` |
| PostgreSQL | How do transactions, indexes and staging support jobs? | `labs/postgres`, `platform/migrations` |
| CQRS/Outbox | How do we avoid a database/Kafka dual-write gap? | `labs/cqrs-basic`, `platform/internal/adapter/postgres` |
| Kafka | How are ordering, retries and duplicate delivery handled? | `platform/internal/adapter/kafka` |
| Next.js | How does a UI observe eventual consistency? | `labs/nextjs-import-dashboard` |
| System design | How do failures recover without hiding trade-offs? | `docs/03-system-design.md` |

## Main path

```text
HTTP.CreateJob -> App.JobService -> Postgres.CreateJob (jobs + outbox)
  -> OutboxPublisher -> Kafka.jobs.events.v1 -> Kafka.JobConsumer
  -> ProcessedEvents + JobState -> Redis.JobView (fallback Postgres)
  -> Next.js status page
```

## Graphify queries

```text
graphify path "HTTP.CreateJobHandler" "Postgres.OutboxEvents"
graphify path "Kafka.JobConsumer" "Postgres.ProcessedEvents"
graphify explain "App.JobService"
```

Graphify output is generated from the repository and architecture annotations;
update it after each lab implementation.
