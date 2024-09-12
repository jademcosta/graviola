package app

import (
	"net/http"
	"testing"
	"time"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const configOneGroupWithOneRemote = `
api:
  port: 8091
  timeout: 3m

query:
  max_samples: 1000
  lookback_delta: 5m
  max_concurrent_queries: 30

log:
  level: error

storages:
  merge_strategy:
    type: keep_biggest
  groups:
    - name: "the solo group"
      on_query_fail: fail_all
      remotes:
        - name: "the server 1"
          address: "http://localhost:9090"
`

func TestIntegrationAnswers500OnPanic(t *testing.T) {

	conf := config.GraviolaConfig{}
	err := yaml.Unmarshal([]byte(configOneGroupWithOneRemote), &conf)
	if err != nil {
		panic(err)
	}

	sut := NewApp(conf)

	sut.router.Get("/boom", func(_ http.ResponseWriter, _ *http.Request) {
		panic("panic!")
	})

	go func() {
		sut.Start()
	}()

	defer sut.Stop()

	time.Sleep(500 * time.Millisecond)

	resp, err := http.Get("http://localhost:8091/boom")
	require.NoError(t, err, "should not error")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "HTTP status should be 500")
}
