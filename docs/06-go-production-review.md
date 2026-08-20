# Go Production Review

เนื้อหานี้ไม่ทวน syntax พื้นฐาน แต่เน้นสิ่งที่ interviewer ใช้แยกคนที่ “เขียน Go ได้”
ออกจากคนที่เข้าใจ lifecycle, concurrency, database และ failure handling

## 1. Background Import Worker

Use case:

```text
Upload request
  → create import job
  → enqueue work
  → bounded workers
  → validate/transform/write batches
  → update progress
```

### สิ่งที่ไม่ควรทำ

```go
// ทุก request สร้าง goroutine โดยไม่มี limit, ownership หรือ shutdown path
go processFile(file)
```

ปัญหา:

- traffic burst สร้าง goroutine/DB connections มากเกิน
- request context จบแต่ goroutine ยังทำต่อ
- deploy แล้วงานถูกตัดกลาง
- error หายเพราะไม่มีผู้รับผิดชอบผลลัพธ์
- memory พุ่งหากอ่านไฟล์ทั้งก้อน

### แนวทาง production

- persist job ก่อน เพื่อไม่ให้ process memory เป็น source of truth
- queue มีความจุจำกัดและ behavior เมื่อเต็มชัดเจน
- worker pool จำกัดตาม DB/object storage capacity
- root context ถูก cancel ตอน shutdown
- checkpoint/progress ทำให้ resume/retry ได้
- stream file และ batch rows แทนอ่านทั้งหมดเข้า memory
- job state transition idempotent

## 2. Context propagation

เส้นทาง:

```text
HTTP request context
  → service.Execute(ctx)
  → repository.Save(ctx)
  → database driver
```

คำถามที่ต้องตอบ:

**background job ควรใช้ request context หรือไม่?**

ถ้างานต้องอยู่ต่อหลัง API ตอบ ไม่ควรผูก lifecycle กับ request context ต้อง persist/enqueue
แล้ว worker ใช้ process root context พร้อม job-specific timeout แต่ยัง copy เฉพาะ metadata
ที่จำเป็น เช่น correlation ID ไม่ copy context ทั้งก้อน

**timeout ควรตั้ง layer ไหน?**

กำหนด request budget ที่ boundary และ downstream อาจตั้ง budget ย่อยที่สั้นกว่า ห้ามแต่ละ
layer สร้าง timeout ยาวใหม่จนเกิน caller deadline

## 3. Worker pool และ backpressure

Worker count ไม่ควรกำหนดจาก CPU อย่างเดียว:

- CPU-bound: ใกล้จำนวน cores เป็นจุดเริ่ม
- I/O-bound: มากกว่า cores ได้ แต่ต้องดู connection pool/rate limits
- DB-heavy: ถูกจำกัดด้วย DB pool, locks และ query cost
- external API: ถูกจำกัดด้วย provider quota

Backpressure chain ที่ดี:

```text
DB ช้า → workers ช้า → queue เต็ม → producer block/reject → caller เห็น overload
```

ไม่ควรซ่อน overload ด้วย buffer ใหญ่มาก เพราะเพียงเลื่อนปัญหาและเพิ่ม latency/memory

## 4. Goroutine ownership checklist

ก่อนเขียน `go fn()` ตอบให้ได้:

1. ใครเป็นคนเริ่ม
2. ใครเป็นคนหยุด
3. รับ cancellation อย่างไร
4. error ส่งกลับที่ไหน
5. ถ้า output ไม่มีคนอ่านจะ block หรือไม่
6. shutdown รอหรือทิ้งงาน
7. จำนวนสูงสุดเท่าไร

## 5. Channel failure modes

- send ไป channel ที่ไม่มี receiver → block/leak
- worker ปิด shared channel → sender อื่น panic
- ไม่ปิด results → receiver range ไม่จบ
- select ไม่มี `ctx.Done()` → shutdown ไม่ได้
- buffer ใหญ่เกิน → memory queue ที่ไม่มี durability
- ใช้ channel แทน mutex ทั้งที่ shared counter ธรรมดา → code ซับซ้อนโดยไม่จำเป็น

## 6. Transaction boundary

Application service ควรเป็นผู้รู้ว่า use case ใดต้อง atomic:

```text
service.CreateImport
  → begin transaction
  → insert import_job
  → insert outbox_event
  → commit
```

Repository method ที่แยก `InsertJob()` และ `InsertEvent()` แต่เปิด transaction คนละรอบ
ทำให้ invariant หลุด ควรมี transaction abstraction หรือ repository operation ที่สะท้อน
atomic use case

อย่าทำ network call ขณะถือ transaction เพราะ:

- lock ถูกถือนาน
- external timeout ทำให้ pool เต็ม
- rollback local database ไม่ rollback external side effect

## 7. Error taxonomy

แยกอย่างน้อย:

- validation/domain error → 400/409 หรือ job FAILED; ไม่ retry
- not found → 404
- unauthorized/forbidden → 401/403
- optimistic conflict → 409/reload/retry ตาม use case
- transient dependency → 503/retryable
- context timeout/cancel → 504/499-like logging ตาม boundary
- internal bug → 500 + trace/log แต่ไม่เปิดรายละเอียดให้ client

Wrap error พร้อม operation context:

```go
return fmt.Errorf("import chunk %s: insert batch: %w", chunkID, err)
```

Log ที่ boundary หนึ่งครั้งพร้อม structured fields หลีกเลี่ยง log ซ้ำทุก layer

## 8. Interface และ dependency design

Interface ที่ดีสะท้อนสิ่งที่ consumer ต้องใช้:

