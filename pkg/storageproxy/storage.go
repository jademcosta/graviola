package storageproxy

import (
	"context"

	"github.com/jademcosta/graviola/pkg/graviolalog"

	"github.com/prometheus/prometheus/storage"
)

// TODO: Create constructor
type GraviolaStorage struct {
	Logg *graviolalog.Logger
}

// Storage
func (sp *GraviolaStorage) Close() error {
	sp.Logg.Info("GraviolaStorage: Close")
	return nil
}

func (sp *GraviolaStorage) StartTime() (int64, error) {
	sp.Logg.Info("GraviolaStorage: StartTime")
	return 0, nil
}

// Appendable
func (sp *GraviolaStorage) Appender(ctx context.Context) storage.Appender {
	sp.Logg.Info("GraviolaStorage: Appender")
	return nil
}

// Queryable
func (sp *GraviolaStorage) Querier(mint, maxt int64) (storage.Querier, error) {
	sp.Logg.Info("GraviolaStorage: Querier")
	return NewGraviolaQuerier(sp.Logg), nil
}

// ChunkQueryable
func (sp *GraviolaStorage) ChunkQuerier(mint, maxt int64) (storage.ChunkQuerier, error) {
	sp.Logg.Info("GraviolaStorage: ChunkQuerier")
	return nil, nil
}
