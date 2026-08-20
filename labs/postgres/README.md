# PostgreSQL Patterns

ไฟล์นี้เป็น SQL lab สำหรับอ่านและทดลองทีละ transaction ใน `psql` เน้น invariant
ที่มักถูกถาม ไม่ได้เป็น schema ของระบบ production เต็มรูปแบบ

## สิ่งที่แต่ละ constraint ป้องกัน

- `orders.id` primary key — identity ซ้ำ
- `(customer_id, idempotency_key)` unique — client retry สร้าง order ซ้ำ
- `version` — optimistic concurrency/lost update
- `outbox_events.event_id` — event identity
- `(consumer_group, event_id)` — consumer side effect ซ้ำ
- partial outbox index — publisher หาเฉพาะงานค้างได้เร็ว

อ่าน `schema.sql` ก่อน แล้วตามด้วย `transactions.sql`

คำถามที่ต้องตอบ:

1. ทำไม `orders` และ `outbox_events` ต้อง insert ใน transaction เดียวกัน
2. ทำไม outbox ยังต้องมี idempotent consumer
3. affected rows เป็นศูนย์ใน optimistic update หมายถึงอะไร
4. ทำไมไม่ควร call Kafka/payment provider ขณะถือ database transaction
