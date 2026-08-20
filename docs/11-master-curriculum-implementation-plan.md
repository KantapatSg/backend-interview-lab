# Master Curriculum and Implementation Plan

เอกสารนี้เป็นแผนหลักเพียงชุดเดียวสำหรับเนื้อหา, labs, mini-project, Graphify และ MyMap
แผนเดิมที่แบ่ง Phase เพื่อลดเนื้อหาให้ยกเลิก ให้ใช้ความหมายของ Phase ตามไฟล์นี้เท่านั้น

## สิ่งที่ต้องได้ตอนจบ

### 1. Mini-project สำหรับ GitLab

ชื่อที่เสนอ: `backend-interview-lab`

หนึ่ง repository มีหลาย labs ที่ใช้ domain เดียวกัน แล้วค่อยประกอบเป็นระบบเต็ม:

```text
Go fundamentals
  → Go microservice
  → PostgreSQL job/staging design
  → CQRS + transactional outbox
  → Kafka producer/consumer
  → Next.js dashboard
  → integrated data workflow platform
```

### 2. MyMap

MyMap มีสองรูปแบบที่ใช้ร่วมกัน:

- `mymap/` — สรุปแนวคิดและคำตอบ interview ที่คนอ่านได้
- `graphify-out/graph.html` — interactive node graph สำหรับตาม function/data/event flow

MyMap ต้องตอบได้ทั้งแบบ topic และแบบเส้นทาง เช่น:

```text
HTTP Handler → Command Service → PostgreSQL Transaction → Outbox → Kafka
Kafka Consumer → Idempotency Receipt → Projection → Query API → Next.js
Import File → Staging Rows → Worker Chunks → Business Tables → Job Status
```

## หลักการจัดเนื้อหา

เราเก็บหัวข้อฉบับเต็มตามรายการแรก ไม่ตัดทิ้งเพราะเวลาน้อย แต่แบ่งลำดับ:

- Phase 1 — เนื้อหาหลักที่ควรรู้และ implement ก่อน
- Phase 2 — เนื้อหาเสริม/advanced ที่ทำให้ครบและใช้ตอบ follow-up ลึกขึ้น

ทุกหัวข้อใช้ template เดียวกัน:

1. Concept — คืออะไรและแก้ปัญหาอะไร
2. When/Why — ใช้เมื่อไรและไม่ควรใช้เมื่อไร
3. Implementation Plan — จะเพิ่มเข้า domain หลักตรงไหน
4. Commented Code — comment อธิบายเหตุผล ไม่ใช่อธิบาย syntax ทุกบรรทัด
5. Tests/Experiments — พิสูจน์ happy path และ failure case
6. Interview Q&A — คำถามสั้น, follow-up และ trade-off
7. Graphify — nodes/edges และ path ที่ต้องตรวจ
8. MyMap — สรุปหนึ่งหน้าพร้อม links ไป code/docs

## Domain กลางของทุก Lab

ใช้ `Data Workflow Platform` ที่เชื่อมกับประสบการณ์ทำงาน:

```text
User/API
  → Create Job (ตัวอย่าง IJ หรือ SO)
  → Persist Job + Staging Data
  → Workers claim chunks
  → Validate / transform / process
  → Update progress
  → Publish events
  → Build read model
  → Dashboard แสดง status/result
```

ชื่อ `IJ` และ `SO` ให้เก็บเป็นตัวอย่าง `job_type`; ก่อนทำ business rules จริงต้องยืนยันว่า
คำย่อใน domain หมายถึงอะไร ส่วน lab ใช้ชื่อกลาง `IMPORT_JOB` และ `SALES_ORDER`
เพื่อไม่ผูกกับข้อมูล confidential

## Repository Structure เป้าหมาย

```text
backend-interview-lab/
├── README.md
├── go.mod
├── compose.yaml
├── compose.test.yaml
├── .env.example
├── .gitlab-ci.yml
├── infra/
│   ├── kafka/init-topics.sh
│   └── redis/redis.conf
├── scripts/
│   ├── infra-status.ps1
│   └── reset-local-data.ps1
├── docs/
│   ├── concepts/
│   ├── implementation-plans/
│   ├── interview-questions/
│   └── system-design/
├── mymap/
│   ├── index.md
│   ├── go.md
│   ├── microservices.md
│   ├── cqrs.md
│   ├── kafka.md
│   ├── postgresql.md
│   ├── infrastructure.md
│   ├── redis.md
│   ├── nextjs.md
│   └── system-design.md
├── labs/
│   ├── 01-go-language/
│   ├── 02-go-concurrency/
│   ├── 03-go-http-service/
│   ├── 04-postgres-design/
│   ├── 05-cqrs-outbox/
│   ├── 06-kafka-processing/
│   ├── 07-nextjs-dashboard/
│   └── 08-system-failures/
├── platform/
│   ├── cmd/api/
│   ├── cmd/worker/
│   ├── internal/domain/
│   ├── internal/app/
│   ├── internal/adapter/
│   ├── migrations/
│   └── web/
└── graphify-out/
```

