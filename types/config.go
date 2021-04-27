package types

import (
	"encoding/json"
	"net/http"
	"time"
)

var InitConfig = _config()

type AllowRequest func(*HttpContext) (int, bool)

type Cors struct {
	Origin               interface{} `json:"origin,omitempty"`
	Methods              interface{} `json:"methods,omitempty"`
	AllowedHeaders       interface{} `json:"allowedHeaders,omitempty"`
	Headers              interface{} `json:"headers,omitempty"`
	ExposedHeaders       interface{} `json:"exposedHeaders,omitempty"`
	MaxAge               string      `json:"maxAge,omitempty"`
	Credentials          bool        `json:"credentials,omitempty"`
	PreflightContinue    bool        `json:"preflightContinue,omitempty"`
	OptionsSuccessStatus int         `json:"optionsSuccessStatus,omitempty"`
}

type Config struct {
	WsEngine          *string            `json:"wsEngine,omitempty"`
	PingTimeout       *time.Duration     `json:"pingTimeout,omitempty"`
	PingInterval      *time.Duration     `json:"pingInterval,omitempty"`
	UpgradeTimeout    *time.Duration     `json:"upgradeTimeout,omitempty"`
	MaxHttpBufferSize *int               `json:"maxHttpBufferSize,omitempty"`
	Transports        *Set               `json:"transports,omitempty"`
	AllowUpgrades     *bool              `json:"allowUpgrades,omitempty"`
	AllowRequest      *AllowRequest      `json:"allowRequest,omitempty"`
	Cookie            *http.Cookie       `json:"cookie,omitempty"`
	PerMessageDeflate *PerMessageDeflate `json:"perMessageDeflate,omitempty"`
	HttpCompression   *HttpCompression   `json:"httpCompression,omitempty"`
	Cors              *Cors              `json:"cors,omitempty"`
	AllowEIO3         *bool              `json:"allowEIO3,omitempty"`
}

func (c *Cors) Assign(data *Cors) error {
	if data == nil {
		return nil
	}
	if buf, err := json.Marshal(data); err != nil {
		return err
	} else {
		return json.Unmarshal(buf, c)
	}
}

func _config() *Config {
	PingTimeout := time.Duration(20000 * time.Millisecond)
	PingInterval := time.Duration(25000 * time.Millisecond)
	UpgradeTimeout := time.Duration(10000 * time.Millisecond)
	MaxHttpBufferSize := int(1e6)
	AllowUpgrades := true
	AllowEIO3 := false
	return &Config{
		// WsEngine: DEFAULT_WS_ENGINE,
		PingTimeout:       &PingTimeout,
		PingInterval:      &PingInterval,
		UpgradeTimeout:    &UpgradeTimeout,
		MaxHttpBufferSize: &MaxHttpBufferSize,
		Transports:        &Set{"polling": NULL, "websocket": NULL},
		AllowUpgrades:     &AllowUpgrades,
		Cookie: &http.Cookie{
			Name:     "io",
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
		PerMessageDeflate: &PerMessageDeflate{
			Threshold: 1024,
		},
		HttpCompression: &HttpCompression{
			Threshold: 1024,
		},
		Cors:      nil,
		AllowEIO3: &AllowEIO3,
	}
}

func (o *Config) Assign(data *Config) error {
	if data == nil {
		return nil
	}
	if buf, err := json.Marshal(data); err != nil {
		return err
	} else {
		return json.Unmarshal(buf, o)
	}
}
