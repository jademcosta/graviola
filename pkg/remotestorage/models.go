package remotestorage

type LabelNamesResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
	Error  string   `json:"error,omitempty"`
}
