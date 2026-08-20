# Resume-Based Interview Preparation

## Positioning ที่เหมาะกับประสบการณ์

### Production experience ที่พูดได้ตรง ๆ

- Go backend และ REST APIs
- PostgreSQL, relational design, SQL Server และ JSONB
- configuration-driven backend workflows
- asynchronous data import, validation, background processing และ status tracking
- data cleaning/migration/transformation/reporting support
- external integrations, MinIO/S3 และ identity/access management
- React + TypeScript, state management และ frontend caching
- Docker, Nginx, GitLab CI/CD, Swagger/Postman

### Knowledge/project experience ที่ควรระบุขอบเขต

- Clean Architecture และ CQRS
- Kafka consumer semantics
- Next.js App Router/Server Components
- Microservices/Saga/Transactional Outbox

รูปแบบคำตอบ:

> Production experience หลักของผมคือ Go, PostgreSQL และ asynchronous backend
> workflows ส่วน Kafka/CQRS ผมต่อยอดผ่านโปรเจกต์ที่จำลอง failure cases จริง เช่น
> duplicate events, manual offset commit, retry/DLQ และ transactional outbox

อย่าพูดว่าใช้ Kafka/Next.js production หากไม่ได้ใช้จริง Interviewer มักเจาะรายละเอียดและ
ความตรงไปตรงมาดูน่าเชื่อถือกว่าการขยาย scope

## Self-introduction ภาษาไทยประมาณ 60–90 วินาที

> ผมเป็น Software Developer ที่เน้น backend มีประสบการณ์มากกว่า 3 ปี งานปัจจุบันใช้ Go
> และ PostgreSQL ในระบบจัดการข้อมูล ผมรับ requirement ที่ค่อนข้างซับซ้อนมาออกแบบ
> relational model และ backend workflow จุดที่ทำบ่อยคือ data import ที่มี validation,
> background processing, status tracking รวมถึง data migration และ third-party
> integrations ก่อนหน้านี้ทำ full-stack ด้วย React และ TypeScript ในระบบ workflow และ
> warehouse หลายประเภท ช่วงหลังผมลงลึกเรื่อง CQRS, transactional outbox และ Kafka
> consumer semantics เพื่อให้เข้าใจการออกแบบ asynchronous systems ที่ reliable มากขึ้น

## English introduction แบบใช้ประโยคไม่ซับซ้อน

> I am a backend-focused software developer with more than three years of experience.
> My main stack is Go and PostgreSQL. In my current role, I translate business workflows
> into relational data models and backend services. I have worked on asynchronous data
> imports, validation, background processing, status tracking, data migration, and
> external integrations. I also have full-stack experience with React and TypeScript.
> Recently, I have been strengthening my knowledge of CQRS and Kafka by building small
> projects that handle duplicate events, retries, dead-letter queues, and transactional
> outbox patterns.

## Story 1: Asynchronous Data Import

ใช้โครงนี้และเติมข้อมูลจริง:

### Situation

ผู้ใช้ต้องนำเข้าข้อมูลจากไฟล์ ซึ่งมี validation/transformation หลายขั้นและไม่ควร block
HTTP request จนเสร็จ

### Task

ออกแบบ backend workflow ที่ติดตามสถานะ แจ้ง validation error และรองรับข้อมูลปริมาณ
`[ใส่จำนวนจริงถ้าเปิดเผยได้]`

### Action

- สร้าง job/status model
- แยก request acceptance จาก background processing
- validate และ transform ก่อน merge business data
- batch database operations
- เก็บ error ที่อ้างกลับไปยัง row ต้นทางได้
- จำกัด concurrency และจัดการ retry ตามชนิด error
- ทำ frontend status tracking

เลือกเฉพาะสิ่งที่ทำจริง อย่าเติม Kafka หาก implementation จริงไม่ได้ใช้

### Result

ใช้ผลลัพธ์ที่วัดได้จริง เช่นเวลาลดลง, manual work ลดลง, error trace ง่ายขึ้น หรือรองรับ
ข้อมูลมากขึ้น ถ้าเปิดเผยตัวเลขไม่ได้ให้พูดผลเชิงคุณภาพและเหตุผลเรื่อง confidentiality

### Follow-ups

- process restart กลางงานทำอย่างไร
- duplicate upload ป้องกันหรือไม่
- partial success อนุญาตไหม
- transaction ครอบแค่ไหน
- status update concurrent หรือไม่
- file ขนาดใหญ่จัดการ memory อย่างไร

## Story 2: Data Migration

ประเด็นที่ interviewer ต้องการฟัง:

- source/target schema ต่างกันอย่างไร
- mapping และ data quality rules
- dry run/staging
- validation/reconciliation
- idempotent rerun
- rollback/cutover
- audit และข้อมูลที่ผิดจัดการอย่างไร

คำตอบตัวอย่าง:

> ผมเริ่มจาก profile source data และกำหนด mapping/validation rules ก่อนเขียนลง target
> ผมใช้ staging เพื่อแยก raw data จาก business tables ทำ migration เป็น batches และเก็บ
> source identifier เพื่อ rerun/reconcile ได้ หลัง migration เปรียบเทียบ counts, required
> fields และ business totals ไม่ดูเพียงว่า script จบโดยไม่มี error

## Story 3: Configuration-Driven Backend

