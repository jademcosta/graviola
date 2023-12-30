package storageproxy_test

import (
	"log/slog"
	"testing"

	"github.com/jademcosta/graviola/internal/mocks"
	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/jademcosta/graviola/pkg/storageproxy"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/assert"
)

var logg *slog.Logger = graviolalog.NewLogger(config.LogConfig{Level: "error"})

func TestCloseIsCalledOnWrappedGroup(t *testing.T) {
	mock1 := &mocks.RemoteStorageMock{}
	mock2 := &mocks.RemoteStorageMock{}

	sut := storageproxy.NewGraviolaStorage(logg, []storage.Querier{mock1, mock2})

	sut.Close()

	assert.Equal(t, 1, mock1.CloseCalled, "should have called close on underlying queriers")
	assert.Equal(t, 1, mock2.CloseCalled, "should have called close on underlying queriers")
}
