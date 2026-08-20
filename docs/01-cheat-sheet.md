# Interview Cheat Sheet

## Go

### Goroutine และ scheduler

Goroutine เป็นงาน concurrent ที่ Go runtime schedule ลงบน OS threads อีกที
จึงเริ่มได้เบากว่า thread แต่ไม่ใช่ของฟรี ทุก goroutine ยังใช้ stack, scheduler time
และ resource ที่มันถืออยู่ หากไม่มีทางออกจะกลายเป็น goroutine leak

หลักตอบเรื่อง concurrency:

- ใช้ goroutine เมื่อมีงานที่ทำ overlap กันได้ ไม่ใช่เพื่อทำทุก function ให้ async
- จำกัด concurrency ด้วย semaphore หรือ worker pool
- ทุก goroutine ระยะยาวควรมี ownership และ shutdown path
- shared mutable state ต้องป้องกันด้วย mutex หรือให้ goroutine เดียวเป็น owner

### Channel

- Unbuffered channel: send และ receive ต้อง rendezvous กัน
- Buffered channel: sender ไปต่อได้จน buffer เต็ม
- Buffer ช่วยรับ burst แต่ไม่ได้แก้ consumer ที่ช้าถาวร
- ผู้ส่งที่เป็นเจ้าของ lifecycle ควรเป็นผู้ปิด channel
- รับจาก channel ที่ปิดแล้วได้ zero value และ `ok == false`
- ส่งไป channel ที่ปิดแล้ว panic
- การปิด channel ซ้ำ panic

`select` ใช้รอหลายเหตุการณ์ เช่น งานใหม่, timeout และ cancellation

```go
select {
case job := <-jobs:
    handle(job)
case <-ctx.Done():
    return ctx.Err()
}
```

### Context

`context.Context` ใช้ส่ง deadline, cancellation และ request-scoped metadata
ผ่าน call chain

- รับ context เป็น parameter แรก
- อย่าเก็บ context ใน struct ระยะยาว
- เรียก `cancel()` เพื่อคืน timer/resource
- อย่าใช้ `context.Background()` กลาง request chain เพราะจะตัด cancellation
- value ใน context เหมาะกับ request ID ไม่เหมาะกับ dependency หรือ business input

### Mutex กับ channel

- Mutex เหมาะกับการป้องกัน state ใน memory ที่สั้นและตรงไปตรงมา
- Channel เหมาะกับการส่ง ownership/work และประสาน lifecycle
- อย่าเลือก channel เพียงเพราะเป็น Go; mutex มักชัดกว่าสำหรับ counter หรือ cache

### Error handling

```go
if err != nil {
    return fmt.Errorf("load order %s: %w", id, err)
}
```

- `%w` รักษา error chain
- `errors.Is` ตรวจ sentinel error ใน chain
- `errors.As` ดึง typed error
- เพิ่ม context ทุก layer แต่ไม่ log error ซ้ำทุก layer
- domain error ควรถูก map เป็น HTTP status ที่ transport layer

### Interface

Go implement interface แบบ implicit ทำให้ consumer สามารถประกาศ interface ขนาดเล็กตามสิ่งที่ต้องใช้

```go
type OrderRepository interface {
    Save(context.Context, Order) error
}
```

อย่าสร้าง interface ใหญ่ล่วงหน้า และปกติควรประกาศ interface ฝั่ง consumer

### Slice และ map

- Slice เป็น descriptor ที่ชี้ไป backing array; copy slice ไม่ได้ copy elements
- `append` อาจ reuse array เดิมหรือ allocate array ใหม่
- Map ไม่ปลอดภัยสำหรับ concurrent read/write
- Iteration order ของ map ไม่ควรนำไปพึ่งพา
- `nil` slice append ได้ และมัก encode JSON ต่างจาก empty slice ตาม encoder/contract

### HTTP service

เส้นทางที่ควรอธิบายได้:

```text
Handler → application service → domain → repository → database
```

- Handler ดูแล HTTP parsing/validation/status
- Application service orchestrate use case และ transaction boundary
- Domain รักษา business invariant
- Repository ซ่อน persistence detail
- Graceful shutdown หยุดรับงานใหม่ รอ in-flight requests แล้วปิด dependency

## Microservices

### Service boundary