ต้องอธิบายว่า configuration ช่วย reuse workflow อย่างไร และมี safety อย่างไร:

- schema validation
- versioning
- audit history
- default/fallback
- compatibility ระหว่าง config กับ code version
- cache invalidation
- tenant isolation
- rollback

คำถามเจาะ:

> ถ้า configuration เปลี่ยนระหว่าง import job ทำงานจะใช้ค่าใหม่หรือค่าเดิม?

คำตอบที่แข็งแรงคือ snapshot/store `config_version` ตอนสร้าง job เพื่อให้ผล deterministic
หรืออธิบาย policy ที่ใช้จริง

## Story 4: Warehouse/Internal Systems

ระบบ asset, purchase, picking/packing, inventory transfer, goods receipt และ FOC สามารถ
ใช้แสดงความเข้าใจ business workflows ได้

เลือกหนึ่งระบบที่คุณเข้าใจลึกที่สุดแล้วเตรียม:

- state machine
- actors/permissions
- main tables และ constraints
- transaction boundary
- duplicate/concurrent request
- audit trail
- failure/rollback
- reporting query/index

ตัวอย่าง inventory transfer:

```text
DRAFT → SUBMITTED → APPROVED → DISPATCHED → RECEIVED
                    ↘ REJECTED
```

อธิบายว่า transition ใดทำให้ stock เปลี่ยน และป้องกัน negative inventory อย่างไร

## Story 5: Frontend Caching

Resume ระบุว่าเคยออกแบบ caching เพื่อลด redundant API calls เตรียมตอบ:

- cache key สร้างจากอะไร
- TTL/invalidation
- stale data ยอมได้เท่าไร
- cache อยู่ browser/state library/server
- mutation แล้ว update/invalidate อย่างไร
- cache แยก user/tenant หรือไม่
- วัดว่า API calls/loading ลดลงอย่างไร

เชื่อมไป Next.js:

> ประสบการณ์เดิมทำให้ผมเข้าใจ cache key และ invalidation ฝั่ง React ส่วน Next.js เพิ่ม
> server/data cache และ revalidation ที่ต้องระบุ layer ให้ชัด โดยเฉพาะเมื่อ backend เป็น
> eventual-consistent read model

## คำถามที่คาดจาก resume

### “Designed relational database” — ให้ยกตัวอย่าง

เตรียม entity/relationship จริงหนึ่งชุด อธิบาย normalization, keys, constraints, indexes,
transaction และ query patterns ไม่ตอบเพียงว่ามีกี่ tables

### ทำไมใช้ JSONB

บอก field ที่ shape ยืดหยุ่นหรือ raw integration payload และ field ใดควรเป็น typed columns
พร้อม index/query trade-off

### Background processing ใช้อะไร

บอก implementation จริงก่อน แล้วอธิบาย ownership, persistence, concurrency, retry,
status และ shutdown หากถาม Kafka จึงเปรียบเทียบส่วนที่ Kafka เพิ่ม

### External integration ล้มเหลวทำอย่างไร

timeout, retry classification, idempotency, provider reference, rate limit, logging โดยไม่
เก็บ secret และ manual reconciliation

### Clean Architecture ใช้จริงอย่างไร

อธิบาย dependency direction และ business rule ที่ไม่ผูก HTTP/database ไม่ตอบชื่อ folders
อย่างเดียว ยก code จาก `goflow-cqrs/internal/domain/job.go`

### CQRS ช่วยอะไรในระบบของคุณ

หากเป็น project study ให้บอกตรง ๆ แล้วใช้ GoFlow อธิบาย write model/outbox/projection
และพูดว่า CRUD workflow เล็ก ๆ ไม่จำเป็นต้อง CQRS

### ทำไมสนใจเปลี่ยนงาน

โฟกัสการเติบโตด้าน backend/distributed systems, scope และทีม หลีกเลี่ยงโจมตีนายจ้างเดิม

## คำถาม confidentiality

สำหรับ confidential project:

- อธิบาย problem class และ architecture pattern
- เปลี่ยนชื่อ domain/entity
- ไม่บอก client, dataset, endpoint, credential หรือ business-sensitive metrics
- ใช้ช่วง/สัดส่วนเมื่อเปิดเผยตัวเลข exact ไม่ได้
- บอกชัดว่ารายละเอียดบางส่วนเปิดเผยไม่ได้แล้วกลับมาที่ technical decision

ตัวอย่าง:

> ผมเปิดเผย domain data และ client detail ไม่ได้ แต่สามารถอธิบาย architecture ได้ว่าเป็น
> configuration-driven import workflow บน Go/PostgreSQL ที่แยก validation, background
> processing และ status projection

## ช่องที่ต้องเติมคืนนี้

เขียนคำตอบจริงของคุณเอง:

1. Import ใหญ่ที่สุดประมาณกี่ rows/ไฟล์ และใช้เวลาก่อน-หลังเท่าไร
2. Worker/background mechanism ที่ใช้จริงคืออะไร
3. เคยเกิด duplicate/retry/partial failure แบบไหน
4. PostgreSQL query/index ที่เคย optimize
5. Migration ตรวจความถูกต้องด้วย metric ใด
6. Frontend cache ลด API calls/loading อย่างไร
7. Incident ที่มี evidence และ verification ชัดที่สุด
8. ระบบ warehouse ตัวใดที่คุณอธิบาย state machine ได้ลึกที่สุด
