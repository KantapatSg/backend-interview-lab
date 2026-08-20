package domain

import (
	"testing"
	"time"
)

func TestJobTransitionIncrementsVersion(t *testing.T) {
	now := time.Now()
	j, err := NewJob("j1", "t1", "IMPORT_JOB", "k1", nil, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := j.Transition(JobProcessing, now); err != nil {
		t.Fatal(err)
	}
	if j.Status != JobProcessing || j.Version != 2 {
		t.Fatalf("job = %#v", j)
	}
	if err := j.Transition(JobPending, now); err == nil {
		t.Fatal("expected invalid backwards transition")
	}
}
