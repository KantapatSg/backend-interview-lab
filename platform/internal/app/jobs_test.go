package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KantapatSg/backend-interview-lab/platform/internal/domain"
)

type fakeRepo struct{ jobs map[string]domain.Job }

func (f *fakeRepo) CreateJob(_ context.Context, j domain.Job) (domain.Job, error) {
	if f.jobs == nil {
		f.jobs = map[string]domain.Job{}
	}
	f.jobs[j.ID] = j
	return j, nil
}
func (f *fakeRepo) GetJob(_ context.Context, _, id string) (domain.Job, error) {
	j, ok := f.jobs[id]
	if !ok {
		return domain.Job{}, domain.ErrNotFound
	}
	return j, nil
}
func (f *fakeRepo) ProcessJobEvent(context.Context, Event) error { return nil }

type fakeCache struct {
	value      domain.Job
	gets, sets int
}

func (f *fakeCache) Get(_ context.Context, _ string, dst *domain.Job) error {
	f.gets++
	if f.value.ID == "" {
		return domain.ErrNotFound
	}
	*dst = f.value
	return nil
}
func (f *fakeCache) Set(_ context.Context, _ string, value domain.Job, _ time.Duration) error {
	f.sets++
	f.value = value
	return nil
}

func TestCreateAndGetUsesCacheButKeepsDBSourceOfTruth(t *testing.T) {
	repo, cache := &fakeRepo{}, &fakeCache{}
	now := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	svc := NewJobService(repo, cache, func() time.Time { return now })
	job, err := svc.Create(context.Background(), "tenant-a", "IMPORT_JOB", "key-1", []byte(`{"rows":1}`))
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.Get(context.Background(), "tenant-a", job.ID)
	if err != nil || got.ID != job.ID {
		t.Fatalf("Get() = %#v, %v", got, err)
	}
	if cache.sets != 1 {
		t.Fatalf("cache sets = %d, want 1", cache.sets)
	}
	_, _ = svc.Get(context.Background(), "tenant-a", job.ID)
	if cache.gets != 2 {
		t.Fatalf("cache gets = %d, want 2", cache.gets)
	}
}

func TestGetFallsBackWhenCacheMisses(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewJobService(repo, &fakeCache{}, time.Now)
	_, err := svc.Get(context.Background(), "tenant-a", "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}
}
