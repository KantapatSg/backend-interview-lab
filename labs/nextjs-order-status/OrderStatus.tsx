"use client";

import { useEffect, useState } from "react";

type Order = { id: string; status: "PROCESSING" | "CONFIRMED" | "CANCELLED" };

export function OrderStatus({ initial }: { initial: Order }) {
  const [order, setOrder] = useState(initial);

  useEffect(() => {
    if (order.status !== "PROCESSING") return;

    // Polling is intentionally bounded; production may use backoff, SSE or WebSocket.
    let attempts = 0;
    const timer = window.setInterval(async () => {
      attempts += 1;
      const response = await fetch(`/api/orders/${order.id}`);
      if (response.ok) setOrder(await response.json());
      if (attempts >= 10) window.clearInterval(timer);
    }, 1_000);

    return () => window.clearInterval(timer);
  }, [order.id, order.status]);

  return <p>Order {order.id}: {order.status}</p>;
}
