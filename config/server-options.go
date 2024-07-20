package config

import (
	"io"
	"net/http"
	"time"

	"github.com/zishang520/engine.io/v2/transports"
	"github.com/zishang520/engine.io/v2/types"
)

type (
	AllowRequest func(*types.HttpContext) error

	ServerOptionsInterface interface {
		SetPingTimeout(time.Duration)
		GetRawPingTimeout() *time.Duration
		PingTimeout() time.Duration

		SetPingInterval(time.Duration)
		GetRawPingInterval() *time.Duration
		PingInterval() time.Duration

		SetUpgradeTimeout(time.Duration)
		GetRawUpgradeTimeout() *time.Duration
		UpgradeTimeout() time.Duration

		SetMaxHttpBufferSize(int64)
		GetRawMaxHttpBufferSize() *int64
		MaxHttpBufferSize() int64

		SetAllowRequest(AllowRequest)
		GetRawAllowRequest() AllowRequest
		AllowRequest() AllowRequest

		SetTransports(*types.Set[string])
		GetRawTransports() *types.Set[string]
		Transports() *types.Set[string]

		SetAllowUpgrades(bool)
		GetRawAllowUpgrades() *bool
		AllowUpgrades() bool

		SetPerMessageDeflate(*types.PerMessageDeflate)
		GetRawPerMessageDeflate() *types.PerMessageDeflate
		PerMessageDeflate() *types.PerMessageDeflate

		SetHttpCompression(*types.HttpCompression)
		GetRawHttpCompression() *types.HttpCompression
		HttpCompression() *types.HttpCompression

		SetInitialPacket(io.Reader)
		GetRawInitialPacket() io.Reader
		InitialPacket() io.Reader

		SetCookie(*http.Cookie)
		GetRawCookie() *http.Cookie
		Cookie() *http.Cookie

		SetCors(*types.Cors)
		GetRawCors() *types.Cors
		Cors() *types.Cors

		SetAllowEIO3(bool)
		GetRawAllowEIO3() *bool
		AllowEIO3() bool
	}

	ServerOptions struct {
		// how many ms without a pong packet to consider the connection closed
		pingTimeout *time.Duration

		// how many ms before sending a new ping packet
		pingInterval *time.Duration

		// how many ms before an uncompleted transport upgrade is cancelled
		upgradeTimeout *time.Duration

		// how many bytes or characters a message can be, before closing the session (to avoid DoS).
		maxHttpBufferSize *int64

		// A function that receives a given handshake or upgrade request as its first parameter,
		// and can decide whether to continue. Returning an error indicates that the request was rejected.
		allowRequest AllowRequest

		// the low-level transports that are enabled
		transports *types.Set[string]

		// whether to allow transport upgrades
		allowUpgrades *bool

		// parameters of the WebSocket permessage-deflate extension (see ws module api docs). Set to false to disable.
		perMessageDeflate *types.PerMessageDeflate

		// parameters of the http compression for the polling transports (see zlib api docs). Set to false to disable.
		httpCompression *types.HttpCompression

		// wsEngine is not supported
		// wsEngine

		// an optional packet which will be concatenated to the handshake packet emitted by Engine.IO.
		initialPacket io.Reader

		// configuration of the cookie that contains the client sid to send as part of handshake response headers. This cookie
		// might be used for sticky-session. Defaults to not sending any cookie.
		cookie *http.Cookie

		// the options that will be forwarded to the cors module
		cors *types.Cors

		// whether to enable compatibility with Socket.IO v2 clients
		allowEIO3 *bool
	}
)

func DefaultServerOptions() *ServerOptions {
	s := &ServerOptions{}
	return s
}

