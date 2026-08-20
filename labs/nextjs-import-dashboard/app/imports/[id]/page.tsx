import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { ImportProgress, type ImportJob } from "./ImportProgress";

export default async function ImportDetail({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const session = (await cookies()).get("session")?.value;
  if (!session) redirect("/login");
  const response = await fetch(`${process.env.IMPORT_API_URL}/imports/${id}`, {
    headers: { Authorization: `Bearer ${session}` },
    cache: "no-store",
  });
  if (!response.ok) throw new Error("Unable to load import job");

  return <ImportProgress initial={(await response.json()) as ImportJob} />;
}
