package types

type (
	CodeMessage struct {
		Code    int    `json:"code" mapstructure:"code" msgpack:"code"`
		Message string `json:"message,omitempty" mapstructure:"message,omitempty" msgpack:"message,omitempty"`
	}

	ErrorMessage struct {
		*CodeMessage

		Req     *HttpContext   `json:"req,omitempty" mapstructure:"req,omitempty" msgpack:"req,omitempty"`
		Context map[string]any `json:"context,omitempty" mapstructure:"context,omitempty" msgpack:"context,omitempty"`
	}
)
