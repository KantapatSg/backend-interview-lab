package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KantapatSg/backend-interview-lab/platform/internal/app"
	"github.com/KantapatSg/backend-interview-lab/platform/internal/domain"
)

type fakeRepository struct{}

func (fakeRepository) CreateJob(_ context.Context, job domain.Job) (domain.Job, error) {
	return job, nil
}
func (fakeRepository) GetJob(context.Context, string, string) (domain.Job, error) {
	return domain.Job{}, domain.ErrNotFound
}
func (fakeRepository) ProcessJobEvent(context.Context, app.Event) error { return nil }

func newTestHandler(readiness ...func(context.Context) error) http.Handler {
	service := app.NewJobService(fakeRepository{}, nil, func() time.Time { return time.Unix(0, 0) })
	return NewServer(service, readiness...)
}

func TestCreateRequiresIdempotencyKey(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"type":"IMPORT_JOB"}`))
	response := httptest.NewRecorder()

	newTestHandler().ServeHTTP(response, req)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestReadyzReportsDependencyFailure(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()

	newTestHandler(func(context.Context) error { return errors.New("postgres unavailable") }).ServeHTTP(response, req)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