แบ่งตาม business capability และ consistency boundary ไม่ใช่แบ่งทุกตารางเป็น service
ข้อมูลที่ต้องรักษา invariant ใน transaction เดียวกันมักควรอยู่ service เดียวกัน

ข้อดี:

- deploy/scale แยก
- ownership ชัด
- fault isolation เมื่อออกแบบดี

ต้นทุน:

- network failure
- eventual consistency
- deployment/observability ซับซ้อน
- contract และ schema evolution
- integration testing ยากขึ้น

ถ้า domain และทีมยังเล็ก modular monolith มักเป็นจุดเริ่มที่ดีกว่า

### Sync กับ async

HTTP/gRPC เหมาะเมื่อ caller ต้องการผลทันที แต่เกิด temporal coupling และ failure chain
Kafka เหมาะกับ event, fan-out, buffering และ replay แต่เพิ่ม eventual consistency

ทุก remote call ควรคิดถึง:

- timeout
- bounded retry + exponential backoff + jitter
- idempotency
- circuit breaker เมื่อ downstream fail ต่อเนื่อง
- correlation ID และ tracing

ห้าม retry non-idempotent operation แบบไม่ออกแบบ idempotency key

### Distributed transaction

หลีกเลี่ยง transaction ที่ครอบหลาย service/database เมื่อทำได้ ใช้:

- local transaction ต่อ service
- event เพื่อเดิน workflow ต่อ
- Saga compensation
- reconciliation job เพื่อซ่อม state ที่ค้าง

Saga ให้ business consistency ไม่ใช่ ACID rollback จริง Compensation อาจล้มเหลวได้
จึงต้อง retry, observe และ reconcile

## CQRS

CQRS คือแยก model/เส้นทาง Command ที่เปลี่ยน state ออกจาก Query ที่อ่านข้อมูล
ไม่จำเป็นต้องมี Kafka, microservices หรือ Event Sourcing

เหมาะเมื่อ:

- write side มี business rules ซับซ้อน
- read shape ต่างจาก normalized write model มาก
- read และ write ต้อง scale ต่างกัน
- ยอมรับ eventual consistency ได้

ไม่เหมาะเมื่อเป็น CRUD เล็ก ๆ เพราะเพิ่ม model, projection และ operational complexity

### Command

- แสดง intent เช่น `ConfirmOrder` ไม่ใช่ `SetStatus`
- validate authorization และ business invariant
- เปลี่ยน state ภายใน transaction
- duplicate command ควรถูกป้องกันด้วย idempotency key/version

### Query

- ไม่เปลี่ยน business state
- optimize model ตาม use case
- denormalize ได้
- cache และ scale แยกได้

### Eventual consistency

หลัง command commit แล้ว projection อาจยังไม่ทัน จึงมีทางเลือก:

- command response คืน ID/status จาก write side
- UI แสดง `PROCESSING`
- polling/SSE/WebSocket รอ read model
- read-your-own-write ชั่วคราวเมื่อ requirement ต้องการ
- วัด projection lag และมี reconciliation

### CQRS กับ Event Sourcing

- CQRS: แยก command/query model
- Event Sourcing: เก็บ event history เป็น source of truth แล้ว derive current state
- ใช้แยกกันหรือร่วมกันได้

Event Sourcing ต้องจัดการ replay, snapshot, event version, temporal queries และ deletion/privacy
จึงไม่ควรเลือกเพียงเพราะใช้ CQRS

## Transactional Outbox

ปัญหา dual write:

```text
UPDATE database สำเร็จ
PUBLISH Kafka ล้มเหลว
```

แก้ด้วย transaction เดียว:

```text
BEGIN
  UPDATE business state
  INSERT outbox event
COMMIT
```

publisher แยกมาอ่าน outbox แล้ว publish ภายหลัง

Outbox ไม่ได้กำจัด duplicate เพราะ publisher อาจ publish สำเร็จแต่ล้มก่อน mark processed
ดังนั้น consumer ยังต้อง idempotent

วิธี claim หลาย publisher:

```sql
SELECT ...
FROM outbox_events
WHERE published_at IS NULL
FOR UPDATE SKIP LOCKED
LIMIT 100;
```

ต้องดูแล backlog, retry, poison event, retention และ monitoring

## Kafka

### Topic, partition และ key

