# Kafka Deep Dive: จาก Background Job ไปสู่ Event-Driven System

เอกสารนี้ต่อยอดจากประสบการณ์ asynchronous data import, background processing,
status tracking, warehouse workflow และ external integration ให้เป็นคำตอบ Kafka ที่
อธิบายจากปัญหาจริงได้ โดยไม่ต้อง claim ว่าเคยใช้ Kafka production หากยังไม่ได้ใช้

## Mental model ที่ต้องแม่น

```text
Producer → Topic → Partition → Ordered log
                              ↓
                        Consumer Group
                              ↓
                     Local transaction
                              ↓
                       Commit offset
```

Kafka ไม่ใช่เพียง queue แต่เป็น durable partitioned log ที่หลาย consumer groups
อ่านซ้ำ/replay ตาม offset ของตัวเองได้

## 1. Use case: Data Import Pipeline

จากระบบที่ผู้ใช้อัปโหลด Excel/CSV แล้ว backend validate, transform และติดตามสถานะ
สามารถออกแบบ event flow ได้ดังนี้

```text
Upload API
  → เก็บไฟล์ใน MinIO/S3
  → สร้าง import_job = PENDING
  → outbox: ImportRequested
  → Kafka [key = import_job_id]
  → Validation Worker
  → Transformation Worker
  → Database Writer
  → ImportCompleted / ImportFailed
  → Status Projection
  → Frontend status page
```

### ทำไมไม่ส่งไฟล์ทั้งก้อนผ่าน Kafka

- message ใหญ่เพิ่ม network, broker memory/disk และ retry cost
- ควรเก็บไฟล์ใน object storage แล้ว event ถือ URI/object key, checksum และ metadata
- consumer ต้องตรวจ authorization/ownership ผ่าน trusted metadata ไม่รับ arbitrary URL

ตัวอย่าง event:

```json
{
  "eventId": "uuid",
  "eventType": "ImportRequested",
  "importJobId": "job-123",
  "objectKey": "imports/tenant-9/job-123.csv",
  "checksum": "sha256:...",
  "schemaVersion": 1,
  "occurredAt": "2026-08-20T10:00:00Z",
  "correlationId": "request-456"
}
```

### Partition key เลือกอะไร

ใช้ `import_job_id` เมื่อ events ของ job เดียวต้องเรียงลำดับ แต่ jobs คนละตัวทำพร้อมกันได้

หากแตกหนึ่ง job เป็นหลาย chunks:

- key เป็น `import_job_id:chunk_id` เพิ่ม parallelism
- coordinator ต้องรู้ expected chunks และ deduplicate completion
- finalization ต้อง atomic เช่น compare completed count กับ expected count
- อย่าใช้ row number อย่างเดียว เพราะ retry/re-upload อาจชนกัน

### Idempotency ระดับใดบ้าง

1. API: `(tenant_id, idempotency_key)` ป้องกันสร้าง import job ซ้ำ
2. Consumer: `(consumer_group, event_id)` ป้องกัน event เดิมสร้าง side effect ซ้ำ
3. Chunk: `(import_job_id, chunk_id)` ป้องกันเขียน chunk ซ้ำ
4. Business row: natural/business key หรือ staging-row ID ตาม requirement
5. External call: stable idempotency key เมื่อส่งต่อระบบอื่น

### Validation error กับ infrastructure error

- ไฟล์ format ผิด/ข้อมูล business rule ผิด: บันทึก validation result และจบ job แบบ FAILED
- PostgreSQL/MinIO unavailable: transient retry
- event envelope เสียหรือ schema ไม่รองรับ: DLQ
- row ผิดบางส่วน: requirement ต้องบอกว่า fail ทั้ง job หรือ partial success

Retry validation error ซ้ำโดยไม่เปลี่ยน input ไม่มีประโยชน์

### Backpressure

ถ้า upload เร็วกว่าการ import:

- Kafka เก็บ backlog และ consumer lag เพิ่ม
- จำกัด upload/active jobs ต่อ tenant
- จำกัด worker concurrency ตาม database/object-storage capacity
- batch database writes
- แสดงสถานะ QUEUED ไม่สร้าง goroutine ต่อไฟล์แบบไม่จำกัด
- alert จาก lag และ oldest pending job age

### Replay

Replay เหมาะกับ rebuild projection หรือ reprocess หลังแก้ bug แต่ต้องตอบ:

- replay ไป table เดิมหรือ versioned table ใหม่
- side effect ภายนอกต้องถูกปิด/แยก ไม่ส่ง email/charge ซ้ำ
- consumer group ใหม่มี idempotency namespace ใหม่หรือไม่
- schema เก่าจะถูกอ่านอย่างไร

