package batch

type BatchPayload struct {
	FailOnErrors *uint       `json:"failOnErrors"`
	Operations   []Operation `json:"operations"`
}

type Operation struct {
	Method  string      `json:"method"`
	Path    string      `json:"path"`
	Headers []Header    `json:"headers"`
	BulkId  string      `json:"bulk_id"`
	Body    interface{} `json:"body"`
	Status  Status      `json:"status"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Status struct {
	Code    string `json:"code"`
	CodeInt int    `json:"-"`
}
