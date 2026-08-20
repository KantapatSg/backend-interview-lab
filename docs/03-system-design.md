# System Design Drill: Order Platform

## 1. เริ่มด้วย requirement

ถามก่อนวาดระบบ:

- ปริมาณ order ต่อวินาทีและ peak เท่าไร
- ต้อง reserve inventory แบบห้าม oversell หรือยอม reconcile ได้
- Payment authorization ใช้เวลานานเท่าไรและ provider retry ได้หรือไม่
- ผู้ใช้ต้องเห็นสถานะทันทีหรือยอมรับ `PROCESSING`
- ต้องรองรับ cancel/refund หรือไม่
- ordering ต้องต่อ order, user หรือ SKU
- retention/audit/compliance requirement คืออะไร

สมมติฐานสำหรับคำตอบนี้:

- ห้ามสร้าง order ซ้ำจาก client retry
- ห้าม charge ซ้ำ
- inventory และ payment เป็นคนละ service/database
- workflow asynchronous และผู้ใช้เห็นสถานะ pending ได้

## 2. API

```http
POST /orders
Idempotency-Key: 7f42...

GET /orders/{orderId}
POST /orders/{orderId}/cancel
```

Command response:

```json
{
  "orderId": "ord-123",
  "status": "PROCESSING"
}
```

ใช้ unique constraint `(customer_id, idempotency_key)` และเก็บ response/result เดิม
เพื่อให้ request ซ้ำได้ผลเดิม

## 3. Data ownership

- Order Service: order state และ workflow status
- Inventory Service: stock และ reservations
- Payment Service: payment attempt/provider reference
- Order Query Model: denormalized view สำหรับหน้ารายละเอียด/รายการ

แต่ละ service เป็นเจ้าของ database ของตน ห้าม service อื่น join table โดยตรง

## 4. Happy path

```text
Next.js
  → POST /orders + Idempotency-Key
  → Order Service transaction
       ├── INSERT order(PROCESSING)
       └── INSERT outbox(OrderCreated)
  → Outbox Publisher
  → Kafka [key = order_id]
  → Inventory reserves stock
  → InventoryReserved
  → Payment authorizes payment
  → PaymentAuthorized
  → Order becomes CONFIRMED
  → projection updates Order View
  → Next.js polls/subscribes and displays CONFIRMED
```

Event envelope:

```json
{
  "eventId": "uuid",
  "eventType": "OrderCreated",
  "aggregateId": "ord-123",
  "aggregateVersion": 1,
  "occurredAt": "2026-08-20T10:00:00Z",
  "schemaVersion": 1,
  "correlationId": "uuid",
  "causationId": "uuid",
  "payload": {}
}
```

## 5. Saga และ compensation

ตัวอย่าง orchestration:

```text
ReserveInventory
  ├── failed  → CancelOrder
  └── success → AuthorizePayment
                  ├── failed  → ReleaseInventory → CancelOrder
                  └── success → ConfirmOrder
```

Compensation ไม่ใช่ database rollback และอาจล้มเหลว จึงต้อง:

- ใช้ idempotent command
- retry แบบ bounded
- เก็บ saga state/step
- alert เมื่อค้าง
- reconciliation ตรวจ order ที่อยู่ `PROCESSING` นานเกิน threshold

## 6. Transaction และ idempotency

ทุก consumer ทำงานคล้ายกัน:

```text
BEGIN
  INSERT processed_events(event_id) ON CONFLICT DO NOTHING
  ถ้าเป็น event ใหม่:
      validate transition
      update local state
      insert local outbox event
COMMIT
commit Kafka offset
```

นี่ทำให้ local database side effect เป็น idempotent ภายใต้ at-least-once delivery

Payment ต้องส่ง idempotency key ที่ stable ไป payment provider เช่น `payment_attempt_id`
เพื่อให้ timeout แล้ว retry ไม่ charge ซ้ำ

## 7. Ordering และ concurrency

- Kafka key เป็น `order_id` เพื่อเรียง event ต่อ order
- aggregate version ช่วย detect gap/out-of-order
- inventory ของ SKU เดียวอาจเป็น hot row ต้องใช้ atomic conditional update หรือ reservation
- optimistic version เหมาะกับ conflict ต่ำ; row lock เหมาะกับ critical high-contention section

ตัวอย่าง stock update:

```sql
UPDATE inventory
SET available = available - $quantity
WHERE sku = $sku AND available >= $quantity;
```

ถ้า affected rows เป็น 0 แปลว่า stock ไม่พอหรือ record ไม่มีอยู่

## 8. Failure cases

### Database สำเร็จ แต่ Kafka ใช้งานไม่ได้

Outbox event ยังอยู่ใน database Publisher retry ภายหลัง และ monitor oldest outbox age

### Kafka ส่ง event ซ้ำ

Consumer ใช้ `(consumer_group, event_id)` unique key ใน transaction เดียวกับ side effect

### Consumer crash หลัง database commit

event ถูกส่งซ้ำ แต่ idempotency ป้องกัน side effect ซ้ำ

### Event ผิด schema

ส่ง DLQ พร้อม original metadata, alert operator และมี controlled replay หลังแก้

### Payment provider timeout

ยังไม่รู้ว่าสำเร็จหรือไม่ ห้ามสร้าง payment attempt ใหม่ทันที ใช้ provider idempotency key
แล้ว query/reconcile status

### Projection ช้า

Command ยังสำเร็จได้ UI แสดง pending, วัด consumer lag และ scale partition/consumer
โดยตรวจว่า PostgreSQL รองรับ throughput ได้

### Event มาผิดลำดับ

ตรวจ aggregate version เก็บ/retry gap หรือ reject ตาม policy อย่า apply transition ที่ผิด invariant

## 9. Scaling

- Stateless APIs scale horizontal
- Kafka partitions เพิ่ม consumer parallelism
- Read replicas/cache ใช้กับ query ที่ยอมรับ staleness
- Keyset pagination สำหรับรายการใหญ่
- Batch outbox publish แต่รักษา ordering ต่อ aggregate
- แยก hot key และระวัง partition skew
- ตั้ง database connection pool จาก capacity รวมของทุก instance

## 10. Observability

Metrics สำคัญ:

- API request rate/error/latency
- order success/failure/processing duration
- outbox backlog และ oldest event age
- Kafka consumer lag
- retry/DLQ rate
- saga stuck count
- payment/inventory compensation failures
- database pool saturation และ lock wait

Logs และ traces ต้องมี order ID, event ID, correlation ID และ causation ID
แต่ไม่ควร log payment secrets หรือข้อมูลส่วนบุคคลเกินจำเป็น

## 11. Trade-off summary

ระบบนี้เลือก availability และ decoupling ผ่าน asynchronous workflow แลกกับ eventual
consistency และ operational complexity หาก scale/domain ยังเล็ก ควรเริ่ม modular monolith
กับ PostgreSQL outbox ก่อน แล้วแยก service เมื่อ boundary และความต้องการ scale ชัดเจน