func (s *ServerOptions) Assign(data ServerOptionsInterface) ServerOptionsInterface {
	if data == nil {
		return s
	}

	if s.GetRawPingTimeout() == nil {
		s.SetPingTimeout(data.PingTimeout())
	}
	if s.GetRawPingInterval() == nil {
		s.SetPingInterval(data.PingInterval())
	}
	if s.GetRawUpgradeTimeout() == nil {
		s.SetUpgradeTimeout(data.UpgradeTimeout())
	}
	if s.GetRawMaxHttpBufferSize() == nil {
		s.SetMaxHttpBufferSize(data.MaxHttpBufferSize())
	}
	if s.GetRawAllowRequest() == nil {
		s.SetAllowRequest(data.AllowRequest())
	}
	if s.GetRawTransports() == nil {
		s.SetTransports(data.Transports())
	}
	if s.GetRawAllowUpgrades() == nil {
		s.SetAllowUpgrades(data.AllowUpgrades())
	}
	if s.GetRawPerMessageDeflate() == nil {
		s.SetPerMessageDeflate(data.PerMessageDeflate())
	}
	if s.GetRawHttpCompression() == nil {
		s.SetHttpCompression(data.HttpCompression())
	}
	if s.GetRawInitialPacket() == nil {
		s.SetInitialPacket(data.InitialPacket())
	}
	if s.GetRawCookie() == nil {
		s.SetCookie(data.Cookie())
	}
	if s.GetRawCors() == nil {
		s.SetCors(data.Cors())
	}
	if s.GetRawAllowEIO3() == nil {
		s.SetAllowEIO3(data.AllowEIO3())
	}

	return s
}

// how many ms without a pong packet to consider the connection closed
// @default 20_000
func (s *ServerOptions) SetPingTimeout(pingTimeout time.Duration) {
	s.pingTimeout = &pingTimeout
}
func (s *ServerOptions) GetRawPingTimeout() *time.Duration {
	return s.pingTimeout
}
func (s *ServerOptions) PingTimeout() time.Duration {
	if s.pingTimeout == nil {
		return time.Duration(20_000 * time.Millisecond)
	}
	return *s.pingTimeout
}

// how many ms before sending a new ping packet
// @default 25_000
func (s *ServerOptions) SetPingInterval(pingInterval time.Duration) {
	s.pingInterval = &pingInterval
}
func (s *ServerOptions) GetRawPingInterval() *time.Duration {
	return s.pingInterval
}
func (s *ServerOptions) PingInterval() time.Duration {
	if s.pingInterval == nil {
		return time.Duration(25_000 * time.Millisecond)
	}

	return *s.pingInterval
}

// how many ms before an uncompleted transport upgrade is cancelled
// @default 10_000
func (s *ServerOptions) SetUpgradeTimeout(upgradeTimeout time.Duration) {
	s.upgradeTimeout = &upgradeTimeout
}
func (s *ServerOptions) GetRawUpgradeTimeout() *time.Duration {
	return s.upgradeTimeout
}
func (s *ServerOptions) UpgradeTimeout() time.Duration {
	if s.upgradeTimeout == nil {
		return time.Duration(10_000 * time.Millisecond)
	}
	return *s.upgradeTimeout
}

// how many bytes or characters a message can be, before closing the session (to avoid DoS).
// @default 1e5 (100 KB)
func (s *ServerOptions) SetMaxHttpBufferSize(maxHttpBufferSize int64) {
	s.maxHttpBufferSize = &maxHttpBufferSize
}
func (s *ServerOptions) GetRawMaxHttpBufferSize() *int64 {
	return s.maxHttpBufferSize
}
func (s *ServerOptions) MaxHttpBufferSize() int64 {
	if s.maxHttpBufferSize == nil {
		return 1e5
	}
	return *s.maxHttpBufferSize
}

// A function that receives a given handshake or upgrade request as its first parameter,
// and can decide whether to continue or not. The second argument is a function that needs
// to be called with the decided information: fn(err, success), where success is a boolean
// value where false means that the request is rejected, and err is an error code.
func (s *ServerOptions) SetAllowRequest(allowRequest AllowRequest) {
	s.allowRequest = allowRequest
}
func (s *ServerOptions) GetRawAllowRequest() AllowRequest {
	return s.allowRequest
}
func (s *ServerOptions) AllowRequest() AllowRequest {
	if s.allowRequest == nil {
		return nil
	}
	return s.allowRequest
}