```go
type ImportRepository interface {
    SaveBatch(ctx context.Context, jobID string, rows []Row) error
    UpdateProgress(ctx context.Context, jobID string, completed int) error
}
```

Red flags:

- `Repository` มี 30 methods ทุก aggregate
- interface ถูกสร้างข้าง implementation แต่ไม่มี consumer อื่น
- mock ทุกอย่างจน test implementation detail
- domain import HTTP/GORM/Kafka types

ใช้ fake/in-memory repository ใน unit test และ integration test กับ PostgreSQL สำหรับ
transaction/constraint ที่ fake พิสูจน์ไม่ได้

## 9. Slice, allocation และ large-file processing

สำหรับ data import:

- reuse buffer อย่างระวัง อย่าเก็บ slice ที่ชี้ backing array ซึ่งจะถูกเขียนทับ
- preallocate เมื่อรู้ขนาด batch โดยประมาณ
- flush batch เป็นช่วง ไม่สะสมทั้งไฟล์
- limit row/payload size
- ระวัง `append` ทำให้ reference ไปคนละ backing array
- profile ก่อน optimize ใช้ pprof/benchmarks ไม่เดา

ตัวอย่าง batching:

```go
batch := make([]Row, 0, 500)
for scanner.Scan() {
    batch = append(batch, parse(scanner.Bytes()))
    if len(batch) == cap(batch) {
        if err := repo.SaveBatch(ctx, jobID, batch); err != nil { return err }
        batch = batch[:0]
    }
}
```

ถ้า repository เก็บ slice ไว้หลัง return ต้อง clone เพราะ `batch[:0]` จะ reuse backing array

## 10. HTTP API review

ตรวจ handler ว่า:

- limit request body/file size
- validate content type แต่ไม่เชื่อ filename อย่างเดียว
- authn/authz ก่อนเริ่ม expensive work
- request ID/correlation ID
- timeout ที่เหมาะ
- idempotency key สำหรับ retryable command
- error response ไม่ leak internal detail
- status code: 202 สำหรับ accepted async job, 201 สำหรับ created resource

## 11. Graceful shutdown

ลำดับทั่วไป:

1. รับ SIGTERM/SIGINT
2. mark readiness false/หยุดรับ traffic ใหม่
3. shutdown HTTP server ด้วย deadline
4. cancel consumers/pollers
5. หยุดรับ jobs ใหม่
6. รอ in-flight workers ถึง safe checkpoint
7. commit/return work ให้ queue ตาม semantics
8. close Kafka/DB/logger

ไม่มีลำดับเดียวสำหรับทุกระบบ ต้องบอกว่างาน in-flight ถูก resume/retry อย่างไร

## 12. Configuration-driven backend

จากประสบการณ์ระบบที่ behavior ขึ้นกับ configuration interviewer อาจถาม:

- config ถูก validate ตอน startup หรือ runtime
- version/audit/rollback configuration อย่างไร
- cache invalidation เมื่อ config เปลี่ยน
- schema ของ config และ backward compatibility
- ป้องกัน config ที่ทำให้ query/worker อันตรายอย่างไร
- tenant-specific config แยกสิทธิ์อย่างไร

หลักสำคัญ: configuration เปลี่ยน behavior ได้ จึงเป็น code/data contract รูปแบบหนึ่ง ต้องมี
validation, versioning, observability และ safe default

## 13. Code review exercises

### Exercise A

```go
func Import(ctx context.Context, rows []Row) {
    for _, row := range rows {
        go db.Insert(context.Background(), row)
    }
}
```

ปัญหา:

- unbounded goroutines
- ตัด cancellation ด้วย `Background()`
- ไม่มี error handling
- ไม่มี transaction/batching
- shutdown ไม่รู้ว่างานจบหรือไม่
- DB pool อาจเต็ม

### Exercise B

```go
func GetAll(ctx context.Context) ([]Item, error) {
    rows, _ := db.QueryContext(ctx, "SELECT * FROM items")
    var out []Item
    for rows.Next() { rows.Scan(&item); out = append(out, item) }
    return out, nil
}
```

ปัญหา:

- ไม่ตรวจ query error
- ไม่ `defer rows.Close()`
- ไม่ตรวจ Scan error และ `rows.Err()`
- `SELECT *`
- ไม่มี pagination/limit
- `item` declaration/lifecycle อาจผิด

### Exercise C

```go
tx, _ := db.BeginTx(ctx, nil)
tx.ExecContext(ctx, "UPDATE ...")
callThirdParty()
tx.Commit()
```

ปัญหา:

- ignore errors
- ไม่มี deferred rollback
- network call ขณะถือ lock/connection
- external side effect ไม่ rollback ตาม DB
- context timeout และ ambiguous result ไม่ถูกจัดการ

## 14. คำถามที่น่าถูกถามตามประสบการณ์

- background processing ของคุณจำกัด concurrency อย่างไร
- ถ้า process restart กลาง import จะ resume หรือเริ่มใหม่
- status tracking ป้องกัน stale/out-of-order update อย่างไร
- data migration จำนวนมากทำอย่างไรไม่ให้ memory/transaction ใหญ่เกิน
- configuration เปลี่ยนระหว่าง job ทำงานใช้ version ไหน
- third-party timeout แต่ไม่รู้ว่าสำเร็จหรือไม่ทำอย่างไร
- PostgreSQL connection pool ตั้งจากอะไร
- มีวิธีพิสูจน์ว่า retry ไม่สร้างข้อมูลซ้ำอย่างไร
