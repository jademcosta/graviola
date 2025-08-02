package storageproxy_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup"
	"github.com/jademcosta/graviola/pkg/storageproxy"
	"github.com/prometheus/prometheus/storage"
)

func TestGraviolaStorageComplyWithStorageSampleAndChunkQueryable(_ *testing.T) {
	logger := graviolalog.NewNoopLogger()
	groups := []storage.Querier{}
	mergeStrategy := remotestoragegroup.MergeStrategyFactory(config.MergeStrategyAlwaysMerge)

	dummyFunc := func(_ storage.SampleAndChunkQueryable) {}

	sut := storageproxy.NewGraviolaStorage(logger, groups, mergeStrategy)
	dummyFunc(sut)
}
