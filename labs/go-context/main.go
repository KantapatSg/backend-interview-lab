package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// fetch simulates a repository call. Cancellation is cooperative: the function
// must observe ctx.Done(); merely creating a context does not stop work by itself.
func fetch(ctx context.Context, latency time.Duration) (string, error) {
	timer := time.NewTimer(latency)
	defer timer.Stop()

	select {
	case <-timer.C:
		return "order-123", nil
	case <-ctx.Done():
		return "", fmt.Errorf("fetch order: %w", ctx.Err())
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel() // Releases the timer even when the operation finishes early.

	_, err := fetch(ctx, 200*time.Millisecond)
	fmt.Println("deadline exceeded:", errors.Is(err, context.DeadlineExceeded))
}
