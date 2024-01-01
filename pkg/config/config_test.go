package config_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
)

var testConfig string = `
api:
  port: 8091
  timeout: 12m

log:
  level: "error"

storages:
  merge_strategy:
    type: "type 2"
  groups:
    - name: "group 1 name"
      on_query_fail: fail_all
      time_window:
        start: "now-6h"
        end: "now"
      remotes:
        - name: "my server 1"
          address: "https://localhost:9090"
          path_prefix: ""
          timeout: 35s
          time_window:
            start: "now-6h"
            end: "now"
    - name: "group 2 name"
      on_query_fail: partial_response
      time_window:
        start: "now-16h"
        end: "now-1h"
      remotes:
        - name: "my server 11"
          address: "https://localhost:9090"
          path_prefix: "/here"
          timeout: 35s
          time_window:
            start: "now-6h"
            end: "now"
        - name: "my server 12"
          address: "https://localhost:9092"
          path_prefix: "/hello/api/"
          timeout: 35s
          time_window:
            start: "now-6h"
            end: "now"
`

func TestParse(t *testing.T) {
	_, err := config.Parse([]byte("broken yaml"))
	assert.Error(t, err, "should result in error if config is not valid")

	result, err := config.Parse([]byte(testConfig))
	assert.NoError(t, err, "should result in NO error if config is valid")

	expected := config.GraviolaConfig{
		ApiConf: config.ApiConfig{
			Port:    8091,
			Timeout: "12m",
		},
		LogConf: config.LogConfig{
			Level: "error",
		},
		StoragesConf: config.StoragesConfig{
			MergeConf: config.MergeStrategyConfig{
				Type: "type 2",
			},
			Groups: []config.GroupsConfig{
				{
					Name:                "group 1 name",
					OnQueryFailStrategy: "fail_all",
					TimeWindow: config.TimeWindowConfig{
						Start: "now-6h",
						End:   "now",
					},
					Servers: []config.RemoteConfig{
						{
							Name:       "my server 1",
							Address:    "https://localhost:9090",
							PathPrefix: "",
						},
					},
				},
				{
					Name:                "group 2 name",
					OnQueryFailStrategy: "partial_response",
					TimeWindow: config.TimeWindowConfig{
						Start: "now-16h",
						End:   "now-1h",
					},
					Servers: []config.RemoteConfig{
						{
							Name:       "my server 11",
							Address:    "https://localhost:9090",
							PathPrefix: "/here",
						},
						{
							Name:       "my server 12",
							Address:    "https://localhost:9092",
							PathPrefix: "/hello/api/",
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, result, "should have parsed all fields")

	// assert.Equal(t, "error", result.LogConf.Level, "should have parsed log.level correctly")

	// assert.Equal(t, "12m", result.ApiConf.Timeout, "should have parsed api.timeout correctly")
	// assert.Equal(t, 8091, result.ApiConf.Port, "should have parsed api.port correctly")

	// assert.Len(t, result.StoragesConf.Groups, 2, "should have parsed all groups")

	// assert.Len(t, result.StoragesConf.Groups[0].Servers, 1, "should have parsed all remotes inside groups")
	// assert.Len(t, result.StoragesConf.Groups[1].Servers, 2, "should have parsed all remotes inside groups")

	// assert.Equal(t, "my server 1", result.StoragesConf.Groups[0].Servers[0].Name, "should have parsed emote names correctly")
}
