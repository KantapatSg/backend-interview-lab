# PostgreSQL Real-World Use Cases

## 1. Import staging → validation → merge

อย่าเขียนข้อมูลที่ยังไม่ผ่าน validation ลง business tables โดยตรง

```text
object storage file
  → import_jobs
  → staging_import_rows
  → validate
  → merge/upsert business tables
  → import_results
```

ตัวอย่าง schema:

```sql
CREATE TABLE import_jobs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    status TEXT NOT NULL,
    config_version BIGINT NOT NULL,
    total_rows INTEGER NOT NULL DEFAULT 0,
    processed_rows INTEGER NOT NULL DEFAULT 0,
    failed_rows INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE staging_import_rows (
    job_id UUID NOT NULL REFERENCES import_jobs(id),
    row_number INTEGER NOT NULL,
    raw_data JSONB NOT NULL,
    normalized_data JSONB,
    validation_errors JSONB NOT NULL DEFAULT '[]',
    PRIMARY KEY (job_id, row_number)
);
```

Staging ช่วย audit/retry/debug แต่กิน storage ต้องมี retention policy

## 2. JSONB หรือ relational columns

ใช้ relational columns เมื่อ:

- field มี business constraint/type ชัด
- query/join/filter บ่อย
- ต้อง FK/unique/check constraint

ใช้ JSONB เมื่อ:

- raw payload/audit/config มี shape ยืดหยุ่น
- field แตกต่างตาม integration
- query บางส่วนและยอมรับ validation ใน application/schema layer

อย่าเก็บทุกอย่างใน JSONB เพื่อหลีกเลี่ยง schema design เพราะ index/constraint/migration
และ query จะซับซ้อนขึ้น

## 3. Job claiming ด้วย SKIP LOCKED

เหมาะกับ PostgreSQL-backed worker queue:

```sql
WITH claimed AS (
    SELECT id
    FROM import_jobs
    WHERE status = 'PENDING'
    ORDER BY created_at
    FOR UPDATE SKIP LOCKED
    LIMIT 10
)
UPDATE import_jobs j
SET status = 'RUNNING', updated_at = NOW()
FROM claimed
WHERE j.id = claimed.id
RETURNING j.*;
```

ข้อควรระวัง:

- worker crash แล้ว RUNNING ค้าง ต้องมี lease/heartbeat/reaper
- `SKIP LOCKED` เหมาะกับ queue ไม่เหมาะกับ report ที่ต้องเห็นข้อมูลครบ
- transaction claim ต้องสั้น
- ordering อาจไม่ strict เมื่อ row แรกถูก lock

## 4. Progress update

ห้าม read-modify-write counter แบบไม่มี concurrency guard:

```sql
UPDATE import_jobs
SET processed_rows = processed_rows + $1,
    failed_rows = failed_rows + $2,
    updated_at = NOW()
WHERE id = $3;
```

หากต้อง deduplicate chunk completion ให้มี `completed_chunks(job_id, chunk_id)` unique
แล้ว increment counter เฉพาะเมื่อ insert chunk ใหม่สำเร็จใน transaction เดียวกัน

## 5. Idempotent upsert

```sql
INSERT INTO products (tenant_id, external_id, name, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (tenant_id, external_id)
DO UPDATE SET name = EXCLUDED.name, updated_at = NOW();
```

ต้องนิยามว่า duplicate หมายถึงอะไร Business key ไม่ควรถูกเลือกเพียงเพราะ field ดู unique
และ update ควรระวัง event เก่ามาทับข้อมูลใหม่ด้วย source version/timestamp

## 6. Index ตามหน้าจอ status tracking

Query:

```sql
SELECT id, status, processed_rows, failed_rows, created_at
FROM import_jobs
WHERE tenant_id = $1 AND status = $2
ORDER BY created_at DESC, id DESC
LIMIT 50;
```

Candidate:

```sql
CREATE INDEX import_jobs_tenant_status_created_idx
ON import_jobs (tenant_id, status, created_at DESC, id DESC);
```

พิสูจน์ด้วย `EXPLAIN (ANALYZE, BUFFERS)` และ production-like data distribution
อย่าสร้าง index ต่อทุก column เพราะ write/import cost จะสูงขึ้น

## 7. Keyset pagination

```sql
SELECT ...
FROM import_jobs
WHERE tenant_id = $1
  AND (created_at, id) < ($cursor_time, $cursor_id)
ORDER BY created_at DESC, id DESC
LIMIT 50;
```

Cursor ควร encode ทั้ง sort keys และตรวจ tenant/filter จาก request ใหม่ ไม่เชื่อ cursor
เป็น authorization proof

## 8. Migration ปริมาณมาก

แนวทาง:

- backup/restore rehearsal
- staging + validation report
- batches และ checkpoint
- deterministic transformation
- idempotent rerun
- reconciliation counts/checksums
- dry run และ sample verification
- transaction size จำกัด
- throttle เพื่อลดผลกระทบ production
- audit ว่า source row กลายเป็น target ใด

สำหรับ schema change ใหญ่ใช้ expand/contract:

1. เพิ่ม schema ใหม่ที่ backward-compatible
2. deploy code ที่อ่าน/เขียนได้ทั้งสองแบบตามแผน
3. backfill เป็น batches
4. verify/reconcile
5. switch reads
6. ลบของเก่าภายหลัง

## 9. Deadlock ใน inventory transfer

Transaction A lock SKU-1 แล้ว SKU-2 ส่วน B lock SKU-2 แล้ว SKU-1 ทำให้รอกันเป็นวง

แก้โดย sort resource keys แล้ว lock ลำดับเดียวกันทุก transaction:

```sql
SELECT * FROM inventory
WHERE warehouse_id = $1 AND sku = ANY($2)
ORDER BY sku
FOR UPDATE;
```

ยังต้อง retry transaction ที่ PostgreSQL เลือก abort แบบ bounded

## 10. Isolation decisions

- Read Committed: use case ทั่วไปและ explicit row locking
- Repeatable Read: report/operation ที่ต้อง snapshot คงที่
- Serializable: invariant ซับซ้อนที่ยอม transaction retry ได้

Isolation สูงไม่แทน unique/check/FK constraints และไม่ได้แก้ external side effects

## 11. Connection pool

ถ้ามี 10 service instances และแต่ละตัว max 30 connections เท่ากับอาจเปิด 300 connections
ต้องคิด capacity รวมและ reserve สำหรับ migration/admin/other services

วัด:

- active/idle connections
- acquire wait time
- query latency
- transaction duration
- lock waits
- timeout/error rate

## 12. คำถาม scenario

### Import 10 ล้าน rows จะใช้ transaction เดียวหรือไม่

ไม่ เพราะ WAL/locks/rollback/recovery cost สูง ใช้ staging, batches, checkpoints และ final
state transition/reconciliation ตาม atomicity requirement

### JSONB field query ช้า จะทำอย่างไร

ดู query plan/usage ก่อน เลือก expression/GIN index เมื่อเหมาะ หรือ promote field ที่ใช้บ่อย
เป็น typed column พร้อม migration ไม่สร้าง GIN ทุก payload โดยอัตโนมัติ

### Worker สองตัวหยิบ job เดียวกัน

ใช้ atomic claim ด้วย row lock/lease และ idempotent chunk/business writes Worker crash ต้อง
ทำให้ lease หมดอายุและ job ถูก claim ใหม่อย่างปลอดภัย

### Update status พร้อมกันแล้วค่าถูกทับ

ใช้ state transition condition/version เช่น `WHERE status='PENDING' AND version=$v`
ตรวจ affected rows และไม่ใช้ last-write-wins โดยไม่ตั้งใจ
