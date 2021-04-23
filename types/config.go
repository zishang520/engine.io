package types

import (
	"encoding/json"
	"time"
)

var InitConfig = _config()

type AllowRequest func(*HttpContext) (int, bool)

type Cookie struct {
	Name     string `json:"name,omitempty"`
	Path     string `json:"path,omitempty"`
	HttpOnly bool   `json:"httpOnly,omitempty"`
	SameSite string `json:"sameSite,omitempty"`
}

type Cors struct {
	Origin               interface{} `json:"origin,omitempty"`
	Methods              interface{} `json:"methods,omitempty"`
	AllowedHeaders       interface{} `json:"allowedHeaders,omitempty"`
	Headers              interface{} `json:"headers,omitempty"`
	ExposedHeaders       interface{} `json:"exposedHeaders,omitempty"`
	Credentials          bool        `json:"credentials"`
	PreflightContinue    bool        `json:"preflightContinue"`
	OptionsSuccessStatus int         `json:"optionsSuccessStatus"`
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
	Cookie            *Cookie            `json:"cookie,omitempty"`
	PerMessageDeflate *PerMessageDeflate `json:"perMessageDeflate,omitempty"`
	HttpCompression   *HttpCompression   `json:"httpCompression,omitempty"`
	Cors              *Cors              `json:"cors,omitempty"`
	AllowEIO3         *bool              `json:"allowEIO3,omitempty"`
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
		Cookie: &Cookie{
			Name:     "io",
			Path:     "/",
			HttpOnly: true,
			SameSite: "lax",
		},
		PerMessageDeflate: &PerMessageDeflate{
			Threshold: 1024,
		},
		HttpCompression: &HttpCompression{
			Threshold: 1024,
		},
		Cors: &Cors{
			Origin: "*",
		},
		AllowEIO3: &AllowEIO3,
	}
}

func (o *Config) Assign(data *Config) error {
	if buf, err := json.Marshal(data); err != nil {
		return err
	} else {
		return json.Unmarshal(buf, o)
	}
}
