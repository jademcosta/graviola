// nolint:forbidgo
package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/jademcosta/graviola/test/e2e/fixtures"
	"github.com/prometheus/prometheus/prompb"
)

const apiVersionPath = "/api/v1"
const prometheus1URLWithPath = "http://prometheus_1:9091" + apiVersionPath
const graviolaURLWithPath = "http://localhost:9197" + apiVersionPath

var httpCli = &http.Client{
	Timeout: 2 * time.Second,
}

func main() {
	err := sendMetricsTo(fixtures.SingleMetricCounterFixture, prometheus1URLWithPath)
	if err != nil {
		panic(err)
	}

	for idx, ts := range fixtures.SingleMetricCounterFixture.Timeseries {
		err = checkMetricsMatch(ts, graviolaURLWithPath, prometheus1URLWithPath)
		if err != nil {
			panic(fmt.Errorf("comparsion error on index %d: %w", idx, err))
		}
	}
}

func buildWriteRequest(
	timeSeries []prompb.TimeSeries, metadata []prompb.MetricMetadata,
) ([]byte, error) {

	req := &prompb.WriteRequest{
		Timeseries: timeSeries,
		Metadata:   metadata,
	}

	pBuf := proto.NewBuffer(nil)
	err := pBuf.Marshal(req)
	if err != nil {
		return nil, err
	}

	buf := &[]byte{}

	compressed, err := compressPayload(buf, pBuf.Bytes())
	if err != nil {
		return nil, err
	}
	return compressed, nil
}

func compressPayload(tmpbuf *[]byte, inp []byte) ([]byte, error) {
	compressed := snappy.Encode(*tmpbuf, inp) //TODO: improve?
	if n := snappy.MaxEncodedLen(len(inp)); n > len(*tmpbuf) {
		// grow the buffer for the next time
		*tmpbuf = make([]byte, n)
	}
	return compressed, nil
}

func successStatus(statusCode int) bool {
	return statusCode/200 == 1
}

func sendMetricsTo(writeRequest *prompb.WriteRequest, prometheusURL string) error {

	request1, err := buildWriteRequest(writeRequest.Timeseries, nil)
	if err != nil {
		return err
	}

	req1, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/write", prometheusURL), bytes.NewReader(request1))
	if err != nil {
		return err
	}

	resp, err := httpCli.Do(req1)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !successStatus(resp.StatusCode) {
		return fmt.Errorf("status code should be 2XX but it is %d", resp.StatusCode)
	}

	return nil
}

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

func labelsToQuery(labels []prompb.Label) string {
	query := ""
	for _, lbl := range labels {
		query += fmt.Sprintf(`%s="%s",`, lbl.Name, lbl.Value)
	}
	query = strings.TrimRight(query, ",")

	return fmt.Sprintf("{%s}", query)
}

func execRemoteQuery(
	prometheusURL string, path string, query string, start time.Time, end time.Time,
) ([]byte, error) {
	formData := url.Values{
		"query": {query},
		"start": {start.Format(time.RFC3339)},
		"end":   {end.Format(time.RFC3339)},
		"step":  {"15s"},
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", prometheusURL, path), strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error when reading body (and status was %d): %w", response.StatusCode, err)
	}

	if !successStatus(response.StatusCode) {
		return body, fmt.Errorf("response status code is %d", response.StatusCode)
	}

	return body, nil
}