## 2. Use case: Inventory Transfer

จาก workflow โอนสินค้าระหว่างคลัง:

```text
TransferRequested
  → ReserveSourceStock
  → SourceStockReserved
  → DispatchTransfer
  → GoodsReceived
  → AddDestinationStock
  → TransferCompleted
```

### Ordering key เป็น transfer ID หรือ SKU

- `transfer_id`: events ของ transfer เดียวเรียง แต่หลาย transfers ที่แตะ SKU เดียวอาจชนกัน
- `sku_id`: stock changes ของ SKU เดียวเรียง แต่ transfer หลาย SKU กระจายหลาย partitions
- ไม่มี key เดียวแก้ทุก invariant ต้องให้ database เป็น final concurrency guard

คำตอบที่ดี:

> ผมใช้ transfer ID เพื่อ workflow ordering และให้ Inventory Service ใช้ atomic conditional
> update/row lock ต่อ SKU เพื่อป้องกัน oversell เพราะ Kafka ordering ไม่แทน database
> invariant ข้าม aggregate

### Stock reservation

```sql
UPDATE inventory
SET available = available - $1,
    reserved = reserved + $1
WHERE warehouse_id = $2
  AND sku = $3
  AND available >= $1;
```

ถ้า affected rows เป็นศูนย์ให้ emit `StockReservationRejected` ไม่ retry แบบ infrastructure

### Compensation

หาก destination รับสินค้าไม่ได้หลัง reserve:

- emit `ReleaseSourceReservation`
- command ต้อง idempotent
- compensation failure ต้อง retry/alert/reconcile
- physical inventory อาจต้อง manual resolution จึงไม่ควร claim ว่า rollback ได้สมบูรณ์

## 3. Use case: External API/Webhook

เช่น approval API หรือ third-party integration:

```text
Local transaction
  → outbox IntegrationRequested
  → Kafka
  → Integration Worker
  → Third-party API
  → IntegrationSucceeded / IntegrationFailed
```

จุดสำคัญ:

- timeout ไม่แปลว่า third party ไม่ทำงาน อาจสำเร็จแต่ response หาย
- ใช้ stable idempotency key/provider reference
- rate limit และ retry budget แยกต่อ provider
- circuit breaker ป้องกัน worker ค้างทั้งหมด
- เก็บ sanitized request/response metadata ไม่ log secret/PII
- webhook ขาเข้าต้อง verify signature, timestamp และ deduplicate event ID

## 4. Delivery semantics แบบ step-by-step

### At-most-once

```text
commit offset → process
```

Crash ระหว่างกลางทำให้ event หาย เหมาะเฉพาะข้อมูลที่ยอมสูญได้จริง

### At-least-once

```text
process → local DB commit → commit offset
```

Crash หลัง DB commit ทำให้ event ถูกส่งซ้ำ เป็นแนวทางทั่วไปเมื่อ consumer idempotent

### Idempotent database consumer

```text
BEGIN
  INSERT processed_events(group, event_id) ON CONFLICT DO NOTHING
  ถ้า insert ได้:
      validate state transition
      update projection/domain state
      insert outbox สำหรับ event ถัดไป
COMMIT
commit Kafka offset
```

Receipt ต้องอยู่ transaction เดียวกับ side effect ไม่เช่นนั้นอาจ mark processed แต่ update
ยังไม่เกิด หรือ update แล้วแต่ receipt ยังไม่เกิด

## 5. Ordering, retry และ poison messages

Kafka รับประกัน ordering เฉพาะ record ที่อยู่ partition เดียวกัน

สถานการณ์:

```text
offset 10: ImportStarted      → transient failure
offset 11: ImportCompleted    → พร้อม process
```

ทางเลือก:

- หยุด partition แล้ว retry offset 10: รักษาลำดับแต่ block key อื่นใน partition
- ส่ง offset 10 ไป retry topic แล้วเดินต่อ: throughput ดีแต่ ordering เปลี่ยน
- per-key queue/state machine: ละเอียดแต่ซับซ้อนและต้องดูแล memory/durability

ต้องเลือกตาม invariant ไม่ใช่บอกว่า retry topic ถูกเสมอ

Poison message ต้องไม่ทำให้ partition ค้างตลอดไป:

- validate envelope ก่อน business logic
- จำกัด retry
- DLQ พร้อม original topic/partition/offset/error/schema version
- alert และมี replay tooling

## 6. Consumer group และ scaling

```text
Topic: 12 partitions
Group A: Import Projector, 4 consumers
Group B: Audit Writer, 2 consumers
Group C: Notification, 12 consumers
```

แต่ละ group อ่าน event ชุดเดียวกันได้อิสระ ภายใน group partition ถูกแบ่งกัน