Labs เป็นตัวอย่างเล็ก ส่วน `platform/` คือการนำ concept เดียวกันมารวมเป็น mini-project
จึงไม่ copy ทุก lab เข้า production code แบบตรง ๆ

# Phase 0 — Foundation and Mapping

Phase 0 ทำครั้งเดียวก่อนเพิ่ม labs ใหม่

## Implement

- [ ] เปลี่ยนชื่อ/คำอธิบาย repo จาก crash kit เป็น backend interview lab
- [ ] สร้าง directory structure เป้าหมายโดยรักษาไฟล์เดิม
- [ ] ย้ายเอกสารเดิมเข้า topic ที่เหมาะสมโดยไม่ทำ links เสีย
- [ ] สร้าง `mymap/index.md` และ template ต่อ topic
- [ ] สร้าง lab template: README, plan, code, test, questions
- [ ] กำหนด naming สำหรับ nodes และ domain events
- [ ] สร้าง `.env.example`, Compose project naming และกติกาจัดการ secrets
- [ ] สร้าง infrastructure contract กำหนด service/profile/health/readiness/reset
- [ ] สร้าง Graphify baseline แบบ directed

## Graphify Gate

ใช้:

```text
graphify <repo> --directed --wiki
```

ผลลัพธ์ขั้นต่ำ:

- `graphify-out/graph.html`
- `graphify-out/graph.json`
- `graphify-out/GRAPH_REPORT.md`
- graph health ไม่มี missing/dangling endpoints ที่เกิดจาก extraction ผิด

หลังแต่ละ lab ใช้ incremental update:

```text
graphify <repo> --update --directed
```

# Infrastructure Implementation and Verification Plan

ส่วนนี้เป็น implementation contract สำหรับ Docker/PostgreSQL/Kafka/Redis
เพื่อให้ผู้ implement ไม่ต้องเดา architecture เอง

## ลำดับการเพิ่ม Infrastructure

1. Phase 0 — สร้าง Compose skeleton, `.env.example`, scripts และกติกา reset
2. ก่อน Lab 4 — เพิ่ม PostgreSQL + one-shot migration runner
3. ก่อน Lab 6 — เพิ่ม Kafka KRaft + idempotent topic initializer
4. หลังมี query/polling use case — เพิ่ม Redis cache lab ใน Phase 1B
5. ตอนรวม `platform/` — เพิ่ม API, worker, outbox publisher, web และ full E2E

PostgreSQL และ Kafka เป็น Phase 1 core; Redis เป็น optional derived cache
และจะเป็น required เฉพาะใน cache/resilience lab ที่มี use case ชัดเจน

## Compose Services and Profiles

| Service | Profile | หน้าที่ |
|---|---|---|
| `postgres` | default | durable source of truth |
| `migrate` | default | รัน versioned schema migrations แล้ว exit |
| `kafka` | `eventing`, `full` | local single-node KRaft broker |
| `kafka-init` | `eventing`, `full` | สร้าง/ตรวจ topics แบบ idempotent |
| `redis` | `cache`, `full` | optional ephemeral/derived cache |
| `api` | `app`, `full` | command/query HTTP API |
| `worker` | `eventing`, `full` | Kafka consumer และ job processing |
| `outbox-publisher` | `eventing`, `full` | ส่ง pending outbox events |
| `web` | `app`, `full` | Next.js dashboard |

ไม่กำหนด `container_name` เพื่อให้ local/CI/test หลายชุดแยกกันด้วย
Compose project name ได้ และ pin image versions เพื่อลด behavior drift

คำสั่งเป้าหมาย:

```text
docker compose up -d postgres migrate
docker compose --profile eventing up -d
docker compose --profile cache up -d redis
docker compose --profile full up --build -d
docker compose ps
```

## PostgreSQL and Migrations

- ใช้ named volume และ `pg_isready` health check
- ใช้ one-shot `migrate` service กับ versioned up/down SQL
- ไม่ใช้ `/docker-entrypoint-initdb.d` เป็น migration mechanism หลัก
  เพราะจะรันเฉพาะตอนสร้าง volume ครั้งแรก
- migration failure ต้องทำให้ API/worker ไม่ ready
- migration ต้องรันจาก empty database ได้และรันซ้ำได้โดยไม่เปลี่ยน state
- seed/test fixtures แยกจาก production migrations
- Phase 2 เพิ่ม expand/contract และ migration-from-previous-version test

## Kafka and Topic Initialization

