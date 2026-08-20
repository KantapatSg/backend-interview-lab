import Link from "next/link";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

type ImportJob = { id: string; fileName: string; status: string; processedRows: number };

export default async function ImportsPage() {
  // Server Component keeps the session token out of the browser bundle.
  const session = (await cookies()).get("session")?.value;
  if (!session) redirect("/login");
  const response = await fetch(`${process.env.IMPORT_API_URL}/imports`, {
    headers: { Authorization: `Bearer ${session}` },
    cache: "no-store", // Job progress changes frequently and is user-specific.
  });
  if (!response.ok) throw new Error("Unable to load import jobs");

  const jobs: ImportJob[] = await response.json();
  return (
    <main>
      <Link href="/imports/new">New import</Link>
      {jobs.map((job) => (
        <p key={job.id}>
          <Link href={`/imports/${job.id}`}>{job.fileName}</Link>: {job.status}
        </p>
      ))}
    </main>
  );
}
