package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/KantapatSg/backend-interview-lab/platform/internal/app"
	"github.com/KantapatSg/backend-interview-lab/platform/internal/domain"
)

type Server struct {
	jobs      *app.JobService
	readiness func(context.Context) error
}

// NewServer accepts an optional dependency probe so /readyz can describe
// whether this process may receive traffic, while /healthz remains a cheap
// process liveness check. Keeping the probe injectable makes the HTTP layer
// testable without starting PostgreSQL.
func NewServer(jobs *app.JobService, probes ...func(context.Context) error) http.Handler {
	s := &Server{jobs: jobs}
	if len(probes) > 0 {
		s.readiness = probes[0]
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /readyz", s.ready)
	mux.HandleFunc("POST /jobs", s.create)
	mux.HandleFunc("GET /jobs/{id}", s.get)
	return requestID(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if s.readiness != nil {
		if err := s.readiness(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

type createRequest struct {
	TenantID string          `json:"tenant_id"`
	Type     string          `json:"type"`
	Payload  json.RawMessage `json:"payload"`
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Idempotency-Key is required"})
		return
	}
	var req createRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.Type == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type and valid JSON are required"})
		return
	}
	job, err := s.jobs.Create(r.Context(), valueOrDefault(req.TenantID, "default"), req.Type, idempotencyKey, req.Payload)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, job)
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := s.jobs.Get(r.Context(), valueOrDefault(r.Header.Get("X-Tenant-ID"), "default"), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, domain.ErrNotFound) {
		status = http.StatusNotFound
	}
	if errors.Is(err, domain.ErrInvalidJobType) {
		status = http.StatusBadRequest
	}
	if errors.Is(err, domain.ErrDuplicate) {
		status = http.StatusConflict
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimSpace(r.Header.Get("X-Request-ID")) == "" {
			w.Header().Set("X-Request-ID", "generated")
		}
		next.ServeHTTP(w, r)
	})
}
func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
