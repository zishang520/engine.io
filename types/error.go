package types

type CodeMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type ErrorMessage struct {
	*CodeMessage

	Req     *HttpContext   `json:"req,omitempty"`
	Context map[string]any `json:"context,omitempty"`
}