เพิ่ม consumer ไม่ช่วยเมื่อ:

- consumers มากกว่า partitions
- partition skew จาก hot key
- database connection/lock เป็น bottleneck
- external API rate limit ต่ำ
- processing เป็น sequential ต่อ aggregate

## 7. Producer concerns

- ใช้ key ตาม ordering boundary
- รอ acknowledgement ตาม durability requirement
- batching/compression เพิ่ม throughput แต่เพิ่ม latency/CPU
- producer retry อาจสร้าง duplicate หากไม่มี idempotent producer/consumer design
- outbox ใช้เมื่อ business transaction กับ publish ต้องไม่หลุดจากกัน
- อย่า publish ก่อน DB commit เพราะ consumer อาจเห็น entity ที่ยังไม่มีหรือถูก rollback

## 8. Schema evolution

กฎ rollout:

1. เพิ่ม optional field ก่อน
2. deploy consumer ที่อ่านได้ทั้ง schema เก่า/ใหม่
3. deploy producer ใหม่
4. monitor unknown/rejected schema
5. ลบ field เก่าหลัง consumer ทุกตัว migrate แล้ว

Event ควรเป็น domain fact เช่น `ImportCompleted` ไม่ใช่ database row dump

## 9. Code-reading path ใน Shipment Tracker

Repo: `../go-kafka-shipment-tracker`

1. `internal/kafka/producer.go`
   - ดู `ShipmentID` เป็น key
   - ถามว่า `ProduceSync` แลก throughput กับอะไร
2. `internal/kafka/projector.go`
   - ดู `DisableAutoCommit`
   - ไล่ `processRecord → retry/DLQ → commit`
3. `internal/postgres/store.go`
   - ดู transaction, `processed_events` และ `FOR UPDATE`
4. `internal/domain/shipment/shipment.go`
   - ดู state transition แยกจาก Kafka transport
5. `integration/shipment_flow_test.go`
   - ดูหลักฐาน duplicate/rejected event
6. `integration/retry_dlq_test.go`
   - ดู permanent failure routing
7. `integration/projector_lifecycle_test.go`
   - ดู restart และ consumer-group scaling

## 10. คำถาม scenario พร้อมแนวตอบ

### Consumer process ช้าและ lag เพิ่ม จะทำอย่างไร

วัด processing latency แยก stage ก่อน ดู partition skew, DB pool/locks, external API และ
batching จากนั้นเพิ่ม consumer ได้ไม่เกิน partitions ปรับ partitions เมื่อ key distribution
เหมาะ และใช้ backpressure/rate limit ไม่โยน load ทั้งหมดไป database

### Event เดิมถูกส่งสองครั้งแต่ event ID ต่างกัน

Technical idempotency ด้วย event ID ช่วยไม่ได้ ต้องมี business idempotency key เช่น
`import_job_id + chunk_id` หรือ provider request ID และแก้ producer contract

### DLQ โตขึ้นเรื่อย ๆ

Alert จาก rate/age, แยก schema/validation/dependency causes, หยุด replay จนแก้ root cause,
สร้าง controlled replay ที่รักษา metadata และตรวจ idempotency

### Kafka ล่มหลัง API รับ request

API commit business state + outbox ได้ Publisher retry เมื่อ Kafka กลับมา ผู้ใช้เห็น
PENDING และระบบ monitor outbox backlog

### จะรับประกันว่า ImportCompleted มาครั้งเดียวได้ไหม

ไม่ควรสัญญา end-to-end once ออกแบบให้ duplicate safe ด้วย event ID/business key และ
state transition/version พร้อมบอกขอบเขตของ Kafka transaction

### จะเลือก Kafka แทน database job queue เมื่อไร

เลือก Kafka เมื่อมีหลาย consumers, replay, partition scaling, durable event stream หรือ
decoupling ชัดเจน หากเป็น background job ภายใน service เดียว PostgreSQL queue กับ
`SKIP LOCKED` อาจง่ายและเหมาะกว่า

## 11. คำตอบเชื่อมกับประสบการณ์

รูปแบบที่ตรงไปตรงมา:

> ในงานจริงผมเคยทำ asynchronous data import ที่มี validation, background processing
> และ status tracking แม้ implementation ตอนนั้นไม่ได้ใช้ Kafka แต่ปัญหาเรื่อง bounded
> workers, retry, idempotency และการแยก transient/business error เหมือนกัน ผมจึงทำ
> Shipment Tracker เพื่อฝึก Kafka-specific semantics เช่น partition ordering, manual
> offset commit, consumer group, retry topic และ DLQ

คำตอบนี้แสดงทั้งประสบการณ์จริงและขอบเขตที่เรียนเพิ่มโดยไม่ overclaim
