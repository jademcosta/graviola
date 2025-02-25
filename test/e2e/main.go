// nolint:forbidgo
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jademcosta/graviola/test/e2e/fixtures"
)

const apiVersionPath = "/api/v1"
const prometheus1URLWithPath = "http://localhost:9091" + apiVersionPath
const graviolaURLWithPath = "http://localhost:9197" + apiVersionPath

var httpCli = &http.Client{
	Timeout: 2 * time.Second,
}

func main() {
	err := sendMetricsTo(fixtures.SingleCounterMetric, prometheus1URLWithPath)
	if err != nil {
		panic(err)
	}

	for idx, ts := range fixtures.SingleCounterMetric.Timeseries {
		// err = ensureMetricsExist() //FIXME: I need to make sure metrics were ingested.

		err = checkMetricsMatch(ts, graviolaURLWithPath, prometheus1URLWithPath)
		if err != nil {
			panic(fmt.Errorf("comparsion error on index %d: %w", idx, err))
		}
	}
}
