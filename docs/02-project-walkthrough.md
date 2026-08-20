# Project Walkthrough

ใช้สอง repo ที่มีอยู่เป็นหลักฐานว่าคุณเข้าใจแนวคิดจาก code จริง ไม่ควรเล่าทุกไฟล์
ให้เลือก decision สำคัญและอธิบาย failure case

## GoFlow CQRS

ตำแหน่ง: `../goflow-cqrs`

### Pitch ประมาณ 90 วินาที

> GoFlow เป็น modular monolith ใน Go ที่ผมใช้ศึกษา CQRS โดยแยก command และ
> query model แต่ยัง deploy เป็น service เดียวเพื่อไม่เพิ่ม distributed complexity
> เร็วเกินไป Command เปลี่ยน domain aggregate แล้ว repository บันทึก write model
> กับ outbox event ภายใน PostgreSQL transaction เดียวกัน จากนั้น projector แบบ
> asynchronous claim event ด้วย `FOR UPDATE SKIP LOCKED` และอัปเดต read model
> Projector มี receipt ป้องกัน duplicate ส่วน aggregate ใช้ version สำหรับ optimistic
> concurrency ผลคือ query scale และออกแบบ model แยกได้ แต่ต้องยอมรับ eventual
> consistency และดูแล projection lag/retry

### Trace code ตามนี้

1. `internal/adapter/httpapi/router.go` — HTTP boundary และ error mapping
2. `internal/app/command/service.go` — use-case orchestration
3. `internal/domain/job.go` — state machine และ invariant
4. `internal/adapter/postgres/write/repository.go` — transaction และ optimistic version
5. `migrations/001_init.sql` — write/read/outbox/receipt schema
6. `internal/projection/runner.go` — claim, worker และ idempotent projection
7. `internal/adapter/postgres/read/repository.go` — query-side SQL
8. `cmd/api/main.go` — dependency wiring และ graceful shutdown

### จุดที่ควรเปิดให้ interviewer ดู

- `SaveJob`: update แบบตรวจ version และเขียน event ใน transaction เดียวกัน
- `Runner.claim`: `FOR UPDATE SKIP LOCKED` ช่วย projector หลาย worker ไม่ claim ซ้ำ
- `projection_receipts`: event เดิม apply ซ้ำแล้วไม่เปลี่ยนผลลัพธ์
- `signal.NotifyContext` และ `server.Shutdown`: graceful lifecycle

### Trade-offs ที่ต้องพูดเองก่อนถูกถาม

- ใช้ GORM write side แต่ `database/sql` read side เพื่อให้ command persistence
  สะดวกและ query projection ชัดเจน แลกกับสอง data-access styles
- Outbox ยังเป็น PostgreSQL poller ไม่ใช่ Kafka เหมาะกับ first slice แต่ throughput
  และ decoupling มีเพดาน
- มี CQRS แต่ไม่ใช่ Event Sourcing; write tables ยังเป็น source of truth
- Read model อาจตามไม่ทัน จึงต้องวัด lag และออกแบบ UX/API สำหรับ pending state

### Follow-up ที่คาดว่าจะถูกถาม

**ถ้า projector crash หลัง update read model แต่ก่อน mark event processed?**

การ apply และ insert receipt/mark processed ควรอยู่ transaction เดียวกัน หาก event
ถูก claim ใหม่ receipt ทำให้ operation idempotent

**ทำไมต้องรักษาลำดับต่อ aggregate?**

ถ้า `JobCompleted` ถูก apply ก่อน `JobStarted` read model อาจผิด state จึง claim เฉพาะ
event ที่ไม่มี version ก่อนหน้าค้างอยู่สำหรับ aggregate เดียวกัน

**ทำไมไม่แยกเป็น microservices เลย?**

CQRS เป็น model separation ไม่จำเป็นต้อง deploy แยก Modular monolith ช่วยพิสูจน์
boundary และ invariant ก่อนรับ network/operations complexity

## Go Kafka Shipment Tracker

ตำแหน่ง: `../go-kafka-shipment-tracker`

