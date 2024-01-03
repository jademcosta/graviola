package app_test

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jademcosta/graviola/pkg/remotestorage"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
)

func doRequest(address string, hints storage.SelectHints, lbls ...*labels.Matcher) *http.Response {

	promQLQuery, err := remotestorage.ToPromQLQuery(lbls)
	if err != nil {
		panic(err)
	}

	params := url.Values{}
	params.Set("query", *promQLQuery)

	if hints.Start != 0 {
		params.Set("start", strconv.FormatInt(hints.Start, 10))
	}

	if hints.End != 0 {
		params.Set("end", strconv.FormatInt(hints.End, 10))
	}

	if hints.Step != 0 {
		params.Set("step", strconv.FormatInt(hints.Step, 10))
	}

	req, err := http.NewRequest(http.MethodPost, address, strings.NewReader(params.Encode()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	return resp
}
