# Question Bank

ฝึกตอบแต่ละข้อ 30–90 วินาที แล้วให้คนถาม follow-up ว่า “ถ้าล้มเหลวตรงนี้ล่ะ?”

## Go

### 1. Goroutine ต่างจาก OS thread อย่างไร

Goroutine ถูก schedule โดย Go runtime ลงบน OS threads มี stack ที่โตได้และเริ่มต้นเบา
กว่า แต่ยังใช้ memory/resource จึงต้องจำกัดจำนวนและมี cancellation path

### 2. Buffered กับ unbuffered channel ต่างกันอย่างไร

Unbuffered ต้องมี sender/receiver พร้อมกัน Buffered ให้ sender ไปต่อได้จนเต็ม เมื่อเต็มจะ
block และสร้าง backpressure Buffer รับ burst ได้แต่ไม่แก้ consumer ช้าถาวร

### 3. ใครควรปิด channel

ฝั่งที่เป็นเจ้าของการส่งและรู้ว่าไม่มี send เพิ่มแล้ว ปกติ receiver ไม่ควรปิด channel
เพราะอาจมี sender อื่นส่งแล้ว panic

### 4. ป้องกัน goroutine leak อย่างไร

กำหนด owner/lifecycle, รับ context, select `ctx.Done()`, ปิด channel อย่างถูกต้อง,
ตั้ง timeout ให้ I/O และทดสอบ shutdown path

### 5. Mutex กับ channel เลือกอย่างไร

Mutex สำหรับป้องกัน shared state สั้น ๆ Channel สำหรับส่ง ownership/work หรือประสาน
lifecycle เลือกสิ่งที่ทำให้ invariant ชัดที่สุด

### 6. ทำไม map จึงเขียนพร้อมกันไม่ได้

Built-in map ไม่มี synchronization Concurrent read/write หรือ write/write ทำให้ race
และอาจ runtime panic ต้องใช้ mutex, `sync.Map` ใน use case เหมาะสม หรือ single owner

### 7. `context.Context` ควรใช้อย่างไร

เป็น parameter แรก ส่งต่อทั้ง call chain ใช้ deadline/cancellation/request metadata
ไม่เก็บใน struct และต้องเรียก cancel ของ derived context

### 8. `%w`, `errors.Is` และ `errors.As` ใช้ทำอะไร

`%w` wrap โดยรักษา chain, `Is` ตรวจ target/sentinel ใน chain และ `As` ดึง typed error

### 9. Pointer receiver กับ value receiver ต่างกันอย่างไร

Pointer receiver แก้ค่าเดิมและหลีกเลี่ยง copy struct ใหญ่ Value receiver ทำงานกับ copy
และ method set ต่างกัน ควรใช้สม่ำเสมอตาม semantics ของ type

### 10. Interface ควรประกาศที่ไหน

มักประกาศใกล้ consumer และให้เล็กตาม method ที่ consumer ต้องใช้ ทำให้ dependency
และ test fake ชัดเจน

### 11. จะทำ graceful shutdown อย่างไร

รับ OS signal, cancel root context, หยุดรับ traffic, shutdown HTTP server ด้วย deadline,
หยุด consumers/workers, flush/close dependencies และอย่าปล่อยงานครึ่ง transaction

### 12. จะจำกัด concurrent requests อย่างไร

ใช้ semaphore/bounded worker pool, server limits, queue capacity และ timeout ต้องกำหนด
behavior เมื่อเต็ม เช่น reject 429/503 ไม่ใช่สร้าง goroutine ไม่จำกัด

## Microservices และ CQRS

### 13. เมื่อไรควรใช้ microservices

เมื่อ business boundaries, team ownership, independent deployment หรือ scaling need
ชัดพอชดเชย network/operations complexity ไม่ใช่เพียงเพราะระบบใหม่

### 14. Modular monolith ต่างจาก microservices อย่างไร