- ใช้ single-node KRaft สำหรับ local/test; ไม่ได้แทน production HA topology
- ปิด automatic topic creation เพื่อให้ topology ตรวจสอบได้
- `kafka-init` สร้างและตรวจ `jobs.events.v1`, `jobs.retry.v1`, `jobs.dlq.v1`
- กำหนด partition count, local replication factor, retention และ config เป็น contract
- ใช้ `job_id` เป็น message key และใช้ versioned event envelope
- topic initializer รันซ้ำได้และต้องตรวจ configuration ไม่ใช่เพียงชื่อ topic
- แยก internal/external listeners เช่น `kafka:9092` และ `localhost:29092`

## Health, Readiness and Runtime Failure

- `/healthz` ตรวจว่า process ยังทำงาน โดยไม่ probe downstream ทุกตัว
- `/readyz` ตรวจเฉพาะ dependency ที่ feature/service นั้นจำเป็นต้องใช้
- PostgreSQL ตรวจ `pg_isready` และ application ping
- Kafka ตรวจ broker metadata/topic operation; Redis ตรวจ `PING`
- worker ready เมื่อ migrations, PostgreSQL และ Kafka พร้อม
- Redis down ไม่ทำให้ query API down ถ้า fallback ไป PostgreSQL ได้
- `depends_on` ช่วย startup order แต่ application ยังต้องมี bounded retry/backoff

## Redis Decision and Lab

ใช้ Redis เมื่อมีปัญหาที่มันแก้ได้จริง:

- cache-aside สำหรับ `job_views` ที่ stale ชั่วคราวได้
- ลด PostgreSQL load จาก status polling
- rate limiting หรือ short-lived coordination เมื่อมี requirement

กำหนด key namespace/version/TTL/invalidation เช่น:

```text
job-view:v1:{tenant_id}:{job_id}
```

Redis ห้ามเป็น source of truth สำหรับ job progress, idempotency,
processed-event receipts, Kafka offsets, outbox หรือ business state ที่กู้คืนไม่ได้
เมื่อ Redis ถูกล้างหรือ unavailable ระบบต้องอ่าน PostgreSQL และ rebuild cache ได้

Redis lab ต้องพิสูจน์ cache hit/miss, TTL expiry, invalidation, stale data,
stampede protection และ outage fallback

## Automated Test Matrix

| ระดับ | Infrastructure | สิ่งที่ต้องพิสูจน์ |
|---|---|---|
| Unit | ไม่ใช้ Docker | domain rules, handlers, event handlers |
| PostgreSQL integration | PostgreSQL | constraints, transaction, outbox, locking, `SKIP LOCKED`, migrations |
| Kafka integration | PostgreSQL + Kafka | publish/consume, duplicate, offset, restart, retry/DLQ |
| Redis integration | PostgreSQL + Redis | hit/miss, TTL, invalidation, outage fallback |
| Full E2E | `full` profile | create job → staging → Kafka → worker → projection → dashboard |
| Resilience | `full` profile | worker/Kafka/Redis restart, duplicate delivery, eventual recovery |

เงื่อนไขการทดสอบ:

- index tests ตรวจ query plan/buffers ไม่ assert เวลาที่เปราะบาง
- แต่ละ run ใช้ Compose project/database/topic namespace แยกกัน
- migration smoke test เริ่มจาก empty volume
- E2E ใช้ bounded polling/eventual assertion และ timeout ไม่ใช้ fixed sleep
- CI เก็บ `docker compose logs` เมื่อ integration/E2E test fail

## Safe Shutdown and Data Reset

การ shutdown ปกติต้องเก็บ named volumes:

```text
docker compose -p backend-interview-lab down --remove-orphans
```

การล้างข้อมูลเป็น script แยก และต้องยืนยันด้วย `-ConfirmDataLoss`:

```text
docker compose -p backend-interview-lab down --volumes --remove-orphans
```

script ต้องใช้ project name ที่ระบุแน่นอน และห้ามใช้ `docker system prune`
หรือ `docker volume prune` ที่อาจกระทบ project อื่น

## Infrastructure Graphify and MyMap

สร้าง `docs/implementation-plans/local-infrastructure.md`, `mymap/infrastructure.md`
และ `mymap/redis.md` โดยตรวจ paths:

```text
Compose.Postgres → Postgres.Migrations → Postgres.Jobs
Compose.Kafka → Kafka.TopicInit → Kafka.JobsEventsV1
Compose.Redis → Cache.JobView → Postgres.JobView
HTTP.Readiness → Postgres.Health
Worker.Readiness → Kafka.Health
Test.E2E → HTTP.CreateJobHandler → Kafka.JobConsumer → Projection.JobView
```

external services ที่ Graphify extract จาก code ไม่ครบต้องมี source-backed
architecture docs/annotations และทุก node สำคัญต้องชี้กลับไปที่ source location ได้

## Infrastructure Acceptance Gate

