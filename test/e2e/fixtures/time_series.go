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

var SingleCounterMetric2 = &prompb.WriteRequest{
	Timeseries: []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "http_requests_total"},
				{Name: "region", Value: "us-east-1"},
				{Name: "system", Value: "cd"},
				{Name: "job", Value: "sys"},
			},
			Samples: []prompb.Sample{
				{Value: 15, Timestamp: ThirtyMinAgo.Add(-MetricStep * 6 * time.Second).UnixMilli()},
				{Value: 17, Timestamp: ThirtyMinAgo.Add(-MetricStep * 5 * time.Second).UnixMilli()},
				{Value: 30, Timestamp: ThirtyMinAgo.Add(-MetricStep * 4 * time.Second).UnixMilli()},
				{Value: 45, Timestamp: ThirtyMinAgo.Add(-MetricStep * 3 * time.Second).UnixMilli()},
				{Value: 90, Timestamp: ThirtyMinAgo.Add(-MetricStep * 2 * time.Second).UnixMilli()},
				{Value: 100, Timestamp: ThirtyMinAgo.Add(-MetricStep * time.Second).UnixMilli()},
			},
		},
	},
}

var SingleCounterMetric3 = &prompb.WriteRequest{
	Timeseries: []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "http_requests_total"},
				{Name: "region", Value: "sa-east-1"},
				{Name: "system", Value: "ab"},
				{Name: "job", Value: "sys"},
			},
			Samples: []prompb.Sample{
				{Value: 24, Timestamp: ThirtyMinAgo.Add(-MetricStep * 6 * time.Second).UnixMilli()},
				{Value: 29, Timestamp: ThirtyMinAgo.Add(-MetricStep * 5 * time.Second).UnixMilli()},
				{Value: 44, Timestamp: ThirtyMinAgo.Add(-MetricStep * 4 * time.Second).UnixMilli()},
				{Value: 59, Timestamp: ThirtyMinAgo.Add(-MetricStep * 3 * time.Second).UnixMilli()},
				{Value: 86, Timestamp: ThirtyMinAgo.Add(-MetricStep * 2 * time.Second).UnixMilli()},
				{Value: 113, Timestamp: ThirtyMinAgo.Add(-MetricStep * time.Second).UnixMilli()},
			},
		},
	},
}
