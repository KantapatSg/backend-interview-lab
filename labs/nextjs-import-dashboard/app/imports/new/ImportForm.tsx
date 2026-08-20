"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

export function ImportForm() {
  const router = useRouter();
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  // The key remains stable if the browser retries this logical submission.
  const [idempotencyKey] = useState(() => crypto.randomUUID());

  async function submit(formData: FormData) {
    setSubmitting(true);
    setError("");
    try {
      const response = await fetch("/api/imports", {
        method: "POST",
        headers: { "Idempotency-Key": idempotencyKey },
        body: formData,
      });
      if (!response.ok) throw new Error("Upload was not accepted");
      const job: { id: string } = await response.json();
      router.push(`/imports/${job.id}`);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unexpected error");
      setSubmitting(false);
    }
  }

  return (
    <form action={submit}>
      <input name="file" type="file" accept=".csv,.xlsx" required />
      <button disabled={submitting}>{submitting ? "Uploading..." : "Start import"}</button>
      {error && <p role="alert">{error}</p>}
    </form>
  );
}
