package storageproxy

import (
	"context"
	"log/slog"

	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/prometheus/prometheus/storage"
)

// GraviolaStorage is a wrapper around a group of groups
type GraviolaStorage struct {
	logg      *slog.Logger
	rootGroup *remotestoragegroup.Group
}

func NewGraviolaStorage(logg *slog.Logger, groups []storage.Querier) *GraviolaStorage {
	return &GraviolaStorage{
		logg:      logg,
		rootGroup: remotestoragegroup.NewGroup(logg, "root", groups),
	}
}

// Storage
func (gravStorage *GraviolaStorage) Close() error {
	return gravStorage.rootGroup.Close()
}

func (gravStorage *GraviolaStorage) StartTime() (int64, error) {
	panic("should not call StartTime")
	return 0, nil //TODO: implement me
}

// Appendable
func (gravStorage *GraviolaStorage) Appender(ctx context.Context) storage.Appender {
	panic("should not call Appender")
	return nil //TODO: implement me
}

// Queryable
func (gravStorage *GraviolaStorage) Querier(mint, maxt int64) (storage.Querier, error) {
	return gravStorage.rootGroup, nil
}

// ChunkQueryable
func (gravStorage *GraviolaStorage) ChunkQuerier(mint, maxt int64) (storage.ChunkQuerier, error) {
	panic("should not call ChunkQuerier")
	return nil, nil //TODO: implement me
}
