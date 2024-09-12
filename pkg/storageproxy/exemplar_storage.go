package storageproxy

import (
	"context"

	"github.com/prometheus/prometheus/storage"
)

// TODO: implement me
type GraviolaExemplarQueryable struct {
}

// ExemplarQueryable
// nolint: nilnil
func (exQuerier *GraviolaExemplarQueryable) ExemplarQuerier(_ context.Context) (storage.ExemplarQuerier, error) {
	return nil, nil
}