- `docker compose config` ผ่าน
- แต่ละ profile start ได้และ services เข้า healthy/ready state
- migration จาก empty database ผ่าน
- topic initialization รันซ้ำได้และ config ตรง contract
- integration tests ของ dependency ที่อยู่ใน scope ผ่าน
- full profile E2E และ restart/failure experiments ผ่าน
- normal shutdown รักษา volume; reset ลบเฉพาะ project volumes
- MyMap และ Graphify ตรงกับ implementation จริง

## Infrastructure Risks ที่ต้องรองรับใน Plan/Test

- Kafka internal/external advertised listeners ตั้งผิดทำให้ host หรือ container connect ไม่ได้
- startup race ยังเกิดได้แม้มี `depends_on`
- migration หลาย instance อาจแย่งกันรัน; ต้องมี single-run/locking policy
- Kafka partition expansion อาจเปลี่ยน key-to-partition mapping และ ordering assumption
- Redis invalidation/stampede หรือการเข้าใจผิดว่า cache เป็น durable store
- port collision, Docker Desktop resource limit และ OneDrive bind-mount/file-locking performance
- CRLF ทำให้ shell scripts ใน Linux containers/CI รันไม่ได้
- CI runner อาจไม่รองรับ Docker/Compose หรือมี resource ไม่พอ
- shared database/topic ทำให้ parallel tests ปนกัน
- dev credentials/secrets หลุดเข้า GitLab
- health probe ที่ผูก dependency ทุกตัวอาจสร้าง restart cascade

# Phase 1 — Important Core Topics First

Phase 1 ไม่ได้แปลว่าเนื้อหาสั้น แต่เป็นหัวข้อที่ต้อง implement ก่อน

## Lab 1 — Go Language for Backend

### Concept

- pointer/value receiver
- interface และ implicit implementation
- error wrapping, `errors.Is`, `errors.As`
- slice backing array และ map concurrency
- dependency injection
- package/internal visibility
- table-driven tests

### Implementation Plan

สร้าง domain `Job` และ interfaces ขนาดเล็ก:

```text
JobService
  → JobRepository
  → Clock
  → EventPublisher
```

โค้ดต้องมีตัวอย่าง:

- custom/domain errors
- interface ประกาศฝั่ง consumer
- fake repository ใน unit tests
- pointer receiver สำหรับ state transition
- comment ว่าทำไม business rule อยู่ domain ไม่อยู่ handler

### Acceptance

- state transition tests ผ่าน
- duplicate/invalid transition คืน typed error
- domain package ไม่ import HTTP/PostgreSQL/Kafka

### Graphify Paths

- `HTTPHandler → JobService → Job.Start`
- `JobService → JobRepository`
- `DomainError → HTTPErrorMapping`

## Lab 2 — Go Concurrency and Worker Lifecycle

### Concept

- goroutine และ Go scheduler
- buffered/unbuffered channel
- worker pool
- mutex เทียบ channel ownership
- context timeout/cancellation
- backpressure
- race, deadlock และ goroutine leak
- graceful shutdown

### Implementation Plan

สร้าง worker pool สำหรับ claim `job_chunks`:

```text
Job Poller
  → bounded jobs channel
  → N workers
  → validate/transform chunk
  → progress result channel
  → coordinator
```

โค้ด comment ต้องอธิบาย:

- ใครเป็นเจ้าของการปิด channel
- ทำไม workers มีจำนวนจำกัด
- ทำไมทุก blocking send/receive ตรวจ `ctx.Done()`
- ทำไม persistent job table เป็น source of truth ไม่ใช่ channel
- shutdown แล้ว chunk ถูก retry/resume อย่างไร

### Tests/Experiments

- concurrency ไม่เกิน configured workers
- cancel แล้ว goroutines จบ
- consumer ช้าทำให้ producer ถูก backpressure
- `go test -race` เมื่อ environment รองรับ CGO
- worker fail แล้ว job ไม่หาย

### Graphify Paths

- `JobPoller → jobs channel → ChunkWorker`
- `RootContext → WorkerPool → Repository`
- `ShutdownSignal → HTTPServer → WorkerPool`

## Lab 3 — Go HTTP Microservice

### Concept

- handler/service/domain/repository boundaries
- request validation และ error mapping
- timeout และ idempotency key
- health/readiness
- structured logging และ correlation ID
- sync HTTP เทียบ async job submission
- modular monolith เทียบ microservices

### Implementation Plan

API ขั้นต่ำ:

```text
POST /jobs
GET  /jobs/{id}
GET  /jobs?status=&cursor=
POST /jobs/{id}/cancel
GET  /healthz
GET  /readyz
```

Flow:

```text
HTTP Handler
  → validate request
  → Application Service
  → Domain invariant
  → Repository transaction
  → response/error mapping
```

Comment ต้องอธิบาย transaction boundary, context propagation และเหตุผลที่ handler
ไม่เรียก SQL/Kafka ตรง

### Tests

