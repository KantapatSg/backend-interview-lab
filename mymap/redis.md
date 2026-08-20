# Redis Cache MyMap

## Definition

Redis is a low-latency derived cache in this lab. PostgreSQL remains the only
durable source of truth.

## Key contract

```text
job-view:v1:{tenant_id}:{job_id}
TTL: 30 seconds
```

The versioned prefix makes a schema change an explicit cache migration. A miss,
expired value or Redis outage falls back to PostgreSQL and can rebuild the key.

## Interview answer

Use Redis when repeated status reads are putting avoidable pressure on the
database and stale data is acceptable for a short window. Do not put job
progress, idempotency receipts, offsets or outbox state only in Redis: losing a
cache must never lose business truth.

## Failure experiments

- first GET is a miss, second GET is a hit;
- wait for TTL expiry and confirm PostgreSQL fallback;
- stop Redis and confirm GET still succeeds;
- update the durable row and verify the cache can be invalidated/rebuilt.

## Code path

`JobService.Get -> Redis.Cache.Get -> Postgres.Repository.GetJob`
