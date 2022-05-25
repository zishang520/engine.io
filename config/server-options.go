package config

import (
	"github.com/imdario/mergo"
	"github.com/zishang520/engine.io/types"
	"io"
	"net/http"
	"time"
)

type AllowRequest func(*types.HttpContext) (int, map[string]interface{})

type ServerOptionsInterface interface {
	SetPingTimeout(time.Duration)
	PingTimeout() time.Duration
	SetPingInterval(time.Duration)
	PingInterval() time.Duration
	SetUpgradeTimeout(time.Duration)
	UpgradeTimeout() time.Duration
	SetMaxHttpBufferSize(int64)
	MaxHttpBufferSize() int64
	SetAllowRequest(AllowRequest)
	AllowRequest() AllowRequest
	SetTransports(*types.Set[string])
	Transports() *types.Set[string]
	SetAllowUpgrades(bool)
	AllowUpgrades() bool
	SetPerMessageDeflate(*types.PerMessageDeflate)
	PerMessageDeflate() *types.PerMessageDeflate
	SetHttpCompression(*types.HttpCompression)
	HttpCompression() *types.HttpCompression
	SetInitialPacket(io.Reader)
	InitialPacket() io.Reader
	SetCookie(*http.Cookie)
	Cookie() *http.Cookie
	SetCors(*types.Cors)
	Cors() *types.Cors
	SetAllowEIO3(bool)
	AllowEIO3() bool
}

type ServerOptions struct {
	// how many ms without a pong packet to consider the connection closed
	InternalPingTimeout *time.Duration `json:"pingTimeout,omitempty"`

	// how many ms before sending a new ping packet
	InternalPingInterval *time.Duration `json:"pingInterval,omitempty"`

	// how many ms before an uncompleted transport upgrade is cancelled
	InternalUpgradeTimeout *time.Duration `json:"upgradeTimeout,omitempty"`

	// how many bytes or characters a message can be, before closing the session (to avoid DoS).
	InternalMaxHttpBufferSize *int64 `json:"maxHttpBufferSize,omitempty"`

	// A function that receives a given handshake or upgrade request as its first parameter,
	// and can decide whether to continue or not. The second argument is a function that needs
	// to be called with the decided information: fn(err, success), where success is a boolean
	// value where false means that the request is rejected, and err is an error code.
	InternalAllowRequest *AllowRequest `json:"allowRequest,omitempty"`

	// the low-level transports that are enabled
	InternalTransports *types.Set[string] `json:"transports,omitempty"`

	// whether to allow transport upgrades
	InternalAllowUpgrades *bool `json:"allowUpgrades,omitempty"`

	// parameters of the WebSocket permessage-deflate extension (see ws module api docs). Set to false to disable.
	InternalPerMessageDeflate *types.PerMessageDeflate `json:"perMessageDeflate,omitempty"`

	// parameters of the http compression for the polling transports (see zlib api docs). Set to false to disable.
	InternalHttpCompression *types.HttpCompression `json:"httpCompression,omitempty"`

	// wsEngine is not supported
	// wsEngine

	// an optional packet which will be concatenated to the handshake packet emitted by Engine.IO.
	InternalInitialPacket io.Reader `json:"initialPacket,omitempty"`

	// configuration of the cookie that contains the client sid to send as part of handshake response headers. This cookie
	// might be used for sticky-session. Defaults to not sending any cookie.
	InternalCookie *http.Cookie `json:"cookie,omitempty"`

	// the options that will be forwarded to the cors module
	InternalCors *types.Cors `json:"cors,omitempty"`

	// whether to enable compatibility with Socket.IO v2 clients
	InternalAllowEIO3 *bool `json:"allowEIO3,omitempty"`
}

func DefaultServerOptions() *ServerOptions {
	s := &ServerOptions{}
	s.SetPingTimeout(time.Duration(20000 * time.Millisecond))
	s.SetPingInterval(time.Duration(25000 * time.Millisecond))
	s.SetUpgradeTimeout(time.Duration(10000 * time.Millisecond))
	s.SetMaxHttpBufferSize(int64(1e6))
	s.SetTransports(types.NewSet("polling", "websocket"))
	s.SetAllowUpgrades(true)
	s.SetHttpCompression(&types.HttpCompression{
		Threshold: 1024,
	})
	s.SetCors(nil)
	s.SetAllowEIO3(false)
	return s
}

func (s *ServerOptions) Assign(data ServerOptionsInterface) (ServerOptionsInterface, error) {
	if data == nil {
		return s, nil
	}
	if err := mergo.Merge(data, *s, mergo.WithOverrideEmptySlice); err != nil {
		return nil, err
	}
	return data, nil
}

// how many ms without a pong packet to consider the connection closed
// @default 20000
func (s *ServerOptions) SetPingTimeout(pingTimeout time.Duration) {
	s.InternalPingTimeout = &pingTimeout
}

func (s *ServerOptions) PingTimeout() time.Duration {
	if s.InternalPingTimeout == nil {
		return time.Duration(20000 * time.Millisecond)
	}
	return *s.InternalPingTimeout
}

