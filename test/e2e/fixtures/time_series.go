package fixtures

import (
	"time"

	"github.com/prometheus/prometheus/prompb"
)

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
				{Value: 1, Timestamp: ThirtyMinAgo.Add(-90 * time.Second).UnixMilli()},
				{Value: 2, Timestamp: ThirtyMinAgo.Add(-75 * time.Second).UnixMilli()},
				{Value: 3, Timestamp: ThirtyMinAgo.Add(-60 * time.Second).UnixMilli()},
				{Value: 4, Timestamp: ThirtyMinAgo.Add(-45 * time.Second).UnixMilli()},
				{Value: 5, Timestamp: ThirtyMinAgo.Add(-30 * time.Second).UnixMilli()},
				{Value: 10, Timestamp: ThirtyMinAgo.Add(-15 * time.Second).UnixMilli()},
			},
		},
	},
}
