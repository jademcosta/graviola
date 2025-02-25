package main

import (
	"fmt"
	"slices"
	"time"

	"github.com/jademcosta/graviola/test/e2e/fixtures"
	"github.com/prometheus/prometheus/prompb"
)

func checkMetricsMatch(timeSeries prompb.TimeSeries, graviolaURL string, prometheusURL string) error {
	query := labelsToQuery(timeSeries.Labels)

	responsePrometheus, err := execRemoteQuery(
		prometheusURL,
		"/query_range",
		query,
		fixtures.ThirtyMinAgo.Add(-2*time.Minute), //TODO: use the fixture tmestamp to decide start and end
		fixtures.ThirtyMinAgo.Add(time.Minute),
	)
	if err != nil {
		return fmt.Errorf("prometheus query failed: %w", err)
	}

	responseGraviola, err := execRemoteQuery(
		graviolaURL,
		"/query_range",
		query,
		fixtures.ThirtyMinAgo.Add(-2*time.Minute), //TODO: use the fixture tmestamp to decide start and end
		fixtures.ThirtyMinAgo.Add(time.Minute),
	)
	if err != nil {
		return fmt.Errorf("graviola query failed: %w", err)
	}

	if slices.Compare(responsePrometheus, responseGraviola) != 0 {
		return fmt.Errorf("responses don't match:\n prometheus:\n%s\ngraviola:\n%s",
			responsePrometheus, responseGraviola)
	}

	return nil
}

func checkMetricsExist(timeSeries prompb.TimeSeries, prometheusURL string) error {
	query := labelsToQuery(timeSeries.Labels)

	responsePrometheus, err := execRemoteQuery(
		prometheusURL,
		"/query_range",
		query,
		fixtures.ThirtyMinAgo.Add(-2*time.Minute), //TODO: use the fixture tmestamp to decide start and end
		fixtures.ThirtyMinAgo.Add(time.Minute),
	)
	if err != nil {
		return fmt.Errorf("prometheus query failed: %w", err)
	}

	fmt.Printf("responsePrometheus: %s", string(responsePrometheus))

	return nil
}
