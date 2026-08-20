package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFetchHonorsDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	_, err := fetch(ctx, time.Second)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("fetch() error = %v, want deadline exceeded", err)
	}
}

func TestFetchCompletesBeforeDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got, err := fetch(ctx, time.Millisecond)
	if err != nil || got != "order-123" {
		t.Fatalf("fetch() = %q, %v", got, err)
	}
}
