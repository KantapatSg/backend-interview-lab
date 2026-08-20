package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestProjectionIsEventuallyVisible(t *testing.T) {
	s := newStore()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.runProjector(ctx, time.Millisecond)

	created, err := s.createOrder("order-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.getOrder(created.ID); !errors.Is(err, errNotFound) {
		t.Fatalf("immediate query error = %v, want not found", err)
	}

	deadline := time.Now().Add(100 * time.Millisecond)
	for time.Now().Before(deadline) {
		if got, err := s.getOrder(created.ID); err == nil && got == created {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("projection did not become visible")
}
