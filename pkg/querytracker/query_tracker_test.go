package querytracker_test

import (
	"context"
	"testing"
	"time"

	"github.com/jademcosta/graviola/pkg/querytracker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertBlocksIfMaxConcurrencyIsReached(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	signalChan := make(chan struct{}, 5)

	sut := querytracker.NewGraviolaQueryTracker(2)

	_, err := sut.Insert(ctx, "")
	require.NoError(t, err, "should not error")
	_, err = sut.Insert(ctx, "")
	require.NoError(t, err, "should not error")

	go func() {
		_, err := sut.Insert(ctx, "")
		assert.NoError(t, err, "should not error")
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

	_, err := sut.Insert(ctx, "")
	require.NoError(t, err, "should not error")
	_, err = sut.Insert(ctx, "")
	require.NoError(t, err, "should not error")

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
		require.Error(t, err, "the failed insert should return an error")
	default:
		assert.Fail(t, "the deletion of a query should allow the blocked insert to proceed")
	}
}

func TestNewPanicsIfConcurrentQueriesIsLessThanOne(t *testing.T) {
	assert.Panics(t, func() { querytracker.NewGraviolaQueryTracker(0) },
		"should panic if concurrent queries is < 1")
}
