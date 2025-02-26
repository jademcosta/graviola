// nolint:forbidgo
package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jademcosta/graviola/test/e2e/fixtures"
)

const apiVersionPath = "/api/v1"
const prometheus1URLWithPath = "http://localhost:9091" + apiVersionPath
const prometheus2URLWithPath = "http://localhost:9092" + apiVersionPath
const graviolaURLWithPath = "http://localhost:9197" + apiVersionPath

var httpCli = &http.Client{
	Timeout: 2 * time.Second,
}

func main() {
	sendMetricsToRemotes()

	checkMetricsWereReallySent()

	err := checkMetricsMatch(fixtures.SingleCounterMetric.Timeseries[0], graviolaURLWithPath, prometheus1URLWithPath)
	if err != nil {
		panic(fmt.Errorf("comparsion error on first sample: %w", err))
	}

	err = checkMetricsMatch(fixtures.SingleCounterMetric2.Timeseries[0], graviolaURLWithPath, prometheus1URLWithPath)
	if err != nil {
		panic(fmt.Errorf("comparsion error on second sample : %w", err))
	}

	os.Exit(0)
}

func sendMetricsToRemotes() {
	// Prometheus 1
	err := sendMetricsToPrometheus(fixtures.SingleCounterMetric, prometheus1URLWithPath)
	if err != nil {
		panic(err)
	}

	err = sendMetricsToPrometheus(fixtures.SingleCounterMetric2, prometheus1URLWithPath)
	if err != nil {
		panic(err)
	}

	//Prometheus 2
	err = sendMetricsToPrometheus(fixtures.SingleCounterMetric3, prometheus2URLWithPath)
	if err != nil {
		panic(err)
	}
}

func checkMetricsWereReallySent() {
	//Prometheus 1
	err := checkMetricsExist(fixtures.SingleCounterMetric.Timeseries[0], prometheus1URLWithPath)
	if err != nil {
		panic(fmt.Errorf("metric 1 existence not confirmed on Prometheus 1: %w", err))
	}

	err = checkMetricsExist(fixtures.SingleCounterMetric2.Timeseries[0], prometheus1URLWithPath)
	if err != nil {
		panic(fmt.Errorf("metric 2 existence not confirmed on Prometheus 1: %w", err))
	}

	//Prometheus 2
	err = checkMetricsExist(fixtures.SingleCounterMetric3.Timeseries[0], prometheus2URLWithPath)
	if err != nil {
		panic(fmt.Errorf("metric 3 existence not confirmed on Prometheus 2: %w", err))
	}
}
