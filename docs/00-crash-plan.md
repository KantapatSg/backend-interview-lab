# แผนเร่งด่วนก่อนวันเสาร์

เวลามีน้อย จึงใช้หลัก **เข้าใจ flow → อ่านโค้ดจริง → พูดอธิบาย → ตอบ failure case**
ไม่พยายามสร้างระบบใหม่ทั้งชุด และไม่เริ่มหัวข้อ infrastructure ที่ไม่อยู่ใน stack

## คืนวันพฤหัสบดี 20 สิงหาคม

### Block 1 — Go concurrency (60 นาที)

1. อ่านหัวข้อ Go ใน cheat sheet 20 นาที
2. รัน `go-context` และ `go-worker-pool` 20 นาที
3. ตอบออกเสียง 20 นาที:
   - goroutine ต่างจาก thread อย่างไร
   - buffered channel สร้าง backpressure อย่างไร
   - ป้องกัน goroutine leak อย่างไร
   - context cancellation ส่งต่ออย่างไร

### Block 2 — CQRS และ PostgreSQL (90 นาที)

เปิด `../goflow-cqrs` แล้ว trace ตามลำดับ:

```text
HTTP handler
  → command.Service
  → write.Repository transaction
  → write tables + outbox_events
  → projection.Runner
  → read.Repository
  → GET response
```

ต้องตอบให้ได้:

- ทำไม command กับ query แยกกัน
- ทำไม write model กับ outbox ต้องอยู่ transaction เดียวกัน
- ทำไม `GET` ทันทีหลัง `POST` อาจยังไม่พบข้อมูล
- optimistic concurrency ป้องกัน lost update อย่างไร

### Block 3 — Kafka (90 นาที)

เปิด `../go-kafka-shipment-tracker` แล้ว trace:

```text
Producer
  → topic + shipment_id key
  → consumer group
  → PostgreSQL projection transaction
  → processed_events
  → commit offset
```

ต้องตอบให้ได้:

- ordering มีขอบเขตแค่ไหน
- commit ก่อนหรือหลัง processing ต่างกันอย่างไร
- process สำเร็จแต่ commit ล้มเหลวจะเกิดอะไร
- retry topic และ DLQ ใช้เมื่อไร

### Block 4 — เล่าโปรเจกต์ (30 นาที)

อัดเสียงตัวเองเล่าแต่ละโปรเจกต์ไม่เกิน 3 นาที โดยใช้โครง:

```text
Problem → Architecture → Critical decision → Failure handling → Trade-off
```

## วันศุกร์ 21 สิงหาคม

### ช่วงเช้า — PostgreSQL (90 นาที)

- ACID และ MVCC
- isolation levels
- `SELECT ... FOR UPDATE`
- optimistic locking
- composite index
- `EXPLAIN ANALYZE`
- keyset pagination
- transaction rollback และ deadlock

รันหรืออ่าน SQL ใน `labs/postgres/` แล้วอธิบายว่าแต่ละ constraint ป้องกันอะไร

### ช่วงสาย — Microservices และ CQRS (90 นาที)

- service boundary
- sync vs async communication
- timeout/retry/circuit breaker
- Saga และ compensation
- transactional outbox
- idempotency
- reconciliation
- logs, metrics, traces และ correlation ID

วาด Order → Inventory → Payment บนกระดาษโดยไม่ดูตัวอย่าง

### ช่วงบ่าย — System design (120 นาที)

ซ้อมโจทย์ Order Platform จาก `03-system-design.md` ตามเวลา:

```text
5 นาที   ถาม requirement
5 นาที   API และ data model
10 นาที  component และ happy path
10 นาที  failure cases
5 นาที   scaling และ observability
```

ทำสองรอบ รอบสองต้องอธิบายสั้นและเป็นระบบกว่ารอบแรก

### ช่วงเย็น — Next.js (60 นาที)

เน้นเฉพาะเรื่องที่ backend/full-stack interview มักถาม:

- Server vs Client Component
- SSR/CSR/SSG/ISR
- caching และ revalidation
- Route Handler/Server Action
- HttpOnly cookie, XSS และ CSRF
- eventual-consistency UX

### ช่วงค่ำ — Mock interview (90 นาที)

สุ่มคำถามจาก question bank 20 ข้อ:

- ตอบสั้น 30 วินาที
- รับ follow-up 2 นาที
- หากตอบไม่ได้ ให้เขียนคำตอบใหม่ด้วยภาษาตัวเองเพียง 3–5 บรรทัด

เตรียมประสบการณ์แบบ STAR อย่างน้อย 4 เรื่อง:

1. Bug หรือ incident ที่วิเคราะห์และแก้
2. การออกแบบระบบหรือเลือก trade-off
3. ความเห็นไม่ตรงกับทีม
4. การปรับ performance หรือ reliability

## เช้าวันเสาร์ 22 สิงหาคม

ห้ามเริ่มหัวข้อใหม่

1. อ่าน cheat sheet 30 นาที
2. เล่า GoFlow และ Shipment Tracker อย่างละ 3 นาที
3. วาด outbox, idempotent consumer และ Saga จากความจำ
4. ทบทวนคำถามที่จะถามบริษัท
5. พักก่อนสัมภาษณ์อย่างน้อย 30 นาที

## วิธีตอบเมื่อไม่รู้

อย่าเดา API หรือ claim guarantee ที่ไม่แน่ใจ ใช้รูปแบบ:

> ผมยังไม่แน่ใจรายละเอียดของ API ตัวนั้น แต่ invariant ที่ผมจะรักษาคือ ...
> ผมจะตรวจสอบด้วย ... และออกแบบ fallback เป็น ...

วิธีนี้แสดง reasoning ได้ดีกว่าพยายามตอบชื่อ function ให้ถูกทุกตัว

## คำถามกลับไปยัง interviewer

- ระบบแบ่ง service boundary ตาม domain หรือทีมอย่างไร
- ปัจจุบันใช้ delivery semantics และ retry strategy แบบใด
- ทีมจัดการ schema evolution ของ Kafka events อย่างไร
- มีวิธีวัด consumer lag และตรวจ reconciliation อย่างไร
- CQRS ใช้กับทุก service หรือเฉพาะส่วนที่ read/write model ต่างกันชัดเจน
