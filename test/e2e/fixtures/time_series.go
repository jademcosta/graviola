package fixtures

import (
	"time"

	"github.com/prometheus/prometheus/prompb"
)

// This is the rate at which metrics will be generated, and should be used as step when querying
const MetricStep = 15

var CurrentTimestamp = time.Now()
var ThirtyMinAgo = CurrentTimestamp.Add(-30 * time.Minute)

// A single counter metric with multiple samples
var SingleCounterMetric = &prompb.WriteRequest{
	Timeseries: []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "http_requests_total"},
				{Name: "region", Value: "us-east-1"},
				{Name: "system", Value: "ab"},
				{Name: "job", Value: "sys"},
			},
			Samples: []prompb.Sample{
				{Value: 1, Timestamp: ThirtyMinAgo.Add(-MetricStep * 6 * time.Second).UnixMilli()},
				{Value: 2, Timestamp: ThirtyMinAgo.Add(-MetricStep * 5 * time.Second).UnixMilli()},
				{Value: 3, Timestamp: ThirtyMinAgo.Add(-MetricStep * 4 * time.Second).UnixMilli()},
				{Value: 4, Timestamp: ThirtyMinAgo.Add(-MetricStep * 3 * time.Second).UnixMilli()},
				{Value: 5, Timestamp: ThirtyMinAgo.Add(-MetricStep * 2 * time.Second).UnixMilli()},
				{Value: 10, Timestamp: ThirtyMinAgo.Add(-MetricStep * time.Second).UnixMilli()},
			},
		},
	},
}
