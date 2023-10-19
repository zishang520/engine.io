package types

type (
	Void struct{}

	Callable func()

	HttpCompression struct {
		Threshold int `json:"threshold,omitempty" mapstructure:"threshold,omitempty" msgpack:"threshold,omitempty"`
	}

	PerMessageDeflate struct {
		Threshold int `json:"threshold,omitempty" mapstructure:"threshold,omitempty" msgpack:"threshold,omitempty"`
	}
)

var (
	NULL Void
)
