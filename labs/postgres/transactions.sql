-- 1) Command + outbox: either both records commit or neither does.
BEGIN;

INSERT INTO orders (id, customer_id, idempotency_key, status)
VALUES (:order_id, :customer_id, :idempotency_key, 'PROCESSING');

INSERT INTO outbox_events (
    event_id, aggregate_id, aggregate_version, event_type, payload
) VALUES (
    :event_id,
    :order_id,
    1,
    'OrderCreated',
    jsonb_build_object('orderId', :order_id)
);

COMMIT;

-- 2) Optimistic concurrency: zero affected rows means the caller used stale state.
UPDATE orders
SET status = 'CONFIRMED', version = version + 1
WHERE id = :order_id AND version = :expected_version;

-- 3) Idempotent consumer: receipt and side effect share one transaction.
BEGIN;

-- The CTE returns a row only for a new event. A duplicate makes accepted empty,
-- so the UPDATE has no side effect.
WITH accepted AS (
    INSERT INTO processed_events (consumer_group, event_id)
    VALUES ('order-projector-v1', :event_id)
    ON CONFLICT DO NOTHING
    RETURNING 1
)
UPDATE orders
SET status = 'CONFIRMED', version = version + 1
WHERE id = :order_id AND EXISTS (SELECT 1 FROM accepted);

COMMIT;

-- 4) Multiple publishers can claim different rows without waiting on each other.
BEGIN;

SELECT sequence, event_id, payload
FROM outbox_events
WHERE published_at IS NULL AND available_at <= NOW()
ORDER BY sequence
FOR UPDATE SKIP LOCKED
LIMIT 100;

-- Publish outside a long-held transaction in a production design, or use a short
-- claim/lease state. Never hold row locks while an unbounded network call waits.
ROLLBACK;
