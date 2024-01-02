package storageproxy

import (
	"context"

	"github.com/prometheus/prometheus/storage"
)

// TODO: implement me
type GraviolaExemplarQueryable struct {
}

// ExemplarQueryable
func (exQuerier *GraviolaExemplarQueryable) ExemplarQuerier(ctx context.Context) (storage.ExemplarQuerier, error) {
	return nil, nil
}
