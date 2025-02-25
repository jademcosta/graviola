package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/jademcosta/graviola/test/e2e/fixtures"
	"github.com/prometheus/prometheus/prompb"
)

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

func sendMetricsToPrometheus(writeRequest *prompb.WriteRequest, prometheusURL string) error {

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

func execRemoteQuery(
	prometheusURL string, path string, query string, start time.Time, end time.Time,
) ([]byte, error) {
	formData := url.Values{
		"query": {query},
		"start": {start.Format(time.RFC3339)},
		"end":   {end.Format(time.RFC3339)},
		"step":  {fmt.Sprintf("%ds", fixtures.MetricStep)},
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

func labelsToQuery(labels []prompb.Label) string {
	query := ""
	for _, lbl := range labels {
		query += fmt.Sprintf(`%s="%s",`, lbl.Name, lbl.Value)
	}
	query = strings.TrimRight(query, ",")

	return fmt.Sprintf("{%s}", query)
}
