package domain

import (
	"errors"
	"time"
)

// JobStatus is deliberately small for the MVP. More states can be added by a
// migration, but callers should not infer business completion from HTTP 202.
type JobStatus string

const (
	JobPending    JobStatus = "PENDING"
	JobProcessing JobStatus = "PROCESSING"
	JobCompleted  JobStatus = "COMPLETED"
	JobFailed     JobStatus = "FAILED"
)

var (
	ErrNotFound       = errors.New("job not found")
	ErrDuplicate      = errors.New("duplicate idempotency key")
	ErrInvalidJobType = errors.New("invalid job type")
)

type Job struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	Type           string    `json:"type"`
	Status         JobStatus `json:"status"`
	Payload        []byte    `json:"payload"`
	IdempotencyKey string    `json:"-"`
	Version        int64     `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func NewJob(id, tenantID, jobType, idempotencyKey string, payload []byte, now time.Time) (Job, error) {
	if jobType == "" {
		return Job{}, ErrInvalidJobType
	}
	return Job{ID: id, TenantID: tenantID, Type: jobType, Status: JobPending,
		Payload: payload, IdempotencyKey: idempotencyKey, Version: 1, CreatedAt: now, UpdatedAt: now}, nil
}

// Transition owns the invariant so HTTP handlers and Kafka consumers cannot
// independently invent valid state transitions.
func (j *Job) Transition(next JobStatus, now time.Time) error {
	valid := (j.Status == JobPending && next == JobProcessing) ||
		(j.Status == JobProcessing && (next == JobCompleted || next == JobFailed))
	if !valid {
		return errors.New("invalid job transition")
	}
	j.Status, j.Version, j.UpdatedAt = next, j.Version+1, now
	return nil
}
