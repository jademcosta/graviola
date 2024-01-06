package querytracker

import (
	"context"
	"fmt"
	"sync"
)

// FIXME: add tests
type GraviolaQueryTracker struct {
	mu                  sync.Mutex
	concurrencyLimmiter chan struct{}
	queries             []string
}

func NewGraviolaQueryTracker(maxConcurrentQueries int) *GraviolaQueryTracker {
	if maxConcurrentQueries < 1 {
		panic("maxConcurrentQueries < 1") //TODO: add a logger
	}

	return &GraviolaQueryTracker{
		concurrencyLimmiter: make(chan struct{}, maxConcurrentQueries),
		queries:             make([]string, maxConcurrentQueries),
	}
}

func (tracker *GraviolaQueryTracker) GetMaxConcurrent() int {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	return len(tracker.queries)
}

func (tracker *GraviolaQueryTracker) Insert(ctx context.Context, query string) (int, error) {
	select {
	case tracker.concurrencyLimmiter <- struct{}{}:
		tracker.mu.Lock()
		defer tracker.mu.Unlock()
		idx, err := tracker.findEmptyQuerySlot()
		if err != nil {
			//TODO: log msg
			fmt.Println("ERROR graviola: ", err)
			return 0, err
		}

		tracker.queries[idx] = query
		return idx, nil

	case <-ctx.Done():
		return 0, fmt.Errorf("when waiting for query concurrency slot: %w", context.DeadlineExceeded)

	}
}

func (tracker *GraviolaQueryTracker) Delete(insertIndex int) {
	tracker.mu.Lock() //XXX: defer é LIFO, o mais de baixo roda primeiro
	defer func() { <-tracker.concurrencyLimmiter }()
	defer tracker.mu.Unlock()

	tracker.queries[insertIndex] = ""
}

func (tracker *GraviolaQueryTracker) findEmptyQuerySlot() (int, error) {
	for idx, val := range tracker.queries {
		if val == "" {
			return idx, nil
		}
	}
	return 0, fmt.Errorf("no empty query slot on query tracker")
}
