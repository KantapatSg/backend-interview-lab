"use client";

import { useEffect, useState } from "react";

export type ImportJob = {
  id: string;
  status: "PENDING" | "RUNNING" | "COMPLETED" | "FAILED";
  processedRows: number;
  totalRows: number;
};

export function ImportProgress({ initial }: { initial: ImportJob }) {
  const [job, setJob] = useState(initial);

  useEffect(() => {
    if (job.status === "COMPLETED" || job.status === "FAILED") return;

    const controller = new AbortController();
    let attempts = 0;
    const timer = window.setInterval(async () => {
      attempts += 1;
      const response = await fetch(`/api/imports/${job.id}`, { signal: controller.signal });
      if (response.ok) setJob(await response.json());
      if (attempts >= 60) window.clearInterval(timer); // Bound polling to five minutes.
    }, 5_000);

    return () => {
      controller.abort();
      window.clearInterval(timer);
    };
  }, [job.id, job.status]);

  return (
    <section>
      <h1>Import {job.id}</h1>
      <p>Status: {job.status}</p>
      <p>{job.processedRows} / {job.totalRows} rows</p>
    </section>
  );
}