Modular monolith มี boundary ใน code แต่ deploy/process เดียวและทำ local transaction ได้
Microservices แยก deployment/data ownership แต่ต้องรับ distributed failure

### 15. Database per service มีเหตุผลอะไร

รักษา data ownership และ deploy schema แยก ลด coupling แต่ cross-service query และ
transaction ต้องใช้ API/event/read model แทน join ตรง

### 16. CQRS คืออะไร

แยก command ที่เปลี่ยน state จาก query ที่อ่าน เพื่อให้ model และ scaling เหมาะกับงาน
ไม่จำเป็นต้องเป็น microservices หรือ Event Sourcing

### 17. CQRS มีต้นทุนอะไร

หลาย model, projection code, eventual consistency, duplicate/order handling,
schema evolution, monitoring และ integration tests

### 18. Command ควร return อะไร

อย่างน้อย acknowledgement/aggregate ID/version/status จาก write side ไม่ควรพึ่งว่า
read projection พร้อมแล้ว อาจคืน 202 เมื่อ workflow ทำต่อแบบ async

### 19. ถ้า GET หลัง POST แล้วไม่พบข้อมูลทำอย่างไร

คืน ID จาก command, แสดง pending, poll/subscription หรือ read-your-own-write strategy
ตาม requirement พร้อมวัด projection lag

### 20. CQRS ต่างจาก Event Sourcing อย่างไร

CQRS แยก read/write responsibility ส่วน Event Sourcing เก็บ events เป็น source of truth
ใช้แยกกันได้

### 21. Saga คืออะไร

ชุด local transactions ข้าม service เชื่อมด้วย commands/events เมื่อ step ล้มเหลวใช้
compensation เพื่อคืน business state ไม่ใช่ ACID rollback

### 22. Orchestration กับ choreography ต่างกันอย่างไร

Orchestrator เก็บ workflow state กลาง มอง flow ง่ายแต่เป็น dependency สำคัญ Choreography
ให้ services react ต่อ events ลด central coordinator แต่ flow กระจายและ debug ยาก

### 23. Retry ควรทำเมื่อไร

เฉพาะ transient failure และ operation ต้อง idempotent ใช้ max attempts, exponential
backoff, jitter และ timeout Validation/business rejection ไม่ควร retry อัตโนมัติ

### 24. Circuit breaker ช่วยอะไร

หยุดยิง dependency ที่ fail ต่อเนื่อง ลด resource exhaustion และเปิดทางให้ recover
แต่ต้องมี fallback/clear error และไม่แทน timeout/retry budget

### 25. จะ trace request ที่กลายเป็น event อย่างไร

ส่ง correlation ID ตลอด workflow และ causation ID ชี้ event/command ต้นเหตุ แยกจาก
event ID ที่ใช้ identity/idempotency

## Kafka

### 26. Kafka รับประกัน ordering หรือไม่

รับประกันภายใน partition เท่านั้น ใช้ aggregate ID เป็น key หากต้องรักษาลำดับต่อ aggregate

### 27. 10 partitions กับ 20 consumers ใน group เดียวเกิดอะไร

ทำงานพร้อมกันสูงสุด 10 consumers ที่เหลือไม่มี partition assignment

### 28. Consumer group ใช้ทำอะไร

แบ่ง partitions ระหว่าง consumers เพื่อ scale การประมวลผล และแต่ละ group มี offset
ของตัวเองจึงสร้าง projection/consumer use case แยกกันได้

### 29. Commit offset ก่อน process มีผลอย่างไร

ถ้า crash หลัง commit ก่อน process event จะหายจากมุม consumer เป็น at-most-once

### 30. Process แล้วค่อย commit มีผลอย่างไร

ถ้า crash หลัง side effect ก่อน commit event ถูกส่งซ้ำ จึงเป็น at-least-once และต้อง
idempotent

### 31. ทำ idempotent consumer อย่างไร

