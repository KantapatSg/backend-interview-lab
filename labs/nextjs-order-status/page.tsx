import { OrderStatus } from "./OrderStatus";

type Order = { id: string; status: "PROCESSING" | "CONFIRMED" | "CANCELLED" };

// Server Component: backend URL/token remains on the server side.
export default async function OrderPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const response = await fetch(`${process.env.ORDER_API_URL}/orders/${id}`, {
    cache: "no-store", // Status is user-specific and changes during the workflow.
  });

  // A real app should distinguish forbidden, unavailable and projection-not-ready.
  const initial: Order = response.ok
    ? await response.json()
    : { id, status: "PROCESSING" };

  return <OrderStatus initial={initial} />;
}
