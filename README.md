# Backend Interview Lab

mini-project สำหรับ Technical Interview วันที่ **22 สิงหาคม 2026** โดยเน้น
Go, Microservices, CQRS, Kafka, PostgreSQL และ Next.js

เป้าหมายไม่ใช่จำทุก API แต่ต้องอธิบายได้ว่าแต่ละแนวคิดแก้ปัญหาอะไร มี trade-off
อะไร และระบบจะทำอย่างไรเมื่อบางส่วนล้มเหลว

## Phase 1 — Foundation

1. [แผนสองวัน](docs/00-crash-plan.md)
2. [Cheat sheet](docs/01-cheat-sheet.md)
3. [Project walkthrough](docs/02-project-walkthrough.md)
4. [System design](docs/03-system-design.md)
5. [Question bank](docs/04-question-bank.md)
6. รัน mini-labs แล้วอธิบายโค้ดออกเสียง

## Phase 2 — Production Deep Dives

Phase นี้เพิ่มจากประสบการณ์ใน resume และใช้สถานการณ์ที่เกิดในงานจริง:

1. [Kafka: import, inventory และ integration](docs/05-kafka-deep-dive.md)
2. [Go production review](docs/06-go-production-review.md)
3. [Next.js สำหรับ backend-focused interview](docs/07-nextjs-deep-dive.md)
4. [PostgreSQL use cases](docs/08-postgresql-use-cases.md)
5. [Resume-based questions และ story preparation](docs/09-resume-based-prep.md)
6. [Coverage matrix ตรวจว่าหัวข้อเดิมอยู่ครบ](docs/10-coverage-matrix.md)
7. [Master curriculum, implementation, Graphify และ GitLab plan](docs/11-master-curriculum-implementation-plan.md)

## Runnable platform vertical slice

The first integrated slice lives under `platform/` and includes a Go HTTP API,
PostgreSQL transactional outbox, Kafka producer/consumer and Redis cache-aside
adapter. See [Infrastructure MyMap](mymap/infrastructure.md) and run:

```powershell
.\scripts\infra-up.ps1 -Profile full -Build
.\scripts\infra-status.ps1
```

Infrastructure definitions are kept explicit in `compose.yaml`, the SQL
migration at `platform/migrations/001_init.sql`, and Kafka topic creation at
`infra/kafka/init-topics.sh`. For incremental labs, start only the foundation
(`PostgreSQL + migration`), eventing (`Kafka + worker`), or cache (`Redis`)
profile; see [the local infrastructure plan](docs/implementation-plans/local-infrastructure.md)
for the exact commands and failure experiments.

ถ้าเวลาน้อยให้อ่าน Kafka → Resume preparation → Next.js → Go production → PostgreSQL
เพราะ Go/PostgreSQL เป็นพื้นฐานจากงานอยู่แล้ว ส่วน Kafka/Next.js เป็นจุดที่ต้องเตรียมภาษา
อธิบายและ follow-up ให้ชัดกว่าเดิม

## Mini-labs

```powershell
go run ./labs/go-context
go run ./labs/go-worker-pool
go run ./labs/cqrs-basic
go test ./...
```

แต่ละ lab ตั้งใจให้สั้นและเน้นแนวคิดเดียว:

- `go-context` — timeout, cancellation และ goroutine leak
- `go-worker-pool` — bounded concurrency, channel และ backpressure
- `cqrs-basic` — แยก command/query และแสดง eventual consistency
- `postgres` — transaction, optimistic locking, idempotency และ outbox ด้วย SQL

Next.js มี code-reading lab เพิ่มที่ `labs/nextjs-import-dashboard` ครอบคลุม Server
Component, Client upload form, Route Handler/BFF และ bounded status polling

## Runnable platform vertical slice

โค้ดใน `platform/` รวม flow ที่เอาไปอธิบายใน interview ได้จริง:

```text
POST /jobs
  → PostgreSQL transaction (jobs + outbox_events)
  → Kafka jobs.events.v1
  → idempotent worker receipt
  → GET /jobs/{id} (Redis cache-aside + PostgreSQL fallback)
```

```powershell
go test ./...
docker compose up -d postgres migrate
docker compose --profile full up --build -d
Invoke-RestMethod http://localhost:8080/healthz
```

`Idempotency-Key` เป็น header ที่จำเป็นสำหรับ `POST /jobs` เพื่อให้ command
retry แล้วไม่สร้าง job ซ้ำ ส่วน Redis เป็น cache เท่านั้น ไม่ใช่ source of truth

## โปรเจกต์จริงที่ใช้ประกอบคำตอบ

- `../goflow-cqrs` — CQRS, transactional outbox, optimistic concurrency,
  asynchronous projection และ worker pool
- `../go-kafka-shipment-tracker` — Kafka partition key, consumer group,
  manual offset commit, idempotent consumer, retry และ DLQ

ทั้งสองโปรเจกต์ผ่าน `go test ./...` บนเครื่องนี้เมื่อวันที่ 20 สิงหาคม 2026

## Stop condition ก่อนสัมภาษณ์

ถือว่าพร้อมเมื่อสามารถอธิบายโดยไม่เปิดโน้ตได้ 4 เรื่อง:

1. Database update กับ Kafka publish ทำอย่างไรไม่ให้ event หาย
2. Kafka ส่ง event ซ้ำแล้ว consumer ป้องกันข้อมูลซ้ำอย่างไร
3. CQRS ทำไมเกิด eventual consistency และรับมืออย่างไร
4. Order, Payment และ Inventory ล้มเหลวบางส่วนแล้ว recover อย่างไร