ใน transaction เดียว insert `(consumer_group,event_id)` แบบ unique แล้วจึงทำ side
effect ถ้า insert conflict ให้ข้ามผลซ้ำ

### 32. Exactly-once ของ Kafka ครอบ PostgreSQL หรือไม่

ไม่อัตโนมัติ Kafka transactions ครอบ Kafka records/offsets External database ต้องใช้
idempotency/outbox หรือ coordination อื่น

### 33. Message key เลือกอย่างไร

เลือก entity ที่ต้อง ordering และกระจาย load เช่น order ID ระวัง low-cardinality/hot key
ที่ทำให้ partition skew

### 34. Rebalance คืออะไร

Kafka เปลี่ยน partition assignment เมื่อ consumer/group/topic เปลี่ยน งาน in-flight และ
offset ต้องจัดการก่อน revoke ไม่เช่นนั้นเกิด duplicate หรือ delay

### 35. Consumer lag คืออะไร

ระยะห่างระหว่าง latest offset กับ consumed/committed offset ใช้บอก backlog แต่ต้องดู
processing latency และ traffic rate ประกอบ

### 36. DLQ ใช้เมื่อไร

event ที่ประมวลผลต่อไม่ได้หลัง policy เช่น malformed/permanent/exhausted retry ต้องมี
metadata, alert, inspection, repair และ replay ไม่ใช่ทิ้งแล้วจบ

### 37. Retry topic มี trade-off อะไร

แยก retry ออกจาก main throughput ได้ แต่เพิ่ม topic/consumer และอาจทำลาย ordering
ต่อ key เมื่อ event หลังจากนั้นเดินต่อ

### 38. เพิ่ม partition มีผลอะไร

เพิ่ม parallelism แต่ key mapping ของ records ใหม่อาจเปลี่ยน, ordering มีแค่ต่อ partition,
rebalancing และ operational cost เพิ่ม

### 39. Kafka ต่างจาก RabbitMQ อย่างไร

Kafka เน้น durable append log, partitioned replay และ consumer-owned offsets RabbitMQ
เน้น queue/routing/ack delivery ควรเลือกจาก workload ไม่ใช่บอกว่าอันหนึ่งดีกว่าทุกกรณี

## PostgreSQL

### 40. MVCC คืออะไร

เก็บหลาย row versions เพื่อให้ transaction เห็น snapshot ลด reader/writer blocking
แต่ยังต้อง vacuum และยังมี write locks/deadlocks

### 41. Read Committed กับ Repeatable Read ต่างกันอย่างไร

Read Committed snapshot ต่อ statement จึงอ่านซ้ำอาจเห็นค่าใหม่ Repeatable Read ใช้
snapshot คงที่ตลอด transaction แต่อาจต้องจัดการ serialization-like conflict บางกรณี

### 42. `SELECT FOR UPDATE` ใช้เมื่อไร

เมื่อต้องอ่านแล้วตัดสินใจ update โดยไม่ให้ transaction อื่นแก้ row เดียวก่อน commit
ต้องถือ lock สั้นและ lock order สม่ำเสมอ

### 43. Optimistic locking ทำงานอย่างไร

อ่าน version แล้ว update ด้วย `WHERE id=? AND version=?` พร้อม increment ถ้า affected
rows เป็นศูนย์แปลว่ามี concurrent writer

### 44. Composite index column order เลือกอย่างไร

อิง equality/range/order predicates, selectivity และ query patterns จริง แล้วพิสูจน์ด้วย
`EXPLAIN ANALYZE`; index `(a,b)` ไม่เท่ากับ `(b,a)`

### 45. ทำไมมี index แล้ว database ยัง sequential scan

table เล็ก, query คืน row จำนวนมาก, statistics/cost estimate หรือ predicate ไม่เข้ากับ
index อาจทำให้ sequential scan ถูกกว่า

### 46. Offset กับ keyset pagination ต่างกันอย่างไร

