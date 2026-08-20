# Next.js Order Status Reading Lab

ตัวอย่างนี้ตั้งใจให้สั้นเพื่ออ่านก่อนสัมภาษณ์ ไม่ได้รวม `package.json` หรือสร้าง app ใหม่
เพราะ backend interview เน้นการอธิบาย server/client boundary และ consistency มากกว่า setup

Flow:

1. `page.tsx` เป็น Server Component และ fetch initial state โดยไม่ส่ง backend secret ไป client
2. `OrderStatus.tsx` เป็น Client Component เพราะต้องมี timer/state
3. เมื่อ CQRS read model ยังไม่พร้อม API คืน `PROCESSING` หรือ `404` ชั่วคราว
4. Client poll แบบ bounded จนเจอ terminal state ไม่ poll ตลอดไป

Production ต้องเพิ่ม authentication, authorization, abortable fetch, backoff และ UI error
ตาม requirement
