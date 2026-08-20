package app

import (
	"context"
	"errors"
	"time"

	"github.com/KantapatSg/backend-interview-lab/platform/internal/domain"
	"github.com/google/uuid"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job domain.Job) (domain.Job, error)
	GetJob(ctx context.Context, tenantID, id string) (domain.Job, error)
	ProcessJobEvent(ctx context.Context, event Event) error
}

type JobCache interface {
	Get(ctx context.Context, key string, dst *domain.Job) error
	Set(ctx context.Context, key string, value domain.Job, ttl time.Duration) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}

type Event struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	JobID        string    `json:"job_id"`
	TenantID     string    `json:"tenant_id"`
	AggregateVer int64     `json:"aggregate_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	Payload      []byte    `json:"payload"`
}

type JobService struct {
	repo  JobRepository
	cache JobCache
	now   func() time.Time
}

func NewJobService(repo JobRepository, cache JobCache, now func() time.Time) *JobService {
	if now == nil {
		now = time.Now
	}
	return &JobService{repo: repo, cache: cache, now: now}
}

func (s *JobService) Create(ctx context.Context, tenantID, jobType, idempotencyKey string, payload []byte) (domain.Job, error) {
	job, err := domain.NewJob(uuid.NewString(), tenantID, jobType, idempotencyKey, payload, s.now())
	if err != nil {
		return domain.Job{}, err
	}
	// The repository writes the job and its outbox event in one transaction.
	return s.repo.CreateJob(ctx, job)
}

func (s *JobService) Get(ctx context.Context, tenantID, id string) (domain.Job, error) {
	key := "job-view:v1:" + tenantID + ":" + id
	if s.cache != nil {
		var cached domain.Job
		if err := s.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		} else if !errors.Is(err, domain.ErrNotFound) {
			// Cache is an accelerator. A cache outage must not hide durable state.
		}
	}
	job, err := s.repo.GetJob(ctx, tenantID, id)
	if err != nil {
		return domain.Job{}, err
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, key, job, 30*time.Second)
	}
	return job, nil
}
