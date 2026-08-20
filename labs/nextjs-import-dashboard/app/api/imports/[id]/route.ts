import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function GET(_: Request, { params }: { params: Promise<{ id: string }> }) {
  const session = (await cookies()).get("session")?.value;
  if (!session) return NextResponse.json({ error: "unauthorized" }, { status: 401 });

  const { id } = await params;
  const response = await fetch(`${process.env.IMPORT_API_URL}/imports/${id}`, {
    headers: { Authorization: `Bearer ${session}` },
    cache: "no-store",
  });
  return NextResponse.json(await response.json(), { status: response.status });
}
