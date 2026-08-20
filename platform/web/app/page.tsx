"use client";

import { FormEvent, useEffect, useState } from "react";

type Job = { id: string; type: string; status: string; created_at: string };

const apiBase = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export default function Dashboard() {
  const [job, setJob] = useState<Job | null>(null);
  const [jobType, setJobType] = useState("IMPORT_JOB");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError("");
    const idempotencyKey = crypto.randomUUID();
    const response = await fetch(`${apiBase}/jobs`, {
      method: "POST",
      headers: { "Content-Type": "application/json", "Idempotency-Key": idempotencyKey },
      body: JSON.stringify({ tenant_id: "demo", type: jobType, payload: { source: "dashboard" } }),
    });
    if (!response.ok) {
      setError(await response.text());
      setLoading(false);
      return;
    }
    const created = (await response.json()) as Job;
    setJob(created);
    setLoading(false);
  }

  useEffect(() => {
    if (!job || ["COMPLETED", "FAILED"].includes(job.status)) return;
    // Polling is bounded by the terminal state; production can replace this
    // with SSE without changing the command/query boundary.
    const timer = window.setTimeout(async () => {
      const response = await fetch(`${apiBase}/jobs/${job.id}`, { cache: "no-store" });
      if (response.ok) setJob((await response.json()) as Job);
    }, 1000);
    return () => window.clearTimeout(timer);
  }, [job]);

  return (
    <main className="shell">
      <p className="eyebrow">BACKEND INTERVIEW LAB</p>
      <h1>Data workflow dashboard</h1>
      <p className="lede">Create a job, observe eventual consistency, and trace the Go → PostgreSQL → Kafka flow.</p>
      <form onSubmit={submit} className="card">
        <label htmlFor="job-type">Job type</label>
        <select id="job-type" value={jobType} onChange={(event) => setJobType(event.target.value)}>
          <option>IMPORT_JOB</option>
          <option>SALES_ORDER</option>
        </select>
        <button disabled={loading} type="submit">{loading ? "Submitting…" : "Create job"}</button>
        {error && <p className="error">{error}</p>}
      </form>
      {job && <section className="card"><h2>Job status</h2><p className="status">{job.status}</p><code>{job.id}</code></section>}
    </main>
  );
}