- handler validation/status mapping
- idempotency key ซ้ำได้ job เดิม
- timeout/cancel ถูกส่งถึง repository
- readiness fail เมื่อ dependency หลักไม่พร้อม

### Graphify Paths

- `POST /jobs → Handler → CommandService → Repository`
- `DomainError → HTTPStatus`
- `SignalContext → GracefulShutdown`

## Lab 4 — PostgreSQL Essential Design

### Concept

- relational model, PK/FK/unique/check constraints
- ACID, MVCC และ isolation levels
- transaction boundary
- optimistic/pessimistic locking
- indexes และ composite index order
- `EXPLAIN (ANALYZE, BUFFERS)`
- offset เทียบ keyset pagination
- JSONB เทียบ typed columns
- connection pooling

### Core Tables

```text
jobs
job_chunks
staging_rows
outbox_events
processed_events
job_views
```

### Staging/Job Table Design

`staging_rows` เป็น persistent table สำหรับพักข้อมูลที่ต้อง process ไม่ใช่ PostgreSQL
temporary table เพราะต้องอยู่ข้าม transaction/process restart และให้หลาย goroutines/
workers claim งานได้

ตัวอย่าง fields:

```text
job_id
chunk_id
row_number
job_type (IMPORT_JOB / SALES_ORDER)
raw_payload JSONB
normalized_payload JSONB
validation_errors JSONB
status
attempt_count
claim_until
created_at / processed_at
```

Worker claim ด้วย transaction และ `FOR UPDATE SKIP LOCKED` หรือ lease pattern

### Index Lab

สร้าง dataset และวัดก่อน/หลัง:

```sql
SELECT id, status, processed_rows, created_at
FROM jobs
WHERE tenant_id = $1 AND status = $2
ORDER BY created_at DESC, id DESC
LIMIT 50;
```

ทดลอง index:

```sql
(tenant_id, status, created_at DESC, id DESC)
```

เอกสารต้องอธิบาย query plan, rows, buffers, execution time และ write cost ไม่สรุปเพียง
ว่า “มี index แล้วเร็วขึ้น”

### Tests/Experiments

- unique idempotency constraint
- optimistic version conflict
- concurrent chunk claim ไม่ซ้ำ
- index plan before/after
- keyset pagination ไม่ซ้ำ/ข้ามเมื่อมี insert

### Graphify Paths

- `JobRepository → jobs`
- `ChunkRepository.Claim → staging_rows index`
- `QueryRepository.ListJobs → composite index`

## Lab 5 — CQRS and Transactional Outbox

### Concept

- Command เทียบ Query
- write model เทียบ read model
- eventual consistency
- projection
- CQRS เทียบ Event Sourcing
- transactional outbox
- idempotent projector
- aggregate version/order

### Implementation Plan

```text
CreateJobCommand
  → Write Model + Outbox ใน transaction เดียว
  → Outbox Publisher
  → JobCreated event
  → Projector
  → Read Model
  → GetJobQuery
```

Comment ต้องอธิบาย:

- ทำไม DB update + publish ตรง ๆ เป็น dual-write problem
- ทำไม outbox ยังสร้าง duplicate ได้
- ทำไม projector ต้องมี `processed_events`/receipt
- ทำไม GET หลัง POST อาจยังไม่เห็นข้อมูล
- ทำไม CQRS ไม่จำเป็นต้องเป็น microservices/Event Sourcing

### Tests

- write + outbox commit/rollback พร้อมกัน
- duplicate event ไม่ update projection ซ้ำ
- event version ก่อนหน้าค้างแล้ว event หลังไม่ถูก apply
- rebuild read model ได้

### Graphify Paths

- `CreateJobCommand → Transaction → OutboxEvent`
- `OutboxPublisher → DomainEvent → Projector`
- `Projector → ProjectionReceipt → JobView`

## Lab 6 — Go + Kafka

### Concept

- broker/topic/partition
- message key และ ordering boundary
- producer acknowledgement/batching
- consumer group และ rebalance
- offset commit
- at-most-once/at-least-once/exactly-once scope
- idempotent consumer
- retry topic, backoff และ DLQ
- consumer lag และ backpressure
- schema evolution/replay

### Implementation Plan

```text
Outbox Publisher
  → Kafka topic jobs.events.v1 [key = job_id]
  → Job Worker consumer group
  → PostgreSQL transaction
       ├── processed_events receipt
       ├── update job/chunk
       └── next outbox event
  → commit Kafka offset
```

### Go + Kafka Advantages

- goroutines ช่วยรัน consumers/workers แบบ concurrent โดยมี worker limit
- context ใช้ควบคุม poll/produce/shutdown lifecycle
- interfaces แยก Kafka transport จาก business handler
- compiled binary/deployment footprint เหมาะกับ stateless workers

ข้อดีเหล่านี้ไม่แทน Kafka guarantees; correctness ยังมาจาก key, offset, transaction และ
idempotency design

