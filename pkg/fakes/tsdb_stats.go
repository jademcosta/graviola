package fakes

import (
	"context"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/tsdb"
)

// type TSDBAdminStats interface {
// 	CleanTombstones() error
// 	Delete(mint, maxt int64, ms ...*labels.Matcher) error
// 	Snapshot(dir string, withHead bool) error
// 	Stats(statsByLabelName string, limit int) (*tsdb.Stats, error)
// 	WALReplayStatus() (tsdb.WALReplayStatus, error)
// }

// TODO: use logger here
type FakeTSDBAdminStats struct{}

func (tsdbadminsts *FakeTSDBAdminStats) CleanTombstones() error {
	panic("should not be called")
}

func (tsdbadminsts *FakeTSDBAdminStats) Delete(ctx context.Context, mint, maxt int64, ms ...*labels.Matcher) error {
	panic("should not be called")
}

func (tsdbadminsts *FakeTSDBAdminStats) Snapshot(dir string, withHead bool) error {
	panic("should not be called")
}

func (tsdbadminsts *FakeTSDBAdminStats) Stats(statsByLabelName string, limit int) (*tsdb.Stats, error) {
	panic("should not be called")
}

func (tsdbadminsts *FakeTSDBAdminStats) WALReplayStatus() (tsdb.WALReplayStatus, error) {
	panic("should not be called")
}
