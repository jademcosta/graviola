package domain_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/domain"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"github.com/stretchr/testify/assert"
)

func TestLabels(t *testing.T) {
	startingLabels := labels.FromStrings("label1", "value1", "label2", "value2")
	sut := &domain.GraviolaSeries{
		Datapoints: []model.SamplePair{{Timestamp: 123, Value: 4.56}},
		Lbs:        startingLabels,
	}

	labelsComparisonEqualResult := 0

	assert.Equal(t, labelsComparisonEqualResult, labels.Compare(startingLabels, sut.Labels()), "labels should contain the values passed on creation")
}

func TestIteratorNext(t *testing.T) {
	startingLabels := labels.FromStrings("label1", "value1", "label2", "value2")
	sut := &domain.GraviolaSeries{
		Datapoints: []model.SamplePair{{Timestamp: 123, Value: 4.56}, {Timestamp: 456, Value: 7.89}},
		Lbs:        startingLabels,
	}

	iter1 := sut.Iterator(nil)
	assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
	assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
	assert.Equal(t, chunkenc.ValNone, iter1.Next(), "should have reached the end")

	_ = sut.Iterator(iter1)
	assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
	assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
	assert.Equal(t, chunkenc.ValNone, iter1.Next(), "should have reached the end")
}

func TestIteratorIterating(t *testing.T) {
	startingLabels := labels.FromStrings("label1", "value1", "label2", "value2")
	sut := &domain.GraviolaSeries{
		Datapoints: []model.SamplePair{{Timestamp: 123, Value: 4.56}, {Timestamp: 456, Value: 7.89}, {Timestamp: 789, Value: 10.1112}, {Timestamp: 101112, Value: 13.1415}},
		Lbs:        startingLabels,
	}

	iter1 := sut.Iterator(nil)
	assert.Panics(t, func() { iter1.At() }, "should panic if try to execute At() without Next called first")
	assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
	ts, val := iter1.At()
	assert.Equal(t, int64(123), ts, "should answer the correct timestamp")
	assert.Equal(t, float64(4.56), val, "should answer the correct value")

	assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
	ts, val = iter1.At()
	assert.Equal(t, int64(456), ts, "should answer the correct timestamp")
	assert.Equal(t, float64(7.89), val, "should answer the correct value")

	assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
	ts, val = iter1.At()
	assert.Equal(t, int64(789), ts, "should answer the correct timestamp")
	assert.Equal(t, float64(10.1112), val, "should answer the correct value")

	assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
	ts, val = iter1.At()
	assert.Equal(t, int64(101112), ts, "should answer the correct timestamp")
	assert.Equal(t, float64(13.1415), val, "should answer the correct value")

	assert.Equal(t, chunkenc.ValNone, iter1.Next(), "should have reached the end")
	assert.Panics(t, func() { iter1.At() }, "should panic if try to execute At() when there's no next")
}

func TestIteratorReusage(t *testing.T) {
	startingLabels := labels.FromStrings("label1", "value1", "label2", "value2")
	sut := &domain.GraviolaSeries{
		Datapoints: []model.SamplePair{{Timestamp: 123, Value: 4.56}, {Timestamp: 456, Value: 7.89}},
		Lbs:        startingLabels,
	}

	iter1 := sut.Iterator(nil)
	assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
	assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
	assert.Equal(t, chunkenc.ValNone, iter1.Next(), "should have reached the end")

	_ = sut.Iterator(iter1)
	assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
	assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
	assert.Equal(t, chunkenc.ValNone, iter1.Next(), "should have reached the end")

	iter2 := sut.Iterator(iter1)
	assert.Equal(t, chunkenc.ValFloat, iter2.Next(), "should be able to advance the iterator")
	assert.Equal(t, chunkenc.ValFloat, iter2.Next(), "should be able to advance the iterator")
	assert.Equal(t, chunkenc.ValNone, iter2.Next(), "should have reached the end")

	_ = sut.Iterator(nil)
	assert.Equal(t, chunkenc.ValNone, iter1.Next(), "should still be at the end")
	// assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
	// assert.Equal(t, chunkenc.ValFloat, iter1.Next(), "should be able to advance the iterator")
}

func TestNonSupportedFunctions(t *testing.T) {
	startingLabels := labels.FromStrings("label1", "value1", "label2", "value2")
	sut := &domain.GraviolaSeries{
		Datapoints: []model.SamplePair{{Timestamp: 123, Value: 4.56}, {Timestamp: 456, Value: 7.89}},
		Lbs:        startingLabels,
	}

	iter1 := sut.Iterator(nil)
	assert.Panics(t, func() { iter1.AtFloatHistogram() }, "should panic if try to execute AtFloatHistogram()")
	assert.Panics(t, func() { iter1.AtHistogram() }, "should panic if try to execute AtHistogram()")
}

func TestErrIsAlwaysNil(t *testing.T) {
	startingLabels := labels.FromStrings("label1", "value1", "label2", "value2")
	sut := &domain.GraviolaSeries{
		Datapoints: []model.SamplePair{{Timestamp: 123, Value: 4.56}, {Timestamp: 456, Value: 7.89}},
		Lbs:        startingLabels,
	}

	iter1 := sut.Iterator(nil)
	assert.Nil(t, iter1.Err(), "should never return an error")
}

func TestAtT(t *testing.T) {
	startingLabels := labels.FromStrings("label1", "value1", "label2", "value2")
	sut := &domain.GraviolaSeries{
		Datapoints: []model.SamplePair{{Timestamp: 123, Value: 4.56}, {Timestamp: 456, Value: 7.89}},
		Lbs:        startingLabels,
	}

	iter1 := sut.Iterator(nil)
	iter1.Next()
	assert.Equal(t, int64(123), iter1.AtT(), "should return the timestamp part of current series")
	iter1.Next()
	assert.Equal(t, int64(456), iter1.AtT(), "should return the timestamp part of current series")
	iter1.Next()
	assert.Equal(t, int64(456), iter1.AtT(), "should keep returning the last timestamp of last series if the iterator reached the end")
	iter1.Next()
	assert.Equal(t, int64(456), iter1.AtT(), "should keep returning the last timestamp of last series if the iterator reached the end")
}