### Important Problems to Improve

- process สำเร็จแต่ offset commit fail → duplicate-safe transaction
- Kafka down หลัง DB commit → transactional outbox
- poison event → validation + bounded retry + DLQ
- retry ทำลาย ordering → เลือก partition pause/retry topic/per-key policy
- consumer lag → measure stage latency, partition skew และ DB bottleneck
- rebalance ระหว่างงาน → cancel/revoke handling และ safe checkpoint
- event schema เปลี่ยน → versioned envelope และ compatible rollout

### Tests

- same key อยู่ partition/order เดียวกันตาม assumption ที่ test ได้
- duplicate record ไม่สร้าง side effect ซ้ำ
- consumer restart resume จาก committed offset
- invalid event ไป DLQ
- retry exhausted ไป DLQ
- two consumers scale ภายใต้ consumer group

### Graphify Paths

- `OutboxPublisher → KafkaProducer → jobs.events.v1`
- `KafkaConsumer → EventHandler → processed_events`
- `EventHandler → JobRepository → NextOutboxEvent → CommitOffset`

## Lab 7 — Next.js Dashboard

### Concept

- Server/Client Components
- SSR/CSR/SSG/ISR
- Route Handler/BFF และ Server Actions
- cache/revalidation
- authentication/authorization
- loading/error/not-found states
- polling/SSE/WebSocket
- eventual-consistency UX

### Implementation Plan

```text
Server Component job list/detail
  → Route Handler/BFF
  → Go Query API

Client Component create/upload
  → stable Idempotency-Key
  → Go Command API
  → redirect status page
  → bounded polling until terminal state
```

Comment ต้องบอก server/client boundary, secret location, cache layer และเหตุผลที่ UI
disable button ไม่แทน backend idempotency

### Tests

- component states: idle/submitting/processing/completed/failed
- BFF auth/error mapping
- polling cleanup/AbortController
- frontend build/lint

### Graphify Paths

- `ImportForm → RouteHandler → GoCommandAPI`
- `JobPage → GoQueryAPI → JobView`
- `ImportProgress → PollingRoute → ReadModel`

## Lab 8 — System Design and Failure Walkthrough

### Concept

- service boundary และ database ownership
- sync/async communication
- Saga/compensation
- timeout/retry/circuit breaker
- API gateway/service discovery/load balancing
- observability: logs/metrics/traces
- security และ service-to-service auth
- scaling, hot keys และ connection budgets

### Implementation/Documentation Plan

หัวข้อนี้เน้นเอกสาร, diagrams และ failure experiments ไม่สร้าง infrastructure เกินจำเป็น

ต้องมี diagrams:

- happy path
- database/Kafka dual-write failure
- duplicate delivery
- worker crash/restart
- Saga compensation
- deployment/shutdown sequence
- observability signals

ทุก design decision ใช้รูปแบบ:

```text
Problem → Decision → Why → Trade-off → Failure → Verification
```

### Graphify Paths

- `API → Command Service → Database Ownership`
- `Order/Job Saga → Compensation`
- `Failure Signal → Metric/Log/Trace → Recovery Action`

# Phase 2 — Complete and Advanced Topics

Phase 2 เติมหัวข้อจากรายการแรกให้ครบและใช้ตอบ follow-up ระดับลึก

## Advanced Go

- scheduler details และ `GOMAXPROCS`
- atomics, `sync.Map`, semaphore และ `errgroup`
- memory allocation/escape analysis
- pprof, benchmarks และ load tests
- race detector และ goroutine leak tests
- generics เฉพาะ use case ที่ช่วย type safety

แต่ละหัวข้อมี experiment และ benchmark ไม่ใช่ theory อย่างเดียว

## Advanced Microservices

- REST เทียบ gRPC
- API gateway
- service discovery/load balancing
- circuit breaker state machine
- bulkhead และ rate limiting
- config/secrets management
- backward-compatible API/event deployment
- distributed tracing
- reconciliation jobs

## Advanced Kafka

- rebalance callbacks และ cooperative rebalancing
- producer idempotence/transactions และขอบเขต exactly-once
- partition expansion/hot partition
- delayed retry strategies
- schema registry/compatibility
- replay/rebuild และ disabling external side effects
- consumer-lag operational runbook

## Advanced PostgreSQL and Data Design

### CTE

CTE (`WITH`) เป็น named query result ภายใน statement ไม่ใช่ persistent temp table

ใช้สำหรับ:

- ทำ query หลายขั้นให้อ่านง่าย
- atomic claim + update ด้วย `WITH ... UPDATE ... RETURNING`
- recursive hierarchy เมื่อจำเป็น

Lab ต้องเปรียบเทียบ execution plan และ materialization behavior ตาม PostgreSQL version

### Temporary Table

