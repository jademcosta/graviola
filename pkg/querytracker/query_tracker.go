package querytracker

import (
	"context"
	"fmt"
	"math"
	"math/rand"
)

type GraviolaQueryTracker struct {
	concurrencyLimmiter  chan struct{}
	maxConcurrentQueries int
}

func NewGraviolaQueryTracker(maxConcurrentQueries int) *GraviolaQueryTracker {
	if maxConcurrentQueries < 1 {
		panic("maxConcurrentQueries < 1 is not allowed")
	}

	return &GraviolaQueryTracker{
		concurrencyLimmiter: make(chan struct{}, maxConcurrentQueries),
	}
}

// QueryTracker
// GetMaxConcurrent returns maximum number of concurrent queries that are allowed by this tracker.
func (tracker *GraviolaQueryTracker) GetMaxConcurrent() int {
	return tracker.maxConcurrentQueries
}

// QueryTracker
// Insert inserts query into query tracker. This call must block if maximum number of queries is already running.
// If Insert doesn't return error then returned integer value should be used in subsequent Delete call.
// Insert should return error if context is finished before query can proceed, and integer value returned in this case should be ignored by caller.
func (tracker *GraviolaQueryTracker) Insert(ctx context.Context, _ string) (int, error) {
	select {
	case tracker.concurrencyLimmiter <- struct{}{}:
		return rand.Intn(math.MaxInt), nil
	case <-ctx.Done():
		return 0, fmt.Errorf("when waiting for query concurrency slot: %w", ctx.Err())
	}
}

// QueryTracker
// Delete removes query from activity tracker. InsertIndex is value returned by Insert call.
func (tracker *GraviolaQueryTracker) Delete(_ int) {
	<-tracker.concurrencyLimmiter
}

// QueryTracker
func (tracker *GraviolaQueryTracker) Close() error {
	return nil
}
