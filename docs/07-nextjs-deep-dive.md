# Next.js Deep Dive for a Backend-Focused Interview

เป้าหมายคืออ่าน code แล้วอธิบาย boundary, data flow, caching, auth และ failure handling
ไม่จำเป็นต้องจำ option ทุกตัวของ framework

## 1. Mental model ของ App Router

```text
Browser
  → Server Component / Route Handler / Server Action
  → Go Backend
  → PostgreSQL / Kafka
```

Server Component เป็น default และ render ฝั่ง server Client Component ใช้เมื่อมี state,
event handler, effect หรือ browser API

หลักสำคัญ: วาง `"use client"` ให้ต่ำที่สุดใน component tree เพื่อไม่ส่ง JavaScript และ
dependencies ที่ไม่จำเป็นไป browser

## 2. Use case: Data Import Dashboard

```text
Import list page       Server Component
Upload form            Client Component
BFF upload endpoint    Route Handler
Initial job detail     Server Component
Live progress          Client Component polling/SSE
Auth/session           Server-only cookie/token handling
```

Flow:

1. User เลือกไฟล์ใน browser
2. Client form ส่ง multipart request ไป Route Handler
3. Route Handler อ่าน HttpOnly session, validate basic request และ forward ไป Go API
4. Go API สร้าง async import job แล้วคืน `202 Accepted + jobId`
5. Next.js redirect ไป status page
6. Server Component fetch initial status
7. Client Component poll เฉพาะเมื่อ status ยังไม่ terminal

## 3. Server Component

เหมาะกับ:

- fetch initial data ใกล้ server
- access server environment/secrets
- render content ที่ไม่ต้องมี browser interaction
- ลด client bundle

ไม่ควร:

- ใช้ `useState`, `useEffect`, event handlers
- import browser-only package
- ส่ง secret/token เป็น props ไป Client Component

ตัวอย่าง:

```tsx
export default async function ImportPage({ params }) {
  const job = await getImportJob(params.id); // server-only data access
  return <ImportProgress initial={job} />;    // ส่งเฉพาะ safe DTO
}
```

## 4. Client Component

ใช้สำหรับ:

- file input และ upload progress
- local form state
- polling/SSE/WebSocket
- optimistic UI
- browser APIs

Client Component ไม่ได้แปลว่า page ต้อง fetch ทุกอย่างใหม่ใน browser สามารถรับ initial
data จาก Server Component แล้วเพิ่ม interaction เฉพาะจุด

## 5. Rendering choices

### CSR

เหมาะกับ highly interactive private dashboard ที่ SEO ไม่สำคัญ แต่ต้องจัด loading/error
และ browser waterfall

### SSR/dynamic rendering

เหมาะกับ request-specific/authorized/fresh data Server cost สูงขึ้นและต้องดู downstream
latency

### SSG

เหมาะกับ public content ที่เปลี่ยนน้อย เช่น documentation ไม่เหมาะกับ user-specific import
status

### ISR/revalidation

เหมาะกับ content ที่ยอม stale ชั่วคราวและ rebuild เป็นช่วง ไม่ใช่ guarantee ว่า CQRS
projection พร้อมแล้ว

## 6. Caching ต้องระบุ layer

เมื่อ UI เก่า ให้ถามว่า stale ที่ไหน:

```text
Browser cache
Next.js request/data cache
CDN
Go API cache
CQRS read projection
PostgreSQL replica
```

`cache: "no-store"` ทำให้ Next fetch fresh ต่อ request แต่ถ้า read model ตาม Kafka ไม่ทัน
ข้อมูลก็ยัง stale ได้ จึงต้องแยก framework caching ออกจาก domain consistency

หลัง mutation:

- revalidate path/tag เมื่อ cached server data เปลี่ยน
- update local state เมื่อ response มีข้อมูลพอ
- สำหรับ async workflow แสดง PENDING และ poll/subscription
- อย่าทำ optimistic CONFIRMED ถ้า backend ยังเพียง accepted command

## 7. Route Handler เป็น BFF

ข้อดี:

- token/session อยู่ server
- browser เห็น same-origin endpoint
- normalize backend errors/DTO
- centralize request limits และ correlation ID

ข้อเสีย:

- เพิ่ม network hop
- อาจ duplicate validation/business logic
- ต้องดู timeout/body streaming ไม่ buffer file ใหญ่โดยไม่จำเป็น

Route Handler ควร validate transport-level concerns แต่ business invariant อยู่ Go service

## 8. Server Action เทียบ Route Handler

Server Action เหมาะกับ mutation จาก React form ที่ผูกกับ app เดียวและต้องการ revalidation
integration Route Handler เหมาะกับ explicit HTTP contract, file upload, external/mobile client
หรือ endpoint ที่ต้องทดสอบแยก

ทั้งสองแบบต้อง:

- authenticate
- authorize
- validate untrusted input
- handle timeout/error
- ไม่เชื่อ hidden fields
- ป้องกัน CSRF ตาม session/cookie design

## 9. Authentication กับ Keycloak/OIDC

Flow โดยย่อ:

```text
Browser → Authorization Code + PKCE → Identity Provider
        → server/session callback
        → HttpOnly session cookie
        → Route Handler/Server Component
        → Go API พร้อม access token หรือ trusted session
```

คำถามสำคัญ:

- token refresh อยู่ server หรือ browser
- access/refresh token เก็บที่ไหน
- logout/revocation ทำอย่างไร
- role/permission mapping อยู่ layer ไหน
- backend validate issuer, audience, signature และ expiry หรือไม่
- tenant/row-level authorization ตรวจทุก request หรือไม่

ซ่อน navigation item ไม่ใช่ authorization ต้องตรวจ Go API ด้วย

## 10. Upload security

- limit body/file size ที่ proxy, Next และ Go
- validate MIME/content ไม่เชื่อ extension
- sanitize filename และสร้าง object key เอง
- scan malware ตาม risk
- upload ตรง object storage ด้วย presigned URL เมื่อไฟล์ใหญ่
- อย่าอ่านไฟล์ใหญ่ทั้งก้อนเข้า memory ใน Route Handler
- authorize tenant/job ownership
- checksum เพื่อ detect corruption/duplicate

สำหรับ presigned upload:

```text
Browser → request upload intent
Next/Go → create job + presigned URL
Browser → upload direct to object storage
Browser → notify upload complete
Backend → verify object/checksum → enqueue import
```

## 11. Polling, SSE หรือ WebSocket

### Polling

ง่ายและ robust เหมาะกับ update ทุก 1–5 วินาที ต้องมี bounded attempts/backoff และหยุดเมื่อ
tab unmount/status terminal

### SSE

server → client ทางเดียว เหมาะกับ progress/events ง่ายกว่า WebSocket แต่ต้องดู proxy
timeouts, reconnect และ event ID

### WebSocket

สองทางและ realtime มากกว่า แต่ connection/state/scale ซับซ้อน ใช้เมื่อ requirement ต้องการ

อย่าเลือก WebSocket เพียงเพราะดู realtime

## 12. Loading และ error boundaries

- `loading.tsx` ให้ route segment แสดง fallback ระหว่าง streaming
- `error.tsx` เป็น Client Component สำหรับ runtime error boundary และ retry UI
- `not-found.tsx` แยก resource ไม่มีจริง
- error message ต่อผู้ใช้ไม่ควรเปิด stack/internal response
- observability ต้องเชื่อม error กับ request/correlation ID

## 13. Form และ mutation states

ควรแยก:

- idle
- validating
- submitting
- accepted/processing
- success
- validation failure
- transient failure/retry

Disable ปุ่มช่วย UX แต่ idempotency ที่ backend ยังจำเป็นเพราะ double click/network retry

## 14. Code-reading lab

อ่าน `labs/nextjs-import-dashboard/` ตามลำดับ:

1. `app/imports/page.tsx` — Server Component และ server-only fetch
2. `app/imports/new/page.tsx` + `ImportForm.tsx` — Client file interaction
3. `app/api/imports/route.ts` — POST BFF, cookie และ stable idempotency key
4. `app/imports/[id]/page.tsx` — initial status
5. `app/api/imports/[id]/route.ts` — authenticated polling BFF
6. `app/imports/[id]/ImportProgress.tsx` — bounded polling/cleanup

อ่านแล้วตอบ:

- ทำไม list/detail fetch อยู่ server
- ทำไม form/progress ต้อง client
- secret อยู่ที่ไหน
- timeout/cancellation ขาดตรงไหน
- ไฟล์ใหญ่ควรเปลี่ยนเป็น presigned upload อย่างไร
- CQRS projection lag แสดงต่อ user อย่างไร

## 15. คำถามสัมภาษณ์

### ทำไมไม่ใส่ `use client` ที่ root layout

จะขยาย client boundary ทำให้ส่ง JS/dependencies มากขึ้นและเสียประโยชน์ Server Components
ควรวางเฉพาะ subtree ที่ต้อง interactive

### Server Component ปลอดภัยเสมอหรือไม่

ไม่ แม้ code ไม่ส่ง browser แต่ output/props อาจ leak secret ต้อง authorize ทุก request,
ไม่ serialize sensitive data และระวัง caching ข้าม user/tenant

### `no-store` แก้ stale data ทุกแบบหรือไม่

ไม่ แก้เฉพาะ Next fetch caching ที่เกี่ยวข้อง Backend cache/read replica/CQRS projection
ยัง stale ได้

### จะป้องกัน duplicate submit อย่างไร

UI disable button + request state เพื่อ UX และ backend idempotency key/unique constraint
เพื่อ correctness

### SSR page ช้าเพราะ Go API ช้า ทำอย่างไร

วัด trace, ตั้ง timeout, parallelize independent fetches, cache เฉพาะข้อมูลที่ยอม stale,
stream boundaries และแก้ backend/query bottleneck ไม่เพียงเพิ่ม spinner

### localStorage กับ HttpOnly cookie

localStorage ถูก JavaScript อ่านได้จึงเสี่ยง XSS HttpOnly cookie ลด token theft จาก JS แต่
cookie-based mutation ต้องจัดการ CSRF/SameSite เลือกจาก threat model

### Next.js ต่างจาก React SPA ที่เคยทำอย่างไร

React เป็น UI library ส่วน Next เพิ่ม routing, server rendering, Server Components,
server-side data access, Route Handlers, caching/revalidation และ deployment conventions
ประสบการณ์ component/state/caching จาก React ใช้ต่อได้ แต่ต้องเรียน server/client boundary
และ cache semantics เพิ่ม