PostgreSQL temporary table อยู่เฉพาะ session/transaction ตามการตั้งค่า เหมาะกับ
intermediate dataset ภายใน connection เดียว แต่ไม่เหมาะเป็น durable worker queue

### Persistent Staging Table

เหมาะกับ data import/job chunks เพราะ:

- survive process restart
- audit/retry/reconcile ได้
- หลาย workers claim งานได้
- index/constraints/retention policy ได้

นี่คือ pattern หลักสำหรับตัวอย่าง IJ/SO

### Stored Procedure / Function

ใช้เมื่อ operation ต้องอยู่ใกล้ data, ลด round trips หรือบังคับ atomic set-based operation
แต่ต้องอธิบาย versioning/testing/ownership และไม่ซ่อน business workflow ทั้งหมดไว้ใน DB

### Trigger

ใช้แบบจำกัด เช่น audit timestamps/history หรือ invariant ที่ database ต้องบังคับ
หลีกเลี่ยง event/workflow side effects ที่มองไม่เห็น เพราะ debug/order/retry ยาก

### Additional Labs

- partial/covering/expression/GIN indexes
- CTE เทียบ subquery/temp table
- stored procedure batch processing
- audit trigger พร้อม test
- zero-downtime expand/contract migration
- deadlock reproduction และ lock ordering fix
- vacuum/statistics/query-plan lab

## Advanced Next.js

- streaming/Suspense boundaries
- Server Action trade-offs
- OIDC/Keycloak flow
- CSRF/XSS/session security
- presigned object upload
- SSE progress updates
- caching tags/invalidation
- frontend observability/performance

## Advanced System Design

- multi-tenant isolation
- disaster recovery
- data retention/privacy
- capacity estimation
- partitioning/sharding
- read replicas/cache consistency
- CI/CD, rolling deployment และ rollback
- SLO/SLI/error budgets

# MyMap Specification

## Human-readable Topic Page

ทุกไฟล์ `mymap/<topic>.md` ต้องมี:

```text
1. One-sentence definition
2. Why it exists
3. Core flow
4. Important failure cases
5. Trade-offs
6. Interview questions: 30 sec / 2 min / deep dive
7. Code links
8. Graphify node/path links
9. Related topics
```

## Graphify Node Naming

ใช้ชื่อสม่ำเสมอเพื่อให้ค้นหาได้:

```text
HTTP.CreateJobHandler
App.CreateJobCommand
Domain.Job
Postgres.JobRepository
Postgres.StagingRows
Outbox.Publisher
Kafka.JobProducer
Kafka.JobConsumer
Projection.JobProjector
Next.JobPage
```

## Graph Queries ที่ต้องตอบได้

```text
graphify path "HTTP.CreateJobHandler" "Outbox.Publisher"
graphify path "Kafka.JobConsumer" "Postgres.ProcessedEvents"
graphify path "Postgres.StagingRows" "Domain.Job"
graphify path "Next.JobPage" "Projection.JobView"
graphify explain "App.CreateJobCommand"
```

## Graphify Delivery Gate ต่อ Lab

Lab ยังไม่เสร็จจนกว่า:

- code/tests/docs ผ่าน
- Graphify update สำเร็จ
- graph health ถูกตรวจ
- nodes สำคัญมี source location
- MyMap page link ไป code และ graph query ที่เกี่ยวข้อง

# GitLab Delivery Plan

เป้าหมายคือ GitLab ตามคำขอล่าสุด ไม่ใช่ GitHub

## Local First

1. ทำงานใน local repo
2. branch ตาม lab เช่น `lab/go-concurrency`
3. tests และ Graphify gate ผ่าน
4. ตรวจ secret/confidential data
5. commit เฉพาะ scope ของ lab

## GitLab Repository

ชื่อเสนอ: `backend-interview-lab`

ไฟล์สำคัญ:

- `.gitlab-ci.yml`
- README พร้อม learning path
- Docker Compose
- test reports
- `graphify-out/graph.html` เป็น CI artifact หรือ GitLab Pages

Pipeline stages:

```text
lint → unit-test → integration-test → frontend-build → graph-check → pages/artifacts
```

## Merge Request ต่อ Lab

แต่ละ lab ใช้ Merge Request แยก:

```text
lab/go-concurrency
lab/postgres-design
lab/cqrs-outbox
lab/kafka-processing
lab/nextjs-dashboard
```

MR description ต้องมี:

- concept ที่พิสูจน์
- implementation flow
- failure case tests
- screenshots/query plan เมื่อเกี่ยวข้อง
- Graphify output/path
- interview questions ที่เพิ่มใน MyMap

# Model Workflow ต่อ Lab/Merge Request

การระบุ workflow ในเอกสารไม่ได้เปลี่ยน model อัตโนมัติ
ตอนเริ่มแต่ละ task ต้องเลือก model/effort ตามนี้โดยตรง

