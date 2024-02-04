package storageproxy

import (
	"errors"
	"log/slog"

	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/prometheus/prometheus/storage"
)

// GraviolaStorage is a wrapper around a group of groups
type GraviolaStorage struct {
	logger    *slog.Logger
	rootGroup *remotestoragegroup.Group
}

func NewGraviolaStorage(logger *slog.Logger, groups []storage.Querier) *GraviolaStorage {
	return &GraviolaStorage{
		logger:    logger,
		rootGroup: remotestoragegroup.NewGroup(logger, "root", groups),
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