### Pitch ประมาณ 90 วินาที

> Shipment Tracker รับ shipment events ผ่าน Kafka และสร้าง PostgreSQL read model
> Producer ใช้ shipment ID เป็น key เพื่อรักษา ordering ต่อ shipment Consumer ใช้
> manual commit และประมวลผลแบบ at-least-once ใน transaction เดียวกัน โดย insert
> event ID ลง processed_events ก่อนเปลี่ยน projection ถ้า event ถูกส่งซ้ำ unique key
> ทำให้ไม่มี database side effect ซ้ำ Invalid event ไป DLQ ส่วน dependency failure
> ไป bounded retry topic เมื่อ database commit สำเร็จจึงค่อย commit Kafka offset
> ผมไม่เรียกสิ่งนี้ end-to-end exactly-once เพราะ PostgreSQL อยู่นอก Kafka transaction

### Trace code ตามนี้

1. `internal/kafka/producer.go` — key และ synchronous publish result
2. `internal/kafka/projector.go` — poll, route error และ manual commit
3. `internal/postgres/store.go` — database transaction/idempotency/row lock
4. `internal/domain/shipment/shipment.go` — state transition
5. `internal/postgres/schema.sql` — projection/history/processed/rejected tables
6. `integration/shipment_flow_test.go` — happy path และ duplicate proof
7. `integration/retry_dlq_test.go` — permanent error ไป DLQ
8. `integration/projector_lifecycle_test.go` — restart และ consumer-group scaling

### จุดสำคัญ

- `Key: []byte(event.ShipmentID)` — ordering scope ถูกเลือกตาม aggregate
- `kgo.DisableAutoCommit()` — application ควบคุมเวลาที่ offset ถูก commit
- `processed_events` primary key `(consumer_group, event_id)` — idempotency เป็นของ
  logical consumer ไม่ใช่ global event lock
- `SELECT ... FOR UPDATE` — serialize transition ของ shipment เดียวกันใน database
- publish retry/DLQ สำเร็จก่อน commit source record — ลดโอกาส event หาย

### Trade-offs ที่ต้องบอก

- Retry topic ที่ทำอยู่ยังไม่มี delayed scheduler/exponential delay จริง
- การส่งไป retry topic ทำให้ strict ordering ของ shipment อาจเปลี่ยน
- `ProduceSync` เข้าใจง่ายและรู้ผลแน่นอน แต่ throughput ต่ำกว่า async batching
- State projection ใช้ row lock จึงถูกต้องง่าย แต่ hot shipment อาจเกิด contention

### Follow-up สำคัญ

**ถ้า database commit สำเร็จแต่ offset commit ล้มเหลว?**

Kafka ส่ง event ซ้ำ `processed_events` ตรวจพบ event ID เดิมและ commit transaction
โดยไม่ apply side effect ซ้ำ จากนั้น consumer commit offset อีกครั้ง

**ถ้า publish retry สำเร็จแต่ commit source offset ล้มเหลว?**

source record อาจถูกอ่านซ้ำและสร้าง retry record ซ้ำ ดังนั้น retry consumer ก็ยังต้อง
idempotent ด้วย event ID ไม่ควรสมมติว่า retry topic ไม่มี duplicate

**ทำไมจำนวน consumer เพิ่มแล้วไม่เร็วขึ้นเสมอ?**

Consumer group parallelism ถูกจำกัดด้วยจำนวน partitions และปลายทาง PostgreSQL
อาจเป็น bottleneck เพิ่ม consumer เกิน partition จะมี consumer ว่าง

## สิ่งที่ไม่ควร claim

- อย่าบอกว่า Kafka guarantee exactly-once ถึง PostgreSQL
- อย่าบอกว่า outbox ไม่มี duplicate
- อย่าบอกว่า CQRS ต้องเป็น microservices
- อย่าบอกว่า DLQ คือจุดจบ; ต้องมี alert, inspect, repair และ replay process
- อย่าบอกว่า retry แก้ทุก error; validation/business error มัก retry ไม่ช่วย