// how many ms before sending a new ping packet
// @default 25000
func (s *ServerOptions) SetPingInterval(pingInterval time.Duration) {
	s.InternalPingInterval = &pingInterval
}
func (s *ServerOptions) PingInterval() time.Duration {
	if s.InternalPingInterval == nil {
		return time.Duration(25000 * time.Millisecond)
	}

	return *s.InternalPingInterval
}

// how many ms before an uncompleted transport upgrade is cancelled
// @default 10000
func (s *ServerOptions) SetUpgradeTimeout(upgradeTimeout time.Duration) {
	s.InternalUpgradeTimeout = &upgradeTimeout
}
func (s *ServerOptions) UpgradeTimeout() time.Duration {
	if s.InternalUpgradeTimeout == nil {
		return time.Duration(10000 * time.Millisecond)
	}
	return *s.InternalUpgradeTimeout
}

// how many bytes or characters a message can be, before closing the session (to avoid DoS).
// @default 1e5 (100 KB)
func (s *ServerOptions) SetMaxHttpBufferSize(maxHttpBufferSize int64) {
	s.InternalMaxHttpBufferSize = &maxHttpBufferSize
}
func (s *ServerOptions) MaxHttpBufferSize() int64 {
	if s.InternalMaxHttpBufferSize == nil {
		return 1e5
	}
	return *s.InternalMaxHttpBufferSize
}

// A function that receives a given handshake or upgrade request as its first parameter,
// and can decide whether to continue or not. The second argument is a function that needs
// to be called with the decided information: fn(err, success), where success is a boolean
// value where false means that the request is rejected, and err is an error code.
func (s *ServerOptions) SetAllowRequest(allowRequest AllowRequest) {
	s.InternalAllowRequest = &allowRequest
}
func (s *ServerOptions) AllowRequest() AllowRequest {
	if s.InternalAllowRequest == nil {
		return nil
	}
	return *s.InternalAllowRequest
}

// the low-level transports that are enabled
// @default ["polling", "websocket"]
func (s *ServerOptions) SetTransports(transports *types.Set[string]) {
	s.InternalTransports = transports
}
func (s *ServerOptions) Transports() *types.Set[string] {
	if s.InternalTransports == nil {
		return types.NewSet("polling", "websocket")
	}
	return s.InternalTransports
}

// whether to allow transport upgrades
// @default true
func (s *ServerOptions) SetAllowUpgrades(allowUpgrades bool) {
	s.InternalAllowUpgrades = &allowUpgrades
}
func (s *ServerOptions) AllowUpgrades() bool {
	if s.InternalAllowUpgrades == nil {
		return true
	}
	return *s.InternalAllowUpgrades
}

// parameters of the WebSocket permessage-deflate extension (see ws module api docs). Set to false to disable.
// @default nil
func (s *ServerOptions) SetPerMessageDeflate(perMessageDeflate *types.PerMessageDeflate) {
	s.InternalPerMessageDeflate = perMessageDeflate
}
func (s *ServerOptions) PerMessageDeflate() *types.PerMessageDeflate {
	return s.InternalPerMessageDeflate
}

// parameters of the http compression for the polling transports (see zlib api docs). Set to false to disable.
// @default true
func (s *ServerOptions) SetHttpCompression(httpCompression *types.HttpCompression) {
	s.InternalHttpCompression = httpCompression
}
func (s *ServerOptions) HttpCompression() *types.HttpCompression {
	if s.InternalHttpCompression == nil {
		return &types.HttpCompression{
			Threshold: 1024,
		}
	}
	return s.InternalHttpCompression
}

// an optional packet which will be concatenated to the handshake packet emitted by Engine.IO.
func (s *ServerOptions) SetInitialPacket(initialPacket io.Reader) {
	s.InternalInitialPacket = initialPacket
}
func (s *ServerOptions) InitialPacket() io.Reader {
	return s.InternalInitialPacket
}

// configuration of the cookie that contains the client sid to send as part of handshake response headers. This cookie
// might be used for sticky-session. Defaults to not sending any cookie.
// @default false
func (s *ServerOptions) SetCookie(cookie *http.Cookie) {
	s.InternalCookie = cookie
}
func (s *ServerOptions) Cookie() *http.Cookie {
	return s.InternalCookie
}

// the options that will be forwarded to the cors module
func (s *ServerOptions) SetCors(cors *types.Cors) {
	s.InternalCors = cors
}
func (s *ServerOptions) Cors() *types.Cors {
	return s.InternalCors
}

// whether to enable compatibility with Socket.IO v2 clients
// @default false
func (s *ServerOptions) SetAllowEIO3(allowEIO3 bool) {
	s.InternalAllowEIO3 = &allowEIO3
}
func (s *ServerOptions) AllowEIO3() bool {
	if s.InternalAllowEIO3 == nil {
		return false
	}

	return *s.InternalAllowEIO3
}
