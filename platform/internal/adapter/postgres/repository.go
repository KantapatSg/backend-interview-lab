package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/KantapatSg/backend-interview-lab/platform/internal/app"
	"github.com/KantapatSg/backend-interview-lab/platform/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) CreateJob(ctx context.Context, job domain.Job) (domain.Job, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	defer tx.Rollback(ctx)
	// Job + outbox share a transaction. This is the key dual-write invariant:
	// a committed job always has a durable event for the publisher to retry.
	_, err = tx.Exec(ctx, `INSERT INTO jobs (id, tenant_id, job_type, status, payload, idempotency_key, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, job.ID, job.TenantID, job.Type, job.Status, job.Payload, job.IdempotencyKey, job.Version, job.CreatedAt, job.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Job{}, domain.ErrDuplicate
		}
		return domain.Job{}, err
	}
	payload, _ := json.Marshal(app.Event{ID: job.ID, Type: "JobCreated.v1", JobID: job.ID, TenantID: job.TenantID, AggregateVer: job.Version, OccurredAt: job.CreatedAt, Payload: job.Payload})
	_, err = tx.Exec(ctx, `INSERT INTO outbox_events (event_id, aggregate_id, aggregate_version, event_type, payload) VALUES (gen_random_uuid(),$1,$2,$3,$4)`, job.ID, job.Version, "JobCreated.v1", payload)
	if err != nil {
		return domain.Job{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Job{}, err
	}
	return job, nil
}

func (r *Repository) GetJob(ctx context.Context, tenantID, id string) (domain.Job, error) {
	var j domain.Job
	err := r.pool.QueryRow(ctx, `SELECT id, tenant_id, job_type, status, payload, idempotency_key, version, created_at, updated_at FROM jobs WHERE tenant_id=$1 AND id=$2`, tenantID, id).
		Scan(&j.ID, &j.TenantID, &j.Type, &j.Status, &j.Payload, &j.IdempotencyKey, &j.Version, &j.CreatedAt, &j.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Job{}, domain.ErrNotFound
	}
	return j, err
}

func (r *Repository) ProcessJobEvent(ctx context.Context, event app.Event) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Insert receipt and state update atomically. A redelivered Kafka record
	// therefore becomes a no-op before it can apply business side effects twice.
	result, err := tx.Exec(ctx, `INSERT INTO processed_events (consumer_group,event_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, "job-worker", event.ID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return tx.Commit(ctx)
	}
	_, err = tx.Exec(ctx, `UPDATE jobs SET status='PROCESSING', version=version+1, updated_at=NOW() WHERE id=$1 AND tenant_id=$2`, event.JobID, event.TenantID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) ClaimOutbox(ctx context.Context, limit int) ([]app.Event, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	// Claiming with SKIP LOCKED plus a short lease prevents two publisher
	// instances from repeatedly sending the same pending row at once. The
	// lease is not a delivery guarantee: if Kafka fails, the row becomes
	// available again and the outbox remains the durable retry source.
	rows, err := tx.Query(ctx, `
		WITH candidates AS (
			SELECT sequence
			FROM outbox_events
			WHERE published_at IS NULL AND available_at <= NOW()
			ORDER BY sequence
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		UPDATE outbox_events AS o
		SET available_at = NOW() + INTERVAL '30 seconds'
		FROM candidates
		WHERE o.sequence = candidates.sequence
		RETURNING o.event_id, o.event_type, o.aggregate_id, o.aggregate_version, o.payload`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []app.Event
	for rows.Next() {
		var e app.Event
		if err := rows.Scan(&e.ID, &e.Type, &e.JobID, &e.AggregateVer, &e.Payload); err != nil {
			return nil, err
		}
		e.TenantID = "default"
		eventPayload := struct {
			TenantID string `json:"tenant_id"`
		}{}
		_ = json.Unmarshal(e.Payload, &eventPayload)
		if eventPayload.TenantID != "" {
			e.TenantID = eventPayload.TenantID
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *Repository) MarkOutboxPublished(ctx context.Context, eventID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE outbox_events SET published_at=NOW() WHERE event_id=$1 AND published_at IS NULL`, eventID)
	return err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