- Topic แบ่งเป็น partitions
- Ordering รับประกันภายใน partition เท่านั้น
- Key เดียวกันปกติถูก route ไป partition เดียวกัน
- ใช้ aggregate ID เช่น `order_id` เป็น key เมื่อ event ของ order เดียวต้องเรียงลำดับ
- เพิ่ม partition อาจเปลี่ยน key-to-partition mapping สำหรับ event ใหม่

### Consumer group

ภายใน group เดียว partition หนึ่งถูก assign ให้ consumer หนึ่งตัว ณ ขณะหนึ่ง

- 10 partitions, 4 consumers: consumer บางตัวรับหลาย partition
- 10 partitions, 20 consumers: 10 consumers ว่าง
- จำนวน partition เป็นเพดาน parallelism ของ consumer group

Rebalance เปลี่ยน assignment จึงต้องหยุด/commit งานอย่างระวัง

### Offset และ delivery semantics

- Commit ก่อน process: มีโอกาสสูญ event → at-most-once
- Process แล้ว commit: มีโอกาสทำซ้ำ → at-least-once
- แนวทางทั่วไปคือ at-least-once + idempotent consumer

กรณีสำคัญ:

```text
database commit สำเร็จ
consumer crash ก่อน offset commit
Kafka ส่ง event เดิมอีกครั้ง
```

ป้องกันด้วย transaction เดียวที่ insert `processed_events(event_id)` และ update projection

### Exactly-once

Kafka transaction ช่วย atomic ระหว่าง Kafka records/offsets ในขอบเขต Kafka
แต่ไม่ได้ทำให้ PostgreSQL หรือ external API กลายเป็น exactly-once อัตโนมัติ
ทุกครั้งที่ใช้คำว่า exactly-once ต้องบอกขอบเขต

### Retry และ DLQ

- transient error เช่น database unavailable: retry
- permanent invalid payload/schema: DLQ
- business rejection: บันทึกเหตุผลและ commit; retry มักไม่ช่วย
- retry ต้องมี max attempts, backoff, jitter และ observability

Retry topic อาจเปลี่ยน ordering ข้าม event ของ aggregate เดียว ต้องยอมรับหรือออกแบบ
partition pause/per-key serialization ตาม requirement

### Schema evolution

- event เป็น contract ระหว่างทีม
- เพิ่ม optional field มักปลอดภัยกว่า rename/remove
- ใส่ `event_type`, `event_id`, `occurred_at`, `schema_version`, `correlation_id`
- consumer ควรรองรับ producer เก่าในช่วง rollout
- หลีกเลี่ยง payload ที่ผูกกับ database row โดยตรง

## PostgreSQL

### ACID และ MVCC

- Atomicity: transaction สำเร็จทั้งหมดหรือไม่เลย
- Consistency: constraints/invariants ยังคงถูกต้อง
- Isolation: concurrent transactions ไม่ควรเห็น intermediate state ตามระดับ isolation
- Durability: committed data รอดหลัง crash ตาม guarantee ของระบบ

MVCC ทำให้ reader และ writer block กันน้อยลงโดยแต่ละ transaction เห็น snapshot
แต่ row locks/deadlocks ยังเกิดได้

### Isolation

- Read Committed: แต่ละ statement เห็นข้อมูล committed ล่าสุด; default ของ PostgreSQL
- Repeatable Read: transaction เห็น snapshot คงที่และ PostgreSQL ป้องกันหลาย anomaly
- Serializable: ผลลัพธ์เสมือนรันทีละ transaction แต่อาจ abort ให้ retry

เลือกสูงขึ้นไม่ได้แปลว่าดีกว่าเสมอ เพราะมี contention/retry cost

### Pessimistic vs optimistic locking

Pessimistic:

```sql
SELECT * FROM inventory WHERE sku = $1 FOR UPDATE;
```

เหมาะเมื่อ conflict สูงและต้อง lock ก่อนตัดสินใจ แต่ lock นานทำให้ contention

Optimistic:

```sql
UPDATE orders
SET status = $1, version = version + 1
WHERE id = $2 AND version = $3;
```

ถ้า affected rows เป็น 0 แปลว่า stale version ให้ reload/retry หรือคืน conflict

### Index

Index เพิ่มความเร็ว read แต่เพิ่ม storage และ write amplification

สำหรับ query:

```sql
WHERE account_id = $1 AND status = $2
ORDER BY created_at DESC
```

candidate index คือ `(account_id, status, created_at DESC)` แต่ต้องดู cardinality,
query distribution และ `EXPLAIN (ANALYZE, BUFFERS)` ไม่เดาจาก syntax อย่างเดียว

PostgreSQL อาจเลือก sequential scan เมื่อ table เล็ก, query คืนข้อมูลสัดส่วนสูง
หรือ statistics ประเมินว่า index แพงกว่า

### Pagination

Offset ใช้ง่ายแต่ช้าเมื่อ offset สูงและผลอาจเลื่อนเมื่อมี insert/delete

Keyset:

```sql
WHERE (created_at, id) < ($cursor_time, $cursor_id)
ORDER BY created_at DESC, id DESC
LIMIT 50;
```

ต้องมี deterministic tie-breaker เช่น `id`

### Deadlock

เกิดเมื่อ transaction รอกันเป็นวง PostgreSQL จะ abort หนึ่ง transaction

ลดความเสี่ยงด้วย:

- lock resources ในลำดับเดียวกัน
- transaction สั้น
- ไม่ call network ขณะถือ transaction
- retry เฉพาะ error ที่ retryable ด้วย bounded backoff

### Connection pool

Connection มีต้นทุนและ database รับได้จำกัด หากทุก service เปิด pool ใหญ่จะรวมกันเกิน limit
ต้องกำหนด max connections, acquisition timeout, idle/lifetime และวัด saturation

## Next.js

### Server และ Client Component

Server Component เป็นค่าเริ่มต้นใน App Router:

- เข้าถึง server resource/secret ได้
- ลด JavaScript ที่ส่ง client
- ใช้ browser state/event handler ไม่ได้

Client Component ต้องมี `"use client"` และใช้เมื่อจำเป็นต้องมี state, effect,
event handler หรือ browser API อย่าทำทั้ง page เป็น client โดยไม่จำเป็น

### Rendering

- CSR: browser fetch/render; interactive ดีแต่ initial load/SEO อาจเสีย
- SSR: render ต่อ request; fresh แต่ server cost สูงกว่า
- SSG: build ล่วงหน้า; เร็วแต่ข้อมูลอาจเก่า
- ISR: static output ที่ revalidate ได้

เลือกตาม freshness, personalization, SEO และ cost ไม่ใช่เลือกแบบเดียวทั้งระบบ

### Cache และ mutation

หลัง mutation ต้องคิดว่า cache ใด stale และ invalidate/revalidate ให้ถูก scope
อาการ “database เปลี่ยนแล้ว UI ยังเก่า” มักเกิดจาก cache layer ไม่ใช่ database

สำหรับ CQRS อย่าสัญญาว่า revalidation ทำให้ projection พร้อมทันที เพราะ source data
เองอาจยัง eventual-consistent UI ควรแสดง pending state

### Authentication

- Authentication: คุณเป็นใคร
- Authorization: คุณทำสิ่งนี้ได้หรือไม่
- Prefer HttpOnly + Secure + SameSite cookie สำหรับ session/token ที่ไม่ต้องให้ JS อ่าน
- localStorage ถูก JavaScript อ่านได้ จึงเสี่ยงเมื่อมี XSS
- Cookie-based mutation ต้องคิดเรื่อง CSRF
- ตรวจ authorization ฝั่ง server ทุกครั้ง; ซ่อนปุ่มใน UI ไม่ใช่ security boundary

### Route Handler และ Server Action

Route Handler เหมาะกับ HTTP endpoint ที่ client/external system เรียก
Server Action เหมาะกับ mutation ที่ผูกกับ React form/UI
ทั้งคู่ยังต้อง validate input, authenticate, authorize และจัดการ error

## Observability

สามเสาหลัก:

- Logs: เหตุการณ์แบบ structured พร้อม service, request/event ID
- Metrics: rate, errors, duration, saturation และ business metrics
- Traces: เส้นทาง request ข้าม service

สำหรับ Kafka เพิ่ม:

- consumer lag
- retry/DLQ rate
- processing latency
- duplicate/rejected events
- outbox backlog/oldest event age

Correlation ID เชื่อม request กับ event ได้ แต่ event ID ต้องแยกเพื่อ idempotency
