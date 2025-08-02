package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/graviolalog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const configOneGroupWithOneRemote = `
api:
  port: 8091

log:
  level: debug
`

type dummyRegisterer struct{}

func (d *dummyRegisterer) Register(_ *route.Router) {}

func TestIntegrationAnswers500OnPanic(t *testing.T) {

	conf := config.GraviolaConfig{}
	err := yaml.Unmarshal([]byte(configOneGroupWithOneRemote), &conf)
	if err != nil {
		panic(err)
	}

	sut := NewGraviolaAPI(
		conf.APIConf, graviolalog.NewLogger(conf.LogConf), prometheus.NewRegistry(), &dummyRegisterer{})

	sut.router.Get("/boom", func(_ http.ResponseWriter, _ *http.Request) {
		panic("panic boooooooommmmm!")
	})

	go func() {
		err := sut.Start()
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("should not error on api start: %v", err))
		}
	}()

	defer sut.Stop()

	time.Sleep(500 * time.Millisecond)

	resp, err := http.Get("http://localhost:8091/boom")
	require.NoError(t, err, "should not error")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "HTTP status should be 500")
}
