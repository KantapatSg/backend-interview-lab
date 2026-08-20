# Coverage Matrix

ไฟล์นี้ตรวจว่าหัวข้อจากแผนฉบับเต็มยังอยู่ครบหลังปรับเป็นแผนเร่งด่วน

| Area | เนื้อหาหลัก | Code/Use case | คำถาม | Priority |
|---|---|---|---|---|
| Go fundamentals | `01-cheat-sheet.md` | `labs/go-context`, `labs/go-worker-pool` | `04-question-bank.md` | ทวน |
| Go production | `06-go-production-review.md` | Import worker, batching, shutdown, review exercises | ท้ายเอกสาร | สูง |
| Microservices | `01-cheat-sheet.md` | Order/Inventory/Payment | `04-question-bank.md` | สูง |
| CQRS | `01-cheat-sheet.md` | `labs/cqrs-basic`, `../goflow-cqrs` | `02-project-walkthrough.md` | สูงมาก |
| Transactional Outbox | `01-cheat-sheet.md` | `../goflow-cqrs`, `labs/postgres` | `02-project-walkthrough.md` | สูงมาก |
| Kafka fundamentals | `01-cheat-sheet.md` | `../go-kafka-shipment-tracker` | `04-question-bank.md` | สูงมาก |
| Kafka production | `05-kafka-deep-dive.md` | Import, inventory, external integration | Scenario section | สูงมาก |
| PostgreSQL | `01-cheat-sheet.md` | `labs/postgres` | `04-question-bank.md` | สูง |
| PostgreSQL production | `08-postgresql-use-cases.md` | staging, SKIP LOCKED, migration, inventory | Scenario section | สูง |
| Next.js fundamentals | `01-cheat-sheet.md` | `labs/nextjs-order-status` | `04-question-bank.md` | กลาง |
| Next.js production | `07-nextjs-deep-dive.md` | `labs/nextjs-import-dashboard` | ท้ายเอกสาร | สูง |
| System design | `03-system-design.md` | Order Platform/Saga | failure cases | สูงมาก |
| Coding review | `06-go-production-review.md` | exercises A–C | review prompts | สูง |
| Behavioral | `04-question-bank.md` | STAR stories | `09-resume-based-prep.md` | สูง |
| Resume alignment | `09-resume-based-prep.md` | import/migration/warehouse/cache | tailored questions | สูงมาก |

## Topics ที่ต้องอธิบายได้โดยวาดจากความจำ

1. Request → DB transaction → outbox → Kafka → idempotent consumer → offset commit
2. Data import upload → object storage → validation → batch write → status projection
3. Inventory reservation → payment → Saga compensation
4. Next.js Server Component → Client Component → Route Handler → Go API
5. Graceful shutdown ของ HTTP server + background workers + Kafka consumer

## Explicit non-goals ก่อนวันสัมภาษณ์

- Kubernetes/service mesh implementation
- Event Sourcing framework เต็มรูปแบบ
- Kafka cluster tuning ระดับ broker internals
- Next.js production deployment setup
- สร้าง microservices capstone ใหม่ซ้ำกับสอง repo ที่มีอยู่

หัวข้อเหล่านี้ตอบเชิง concept ได้ แต่ไม่ควรเบียดเวลาจาก delivery semantics, transaction,
idempotency, failure handling และการเล่าประสบการณ์จริง
