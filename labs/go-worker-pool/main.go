package main

import (
	"context"
	"fmt"
	"sync"
)

type result struct {
	input  int
	square int
}

// runPool bounds concurrency to workers. The jobs buffer absorbs only a small
// burst; once full, the producer blocks and backpressure reaches the caller.
func runPool(ctx context.Context, inputs []int, workers int) ([]result, error) {
	if workers < 1 {
		return nil, fmt.Errorf("workers must be positive")
	}
	jobs := make(chan int, workers)
	results := make(chan result, workers)

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case n, ok := <-jobs:
					if !ok {
						return
					}
					select {
					case results <- result{input: n, square: n * n}:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	// One coordinator owns channel closure, so workers never close a shared channel.
	go func() {
		defer close(jobs)
		for _, n := range inputs {
			select {
			case jobs <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	output := make([]result, 0, len(inputs))
	for item := range results {
		output = append(output, item)
	}
	return output, ctx.Err()
}

func main() {
	output, err := runPool(context.Background(), []int{1, 2, 3, 4}, 2)
	fmt.Println("results:", output, "error:", err)
}