// The low-level transports that are enabled. WebTransport is disabled by default and must be manually enabled:
//
//	opts := &ServerOptions{}
//	opts.SetTransports(types.NewSet("polling", "websocket", "webtransport"))
//	NewServer(opts)
//
// @default ["polling", "websocket"]
func (s *ServerOptions) SetTransports(transports *types.Set[string]) {
	s.transports = transports
}
func (s *ServerOptions) GetRawTransports() *types.Set[string] {
	return s.transports
}
func (s *ServerOptions) Transports() *types.Set[string] {
	if s.transports == nil {
		return types.NewSet(transports.POLLING, transports.WEBSOCKET)
	}
	return s.transports
}

// whether to allow transport upgrades
// @default true
func (s *ServerOptions) SetAllowUpgrades(allowUpgrades bool) {
	s.allowUpgrades = &allowUpgrades
}
func (s *ServerOptions) GetRawAllowUpgrades() *bool {
	return s.allowUpgrades
}
func (s *ServerOptions) AllowUpgrades() bool {
	if s.allowUpgrades == nil {
		return true
	}
	return *s.allowUpgrades
}

// parameters of the WebSocket permessage-deflate extension (see ws module api docs). Set to false to disable.
// @default nil
func (s *ServerOptions) SetPerMessageDeflate(perMessageDeflate *types.PerMessageDeflate) {
	s.perMessageDeflate = perMessageDeflate
}
func (s *ServerOptions) GetRawPerMessageDeflate() *types.PerMessageDeflate {
	return s.perMessageDeflate
}
func (s *ServerOptions) PerMessageDeflate() *types.PerMessageDeflate {
	return s.perMessageDeflate
}

// parameters of the http compression for the polling transports (see zlib api docs). Set to false to disable.
// @default true
func (s *ServerOptions) SetHttpCompression(httpCompression *types.HttpCompression) {
	s.httpCompression = httpCompression
}
func (s *ServerOptions) GetRawHttpCompression() *types.HttpCompression {
	return s.httpCompression
}
func (s *ServerOptions) HttpCompression() *types.HttpCompression {
	if s.httpCompression == nil {
		return &types.HttpCompression{
			Threshold: 1024,
		}
	}
	return s.httpCompression
}

// an optional packet which will be concatenated to the handshake packet emitted by Engine.IO.
func (s *ServerOptions) SetInitialPacket(initialPacket io.Reader) {
	s.initialPacket = initialPacket
}
func (s *ServerOptions) GetRawInitialPacket() io.Reader {
	return s.initialPacket
}
func (s *ServerOptions) InitialPacket() io.Reader {
	return s.initialPacket
}

// configuration of the cookie that contains the client sid to send as part of handshake response headers. This cookie
// might be used for sticky-session. Defaults to not sending any cookie.
// @default false
func (s *ServerOptions) SetCookie(cookie *http.Cookie) {
	s.cookie = cookie
}
func (s *ServerOptions) GetRawCookie() *http.Cookie {
	return s.cookie
}
func (s *ServerOptions) Cookie() *http.Cookie {
	return s.cookie
}

// the options that will be forwarded to the cors module
func (s *ServerOptions) SetCors(cors *types.Cors) {
	s.cors = cors
}
func (s *ServerOptions) GetRawCors() *types.Cors {
	return s.cors
}
func (s *ServerOptions) Cors() *types.Cors {
	return s.cors
}

// whether to enable compatibility with Socket.IO v2 clients
// @default false
func (s *ServerOptions) SetAllowEIO3(allowEIO3 bool) {
	s.allowEIO3 = &allowEIO3
}
func (s *ServerOptions) GetRawAllowEIO3() *bool {
	return s.allowEIO3
}
func (s *ServerOptions) AllowEIO3() bool {
	if s.allowEIO3 == nil {
		return false
	}

	return *s.allowEIO3
}
