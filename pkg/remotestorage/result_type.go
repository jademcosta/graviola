package remotestorage

import "github.com/prometheus/common/model"

type ShallowResultType struct {
	ResultType string      `json:"resultType"`
	Result     interface{} `json:"result"`
}

// type StringResultType struct {
// 	ResultType string    `json:"resultType"`
// 	Result     [2]string `json:"result"`
// }

//TODO: parse stats

type RemoteData struct {
	ResultType string       `json:"resultType"`
	Result     model.Vector `json:"result"`
}
