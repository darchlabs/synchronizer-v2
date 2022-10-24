package api

type Response struct {
	Data interface{} `json:"data,omitempty"`
	Meta interface{} `json:"meta,omitempty"`
	Error interface{} `json:"error,omitempty"`
}