Offset ใช้ง่ายแต่ scan/skip มากและผลเลื่อนได้ Keyset ใช้ last sort key ทำให้คงที่และเร็ว
กว่าในหน้าลึก แต่ jump ไป page ใดก็ได้ยากกว่า

### 47. Deadlock แก้อย่างไร

lock row/table ในลำดับเดียวกัน, transaction สั้น, ไม่ทำ network call ระหว่าง transaction
และ retry transaction ที่ถูก abort แบบ bounded

### 48. Outbox แก้ปัญหาอะไร

แก้ atomicity ระหว่าง business state กับ intent-to-publish โดยเขียนทั้งคู่ transaction
เดียว Publisher ส่งภายหลัง แต่ยังเกิด duplicate ได้

### 49. Connection pool ใหญ่ที่สุดดีที่สุดหรือไม่

ไม่ Connection มากเกินเพิ่ม contention/memory และรวมทุก instance อาจเกิน database limit
ต้อง size จาก capacity และวัด wait/saturation

## Next.js

### 50. Server Component กับ Client Component ต่างกันอย่างไร

Server Component render ฝั่ง server ใช้ secret/data source ได้และส่ง JS น้อย Client
Component ใช้ state/effect/event/browser API และต้องประกาศ `use client`

### 51. SSR, CSR, SSG และ ISR เลือกอย่างไร

ดู freshness, personalization, SEO, interaction และ server cost ไม่มีแบบเดียวเหมาะทุกหน้า

### 52. ทำไม update แล้ว UI ยังเห็นข้อมูลเก่า

อาจเป็น Next/data/CDN cache หรือ CQRS projection lag ต้องระบุ layer ที่ stale แล้ว
invalidate/revalidate หรือแสดง pending ตาม consistency model

### 53. JWT ควรเก็บ localStorage หรือ cookie

HttpOnly Secure SameSite cookie ลดการถูกอ่านจาก XSS แต่ต้องป้องกัน CSRF LocalStorage
ถูก JS อ่านได้จึงเสี่ยงเมื่อมี XSS เลือกตาม threat model

### 54. ซ่อนปุ่มใน UI ถือว่า authorize แล้วหรือไม่

ไม่ ต้องตรวจ authorization ที่ server ทุก mutation/query ที่มีสิทธิ์ เพราะ client ถูกแก้ได้

### 55. Route Handler กับ Server Action ต่างกันอย่างไร

Route Handler เป็น HTTP API boundary ใช้ได้กับ external/client ทั่วไป Server Action
ผูกกับ React mutation flow มากกว่า แต่ทั้งคู่ต้อง validate/auth/error handling

## Behavioral/System Design

### 56. เล่า incident อย่างไร

Situation → signal/evidence → hypothesis → action → verification → prevention เน้นว่า
วัดอะไรและลด blast radius อย่างไร ไม่เล่าเพียงว่าลอง restart

### 57. ถ้าไม่เห็นด้วยกับ architecture ทำอย่างไร

ทำ requirement/trade-off ให้ชัด เสนอ experiment/ADR ใช้ข้อมูลตัดสิน commit ตามทีมเมื่อ
ตัดสินแล้ว และติดตามผลโดยไม่ยึด ego

### 58. เมื่อไรจะไม่ใช้ CQRS

CRUD ตรงไปตรงมา, scale/read shape ไม่ต่าง, team ยังไม่มี operational capacity หรือ
consistency requirement ไม่คุ้ม projection complexity

### 59. ออกแบบระบบควรเริ่มจากอะไร

Clarify functional/non-functional requirements และ scale ก่อน แล้ว API/data ownership,
happy path, failure path, consistency, scaling, security และ observability

### 60. คำถามสำคัญที่สุดของ stack นี้คืออะไร

เมื่อ database, Kafka และหลาย services commit พร้อมกันไม่ได้ จะรักษา invariant และ recover
อย่างไร คำตอบควรเชื่อม local transaction, outbox, idempotency, Saga, retry, DLQ,
reconciliation และ observability
