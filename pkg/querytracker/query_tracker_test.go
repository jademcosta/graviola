package querytracker_test

import (
	"context"
	"testing"
	"time"

	"github.com/jademcosta/graviola/pkg/querytracker"
	"github.com/stretchr/testify/assert"
)

func TestInsertBlocksIfMaxConcurrencyIsReached(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	signalChan := make(chan struct{}, 5)

	sut := querytracker.NewGraviolaQueryTracker(2)

	sut.Insert(ctx, "")
	sut.Insert(ctx, "")

	go func() {
		sut.Insert(ctx, "")
		signalChan <- struct{}{}
	}()

	time.Sleep(100 * time.Millisecond)

	select {
	case <-signalChan:
		assert.Fail(t, "the third query should have been blocked")
	default:
	}

	sut.Delete(0)
	time.Sleep(100 * time.Millisecond)

	select {
	case <-signalChan:
	default:
		assert.Fail(t, "the deletion of a query should allow the blocked insert to proceed")
	}
}

func TestABlockedInsertIsReleasedIfContextIsDone(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())

	signalChan := make(chan error, 5)

	sut := querytracker.NewGraviolaQueryTracker(2)

	sut.Insert(ctx, "")
	sut.Insert(ctx, "")

	go func() {
		_, err := sut.Insert(ctx, "")
		signalChan <- err
	}()

	time.Sleep(100 * time.Millisecond)

	select {
	case <-signalChan:
		assert.Fail(t, "the third query should have been blocked")
	default:
	}

	cancelFn()
	time.Sleep(100 * time.Millisecond)

	select {
	case err := <-signalChan:
		assert.Error(t, err, "the failed insert should return an error")
	default:
		assert.Fail(t, "the deletion of a query should allow the blocked insert to proceed")
	}
}

func TestNewPanicsIfConcurrentQueriesIsLessThanOne(t *testing.T) {
	assert.Panics(t, func() { querytracker.NewGraviolaQueryTracker(0) },
		"should panic if concurrent queries is < 1")
}
