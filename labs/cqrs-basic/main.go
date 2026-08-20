package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var errNotFound = errors.New("order not found in read model")

type order struct {
	ID     string
	Status string
}

type orderCreated struct {
	ID string
}

type store struct {
	mu     sync.RWMutex
	write  map[string]order // Normalized/source-of-truth side in this small lab.
	read   map[string]order // Query-optimized projection.
	events chan orderCreated
}

func newStore() *store {
	return &store{
		write:  make(map[string]order),
		read:   make(map[string]order),
		events: make(chan orderCreated, 8),
	}
}

// createOrder is a command: it validates intent and changes the write model.
// A real service writes state and an outbox row in one database transaction.
func (s *store) createOrder(id string) (order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.write[id]; exists {
		return order{}, fmt.Errorf("duplicate order %s", id)
	}
	created := order{ID: id, Status: "PROCESSING"}
	s.write[id] = created
	s.events <- orderCreated{ID: id}
	return created, nil // Return write-side acknowledgement; projection may lag.
}

// getOrder is a query and never mutates business state.
func (s *store) getOrder(id string) (order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.read[id]
	if !ok {
		return order{}, errNotFound
	}
	return value, nil
}

// runProjector deliberately delays projection to make eventual consistency visible.
func (s *store) runProjector(ctx context.Context, delay time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-s.events:
			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
				s.mu.Lock()
				s.read[event.ID] = s.write[event.ID]
				s.mu.Unlock()
			case <-ctx.Done():
				timer.Stop()
				return
			}
		}
	}
}

func main() {
	s := newStore()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.runProjector(ctx, 50*time.Millisecond)

	created, _ := s.createOrder("order-123")
	_, immediateErr := s.getOrder(created.ID)
	fmt.Println("immediate query not found:", errors.Is(immediateErr, errNotFound))

	time.Sleep(60 * time.Millisecond)
	projected, _ := s.getOrder(created.ID)
	fmt.Println("projected:", projected)
}
