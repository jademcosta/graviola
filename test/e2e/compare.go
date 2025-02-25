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

	_, err := execRemoteQuery(
		prometheusURL,
		"/query_range",
		query,
		fixtures.ThirtyMinAgo.Add(-2*time.Minute), //TODO: use the fixture tmestamp to decide start and end
		fixtures.ThirtyMinAgo.Add(time.Minute),
	)
	if err != nil {
		return fmt.Errorf("prometheus query failed: %w", err)
	}

	// responsePrometheus:
	// {
	// 	"status":"success",
	// 	"data":{
	// 		"resultType":"matrix",
	// 		"result":[
	// 			{
	// 				"metric":{"__name__":"http_requests_total","job":"sys","region":"us-east-1","system":"ab"},
	// 				"values":[
	// 					[1740516680,"1"],
	// 					[1740516695,"2"],
	// 					[1740516710,"3"],
	// 					[1740516725,"4"],
	// 					[1740516740,"5"],
	// 					[1740516755,"10"],
	// 					[1740516770,"10"],
	// 					[1740516785,"10"],
	// 					[1740516800,"10"],
	// 					[1740516815,"10"]
	// 				]
	// 			}
	// 		]
	// 	}
	// }
	return nil
}
