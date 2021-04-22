package types

import (
	"time"
)

type AllowRequest func()

type Cookie struct {
	Name     string `json:"name"`     // "io",
	Path     string `json:"path"`     // "/",
	HttpOnly string `json:"httpOnly"` // opts.cookie.path !== false,
	SameSite string `json:"sameSite"` // "lax"
}

type Opts struct {
	WsEngine          string            `json:"wsEngine"`
	PingTimeout       time.Duration     `json:"pingTimeout"`
	PingInterval      time.Duration     `json:"pingInterval"`
	UpgradeTimeout    time.Duration     `json:"upgradeTimeout"`
	MaxHttpBufferSize time.Duration     `json:"maxHttpBufferSize"`
	Transports        Set               `json:"transports"`
	AllowUpgrades     bool              `json:"allowUpgrades"`
	AllowRequest      AllowRequest      `json:"allowRequest"`
	Cookie            Cookie            `json:"cookie"`
	CookiePath        string            `json:"cookiePath"`
	CookieHttpOnly    string            `json:"cookieHttpOnly"`
	PerMessageDeflate PerMessageDeflate `json:"perMessageDeflate"`
	HttpCompression   HttpCompression   `json:"httpCompression"`
	Cors              bool              `json:"cors"`
	AllowEIO3         bool              `json:"allowEIO3"`
}

func OptsInit() *Opts {
	return &Opts{}
}

func (o *Opts) Assign() {

}
