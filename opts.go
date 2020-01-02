package engineio

import (
	"github.com/zishang520/engine.io/types"
)

type AllowRequest func()

type Opts struct {
	WsEngine          string        `json:"wsEngine"`
	PingTimeout       int           `json:"pingTimeout"`
	PingInterval      int           `json:"pingInterval"`
	UpgradeTimeout    int           `json:"upgradeTimeout"`
	MaxHttpBufferSize int           `json:"maxHttpBufferSize"`
	Transports        []string      `json:"transports"`
	AllowUpgrades     bool          `json:"allowUpgrades"`
	AllowRequest      AllowRequest  `json:"allowRequest"`
	Cookie            string        `json:"cookie"`
	CookiePath        string        `json:"cookiePath"`
	CookieHttpOnly    string        `json:"cookieHttpOnly"`
	PerMessageDeflate string        `json:"perMessageDeflate"`
	HttpCompression   string        `json:"httpCompression"`
	InitialPacket     *types.Packet `json:"initialPacket"`
}
