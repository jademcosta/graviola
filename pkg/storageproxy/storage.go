package storageproxy

import (
	"errors"
	"log/slog"

	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup/queryfailurestrategy"
	"github.com/prometheus/prometheus/storage"
)

// GraviolaStorage is a wrapper around a group of groups
type GraviolaStorage struct {
	logger    *slog.Logger
	rootGroup *remotestoragegroup.Group
}

func NewGraviolaStorage(logger *slog.Logger, groups []storage.Querier) *GraviolaStorage {
	return &GraviolaStorage{
		logger: logger,
		//TODO: should this be the default? Maybe allow to configure it
		rootGroup: remotestoragegroup.NewGroup(logger, "root", groups, &queryfailurestrategy.FailAllStrategy{}),
	}
}

// Queryable
func (gravStorage *GraviolaStorage) Querier(mint, maxt int64) (storage.Querier, error) {
	return gravStorage.rootGroup, nil
}

// ChunkQueryable
func (gravStorage *GraviolaStorage) ChunkQuerier(mint, maxt int64) (storage.ChunkQuerier, error) {
	err := errors.New("should not call ChunkQuerier")
	gravStorage.logger.Error("ChunkQuerier called on GraviolaStorage")
	return nil, err
}
