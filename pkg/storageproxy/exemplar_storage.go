package storageproxy

import (
	"context"

	"github.com/prometheus/prometheus/storage"
)

type GraviolaExemplarQueryable struct {
}

// ExemplarQueryable
func (exQuerier *GraviolaExemplarQueryable) ExemplarQuerier(ctx context.Context) (storage.ExemplarQuerier, error) {
	return nil, nil
}
