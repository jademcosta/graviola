package main

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	"github.com/jademcosta/graviola/test/e2e/fixtures"
	"github.com/prometheus/prometheus/prompb"
)

//To not have to search it every time, this is the Prometheus response format:
//
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

	valuesSet := make(map[float64]struct{})

	_, err = jsonparser.ArrayEach(
		responsePrometheus,
		func(value []byte, _ jsonparser.ValueType, _ int, err error) {
			if err != nil {
				panic(fmt.Errorf("when iterating on Prometheus response: %w", err))
			}

			val, _, _, err := jsonparser.Get(value, "[1]")
			if err != nil {
				panic(fmt.Errorf("when iterating on Prometheus response, on value extraction: %w", err))
			}

			floatValue, err := parseToFloat64(val)
			if err != nil {
				panic(fmt.Errorf("when iterating on Prometheus response, on value parsing: %w", err))
			}

			valuesSet[floatValue] = struct{}{}
		},
		"data", "result", "[0]", "values")

	if err != nil {
		return fmt.Errorf("when iterating on response: %w", err)
	}

	if len(valuesSet) != len(timeSeries.Samples) {
		return fmt.Errorf(
			"the count of items in returned response seems to be different than what is required. Expected: %#v\n. Response: %v",
			timeSeries.Samples, string(responsePrometheus))
	}

	for _, expectedSample := range timeSeries.Samples {
		_, found := valuesSet[expectedSample.Value]
		if !found {
			return fmt.Errorf("value %v was expected but not found on %v", expectedSample.Value, valuesSet)
		}
	}
	return nil
}

func checkItCanMixAllPrometheusData(graviolaURL string, query string) {
	responseBody, err := execRemoteQuery(
		graviolaURL,
		"/query_range",
		query,
		fixtures.ThirtyMinAgo.Add(-2*time.Minute), //TODO: use the fixture tmestamp to decide start and end
		fixtures.ThirtyMinAgo.Add(time.Minute),
	)
	if err != nil {
		panic(fmt.Errorf("graviola query failed: %w", err))
	}

	sampleValues := make([]float64, 0)
	_, err = jsonparser.ArrayEach(
		responseBody,
		func(value []byte, _ jsonparser.ValueType, _ int, err error) {
			if err != nil {
				panic(fmt.Errorf("when parsing single value entry: %w", err))
			}

			val, _, _, err := jsonparser.Get(value, "[1]")
			if err != nil {
				panic(fmt.Errorf("when iterating on Prometheus response, on value extraction: %w", err))
			}

			floatValue, err := parseToFloat64(val)
			if err != nil {
				panic(fmt.Errorf("when iterating on Prometheus response, on value parsing: %w", err))
			}

			sampleValues = append(sampleValues, floatValue)
		},
		"data", "result", "[0]", "values",
	)
	if err != nil {
		panic(fmt.Errorf("when iterating on response values: %w", err))
	}

	sampleValues = dedup(sampleValues)

	expectedLength := len(fixtures.CalculateSumQueryAnswer())
	if len(sampleValues) != expectedLength {
		panic(fmt.Errorf("the answer for sum should have %d entries, but has %d. Response was: %v. Expected values are: %v",
			expectedLength, len(sampleValues), sampleValues, fixtures.CalculateSumQueryAnswer()))
	}
}

func parseToFloat64(val []byte) (float64, error) {
	valAsInt, err := strconv.ParseInt(string(val), 10, 64)
	if err != nil {
		return strconv.ParseFloat(string(val), 64)
	}

	return float64(valAsInt), err
}

func dedup(input []float64) []float64 {
	lastInsert := float64(-1.0)
	output := make([]float64, 0)

	slices.Sort(input)

	for _, value := range input {
		if lastInsert-value != 0 {
			output = append(output, value)
			lastInsert = value
		}
	}
	return output
}
