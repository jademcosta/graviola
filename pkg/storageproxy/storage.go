package storageproxy

import (
	"errors"
	"log/slog"

	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup/queryfailurestrategy"
	"github.com/prometheus/prometheus/storage"
)

// GraviolaStorage is a wrapper around a list of groups. It implements the same interface of a
// Prometheus "Queryable". So, it acts like a "storage" of data
type GraviolaStorage struct {
	logger    *slog.Logger
	rootGroup *remotestoragegroup.Group
}

func NewGraviolaStorage(
	logger *slog.Logger, groups []storage.Querier, mergeStrategy remotestoragegroup.MergeStrategy,
) *GraviolaStorage {
	return &GraviolaStorage{
		logger: logger,
		//TODO: should this fail strategy be the default? Allow to configure it
		rootGroup: remotestoragegroup.NewGroup(logger, "root", groups,
			&queryfailurestrategy.FailAllStrategy{},
			mergeStrategy),
	}
}

// Queryable
// mint, maxt int64
func (gravStorage *GraviolaStorage) Querier(_, _ int64) (storage.Querier, error) {
	return gravStorage.rootGroup, nil
}

// ChunkQueryable
// mint, maxt int64
func (gravStorage *GraviolaStorage) ChunkQuerier(_, _ int64) (storage.ChunkQuerier, error) {
	err := errors.New("should not call ChunkQuerier")
	gravStorage.logger.Error("ChunkQuerier called on GraviolaStorage")
	return nil, err
}
