import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

export async function POST(request: NextRequest) {
  const session = (await cookies()).get("session")?.value;
  if (!session) return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  const idempotencyKey = request.headers.get("Idempotency-Key");
  if (!idempotencyKey) {
    return NextResponse.json({ error: "missing idempotency key" }, { status: 400 });
  }

  // This compact example proxies multipart data. For large files prefer a presigned
  // direct-to-object-storage upload so the Next server does not buffer the payload.
  const body = await request.formData();
  const response = await fetch(`${process.env.IMPORT_API_URL}/imports`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${session}`,
      "Idempotency-Key": idempotencyKey,
    },
    body,
  });

  const payload = await response.json();
  return NextResponse.json(payload, { status: response.status });
}