## 1. Sol high — Plan and Content Contract

- concept, Why/When และ trade-offs
- architecture/implementation plan พร้อม exact scope/files
- acceptance criteria, failure matrix และคำสั่งทดสอบ
- comment intent ว่าจุดใดต้องอธิบาย invariant/เหตุผล/trade-off
- Interview Q&A และ MyMap/Graphify nodes/paths
- non-goals เพื่อป้องกัน overbuild

## 2. Luna high — Implementation and Tests

- implement ตาม contract ทีละ lab/MR
- เพิ่ม Compose, migrations, code และ automated tests ที่อยู่ใน scope
- ใส่ comments เฉพาะจุดตาม comment intent; ไม่ comment syntax ทุกบรรทัด
- รัน acceptance commands และรายงานผล/ข้อจำกัด
- ไม่เปลี่ยน architecture เองเมื่อ contract ไม่ชัด

## 3. Sol high — Review and Finalize

- ตรวจ correctness, concurrency/failure handling และการ overbuild
- ตรวจ comments ให้อธิบายเหตุผลและ invariant ที่ตรงกับ code จริง
- finalize Interview Q&A, MyMap และ Graphify ตาม implementation จริง
- อนุมัติ acceptance gate ก่อน commit/MR

เพื่อลด conflict ไม่ให้ Sol และ Luna แก้ไข files ชุดเดียวกันพร้อมกัน
ให้ทำเป็น handoff ตามลำดับข้างต้น

# ลำดับทำจริง

## Phase 1A — สำคัญที่สุดก่อน

1. Phase 0 repo/MyMap/Graphify foundation
2. Lab 1 Go backend language
3. Lab 2 concurrency/worker lifecycle
4. Lab 3 HTTP microservice
5. Infrastructure A: PostgreSQL + migration runner + integration-test harness
6. Lab 4 PostgreSQL transactions/index/staging
7. Lab 5 CQRS/outbox
8. Infrastructure B: Kafka KRaft + topic initializer + integration-test harness
9. Lab 6 Kafka
10. สรุป MyMap สำหรับ Go + Microservices + PostgreSQL + CQRS + Kafka

## Phase 1B — สำคัญรองลงมาแต่ยังเป็น core

11. Lab 7 Next.js dashboard และ query/polling use case
12. Infrastructure C: optional Redis cache/resilience lab
13. Lab 8 system-design/failure walkthrough
14. รวม labs เข้า `platform/` mini-project
15. Infrastructure D: full Compose profile + E2E/resilience tests
16. CI/Graphify final graph และ MyMap final review
17. เตรียม GitLab README และ Merge Requests

## Phase 2 — ทำให้ครบตามรายการแรก

18. Advanced Go
19. Advanced Microservices
20. Advanced Kafka
21. Advanced PostgreSQL/CTE/temp/staging/procedure/trigger
22. Advanced Next.js
23. Advanced System Design/operations/security
24. Update MyMap, Graphify wiki และ interview question bank รอบสุดท้าย

# Current Status Mapping

ของเดิมไม่ทิ้ง แต่จัดเข้าระบบใหม่:

- `labs/go-context` → Lab 2
- `labs/go-worker-pool` → Lab 2
- `labs/cqrs-basic` → Lab 5
- `labs/postgres` → Lab 4/5
- `labs/nextjs-order-status` → Lab 7
- `labs/nextjs-import-dashboard` → Lab 7
- `goflow-cqrs` → reference implementation สำหรับ Lab 3/5
- `go-kafka-shipment-tracker` → reference implementation สำหรับ Lab 6
- docs 00–10 → source สำหรับ `docs/` และ `mymap/`

งานถัดไปคือ Phase 0: restructure แบบ compatibility-safe แล้วสร้าง Graphify baseline ก่อน
เพิ่ม implementation ใหม่

# Implemented Vertical Slice (current checkpoint)

ส่วนที่ implement แล้วใน local repo:

- Go domain/application service พร้อม typed errors และ unit tests
- HTTP `POST /jobs`, `GET /jobs/{id}`, `/healthz`, `/readyz`
- PostgreSQL jobs + outbox + processed-event receipt schema
- transactional outbox publisher และ Kafka consumer แบบ manual commit
- Redis cache-aside ที่ fallback กลับ PostgreSQL ได้
- Docker Compose สำหรับ PostgreSQL, migration runner, Kafka KRaft, Kafka topic init,
  Redis, Go API/worker และ Next.js dashboard
- Graphify code graph ที่ `graphify-out/` พร้อม report/HTML

สิ่งที่ยังเป็น Phase ถัดไป ไม่ควรอ้างว่าเสร็จแล้ว: retry/DLQ policy แบบเต็ม,
progress projection/`job_views`, `FOR UPDATE SKIP LOCKED` integration experiments,
full Docker E2E และ advanced Next.js/auth/system-design labs
