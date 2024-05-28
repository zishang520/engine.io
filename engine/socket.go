package engine

import (
	"encoding/json"
	"io"
	"strings"
	"sync/atomic"
	"time"

	"github.com/zishang520/engine.io-go-parser/packet"
	_types "github.com/zishang520/engine.io-go-parser/types"
	"github.com/zishang520/engine.io/v2/events"
	"github.com/zishang520/engine.io/v2/log"
	"github.com/zishang520/engine.io/v2/transports"
	"github.com/zishang520/engine.io/v2/types"
	"github.com/zishang520/engine.io/v2/utils"
)

var socket_log = log.NewLog("engine:socket")

type socket struct {
	events.EventEmitter

	protocol int
	// TODO for the next major release: do not keep the reference to the first HTTP request, as it stays in memory
	request       *types.HttpContext
	remoteAddress string

	readyState atomic.Value
	transport  atomic.Pointer[transports.Transport]

	// This is the session identifier that the client will use in the subsequent HTTP requests. It must not be shared with
	// others parties, as it might lead to session hijacking.
	id                string
	server            BaseServer
	upgrading         atomic.Bool
	upgraded          atomic.Bool
	writeBuffer       *types.Slice[*packet.Packet]
	packetsFn         *types.Slice[func(transports.Transport)]
	sentCallbackFn    *types.Slice[any]
	cleanupFn         *types.Slice[types.Callable]
	pingTimeoutTimer  atomic.Pointer[utils.Timer]
	pingIntervalTimer atomic.Pointer[utils.Timer]
}

func (s *socket) Protocol() int {
	return s.protocol
}

func (s *socket) Upgraded() bool {
	return s.upgraded.Load()
}

func (s *socket) Upgrading() bool {
	return s.upgrading.Load()
}

func (s *socket) Id() string {
	return s.id
}

func (s *socket) RemoteAddress() string {
	return s.remoteAddress
}

func (s *socket) Request() *types.HttpContext {
	return s.request
}

func (s *socket) Transport() transports.Transport {
	if v := s.transport.Load(); v != nil {
		return *v
	}
	return nil
}

func (s *socket) Server() BaseServer {
	return s.server
}

func (s *socket) ReadyState() string {
	if v, ok := s.readyState.Load().(string); ok {
		return v
	}
	return ""
}

func (s *socket) SetReadyState(state string) {
	socket_log.Debug("readyState updated from %s to %s", s.ReadyState(), state)

	s.readyState.Store(state)
}

// Client class.
func MakeSocket() Socket {
	s := &socket{
		EventEmitter: events.New(),

		writeBuffer:    types.NewSlice[*packet.Packet](),
		packetsFn:      types.NewSlice[func(transports.Transport)](),
		sentCallbackFn: types.NewSlice[any](),
		cleanupFn:      types.NewSlice[types.Callable](),
	}
	s.readyState.Store("opening")

	return s
}

// Client class.
func NewSocket(id string, server BaseServer, transport transports.Transport, ctx *types.HttpContext, protocol int) Socket {
	s := MakeSocket()

	s.Construct(id, server, transport, ctx, protocol)

	return s
}

func (s *socket) Construct(id string, server BaseServer, transport transports.Transport, ctx *types.HttpContext, protocol int) {
	s.id = id
	s.server = server
	s.request = ctx
	s.protocol = protocol

	// Cache IP since it might not be in the req later
	if ctx.WebTransport != nil {
		s.remoteAddress = ctx.WebTransport.RemoteAddr().String()
	} else if ctx.Websocket != nil && ctx.Websocket.Conn != nil {
		s.remoteAddress = ctx.Websocket.RemoteAddr().String()
	} else {
		s.remoteAddress = ctx.Request().RemoteAddr
	}

	s.setTransport(transport)
	s.onOpen()
}

// Called upon transport considered open.
func (s *socket) onOpen() {
	s.SetReadyState("open")

	// sends an `open` packet
	s.Transport().SetSid(s.id)

	data, err := json.Marshal(map[string]any{
		"sid":          s.id,
		"upgrades":     s.getAvailableUpgrades(),
		"pingInterval": int64(s.server.Opts().PingInterval() / time.Millisecond),
		"pingTimeout":  int64(s.server.Opts().PingTimeout() / time.Millisecond),
		"maxPayload":   s.server.Opts().MaxHttpBufferSize(),
	})

	if err != nil {
		socket_log.Debug("json.Marshal err")
	}
	s.sendPacket(
		packet.OPEN,
		_types.NewStringBuffer(data),
		nil, nil,
	)

	if i := s.server.Opts().InitialPacket(); i != nil {
		s.sendPacket(packet.MESSAGE, i, nil, nil)
	}

	s.Emit("open")

	if s.protocol == 3 {
		// in protocol v3, the client sends a ping, and the server answers with a pong
		s.resetPingTimeout(s.server.Opts().PingInterval() + s.server.Opts().PingTimeout())
	} else {
		// in protocol v4, the server sends a ping, and the client answers with a pong
		s.schedulePing()
	}
}

// Called upon transport packet.
func (s *socket) onPacket(data *packet.Packet) {
	if "open" != s.ReadyState() {
		socket_log.Debug("packet received with closed socket")
		return
	}

	// export packet event
	socket_log.Debug(`received packet %s`, data.Type)
	s.Emit("packet", data)

	// Reset ping timeout on any packet, incoming data is a good sign of
	// other side's liveness
	s.resetPingTimeout(s.server.Opts().PingInterval() + s.server.Opts().PingTimeout())

	switch data.Type {
	case packet.PING:
		if s.Transport().Protocol() != 3 {
			s.onError("invalid heartbeat direction")
			return
		}
		socket_log.Debug("got ping")
		s.sendPacket(packet.PONG, nil, nil, nil)
		s.Emit("heartbeat")
	case packet.PONG:
		if s.Transport().Protocol() == 3 {
			s.onError("invalid heartbeat direction")
			return
		}
		socket_log.Debug("got pong")
		s.pingIntervalTimer.Load().Refresh()
		s.Emit("heartbeat")
	case packet.ERROR:
		s.OnClose("parse error")
	case packet.MESSAGE:
		s.Emit("data", data.Data)
		s.Emit("message", data.Data)
	}
}

// Called upon transport error.
func (s *socket) onError(err any) {
	socket_log.Debug("transport error %v", err)
	s.OnClose("transport error", err)
}

// Pings client every `this.pingInterval` and expects response
// within `this.pingTimeout` or closes connection.
func (s *socket) schedulePing() {
	s.pingIntervalTimer.Store(utils.SetTimeout(func() {
		socket_log.Debug("writing ping packet - expecting pong within %dms", int64(s.server.Opts().PingTimeout()/time.Millisecond))
		s.sendPacket(packet.PING, nil, nil, nil)
		s.resetPingTimeout(s.server.Opts().PingTimeout())
	}, s.server.Opts().PingInterval()))
}

// Resets ping timeout.
func (s *socket) resetPingTimeout(timeout time.Duration) {
	utils.ClearTimeout(s.pingTimeoutTimer.Load())
	s.pingTimeoutTimer.Store(utils.SetTimeout(func() {
		if s.ReadyState() == "closed" {
			return
		}
		s.OnClose("ping timeout")
	}, timeout))
}

// Attaches handlers for the given transport.
func (s *socket) setTransport(transport transports.Transport) {
	onError := func(err ...any) {
		err = append(err, nil)
		s.onError(err[0])
	}
	onPacket := func(packets ...any) {
		if len(packets) > 0 {
			s.onPacket(packets[0].(*packet.Packet))
		}
	}
	flush := func(...any) { s.flush() }
	onClose := func(...any) { s.OnClose("transport close") }

	s.transport.Store(&transport)

	transport.Once("error", onError)
	transport.On("packet", onPacket)
	transport.On("drain", flush)
	transport.Once("close", onClose)

	// s function will manage packet events (also message callbacks)
	s.setupSendCallback()

	s.cleanupFn.Push(func() {
		transport.RemoveListener("error", onError)
		transport.RemoveListener("packet", onPacket)
		transport.RemoveListener("drain", flush)
		transport.RemoveListener("close", onClose)
	})
}

// Upgrades socket to the given transport
func (s *socket) MaybeUpgrade(transport transports.Transport) {
	socket_log.Debug(`might upgrade socket transport from "%s" to "%s"`, s.Transport().Name(), transport.Name())

	s.upgrading.Store(true)

	var check, cleanup func()
	var onPacket, onError, onTransportClose, onClose events.Listener
	var upgradeTimeoutTimer, checkIntervalTimer atomic.Pointer[utils.Timer]

	onPacket = func(datas ...any) {
		data := datas[0].(*packet.Packet)
		sb := new(strings.Builder)
		io.Copy(sb, data.Data)
		if packet.PING == data.Type && "probe" == sb.String() {
			socket_log.Debug("got probe ping packet, sending pong")
			transport.Send([]*packet.Packet{{Type: packet.PONG, Data: strings.NewReader("probe")}})
			s.Emit("upgrading", transport)

			utils.ClearInterval(checkIntervalTimer.Load())
			checkIntervalTimer.Store(utils.SetInterval(check, 100*time.Millisecond))

		} else if packet.UPGRADE == data.Type && s.ReadyState() != "closed" {
			socket_log.Debug("got upgrade packet - upgrading")
			cleanup()
			s.Transport().Discard()

			s.upgraded.Store(true)

			s.clearTransport()
			s.setTransport(transport)
			s.Emit("upgrade", transport)
			s.flush()
			if s.ReadyState() == "closing" {
				transport.Close(func() {
					s.OnClose("forced close")
				})
			}
		} else {
			cleanup()
			transport.Close()
		}
	}

	// we force a polling cycle to ensure a fast upgrade
	check = func() {
		if "polling" == s.Transport().Name() && s.Transport().Writable() {
			socket_log.Debug("writing a noop packet to polling for fast upgrade")
			s.Transport().Send([]*packet.Packet{{Type: packet.NOOP}})
		}
	}

	cleanup = func() {
		s.upgrading.Store(false)

		utils.ClearInterval(checkIntervalTimer.Load())
		utils.ClearTimeout(upgradeTimeoutTimer.Load())

		if transport != nil {
			transport.RemoveListener("packet", onPacket)
			transport.RemoveListener("close", onTransportClose)
			transport.RemoveListener("error", onError)
		}
		s.RemoveListener("close", onClose)
	}

	onError = func(err ...any) {
		socket_log.Debug("client did not complete upgrade - %v", err[0])
		cleanup()
		if transport != nil {
			transport.Close()
			transport = nil
		}
	}

	onTransportClose = func(...any) {
		onError("transport closed")
	}

	onClose = func(...any) {
		onError("socket closed")
	}

	// set transport upgrade timer
	upgradeTimeoutTimer.Store(utils.SetTimeout(func() {
		socket_log.Debug("client did not complete upgrade - closing transport")
		cleanup()
		if transport != nil {
			if "open" == transport.ReadyState() {
				transport.Close()
			}
		}
	}, s.server.Opts().UpgradeTimeout()))

	transport.On("packet", onPacket)
	transport.Once("close", onTransportClose)
	transport.Once("error", onError)

	s.Once("close", onClose)
}

// Clears listeners and timers associated with current transport.
func (s *socket) clearTransport() {
	s.cleanupFn.Range(func(cleanup types.Callable, _ int) bool {
		cleanup()
		return true
	})
	s.cleanupFn.Clear()

	// silence further transport errors and prevent uncaught exceptions
	s.Transport().On("error", func(...any) {
		socket_log.Debug("error triggered by discarded transport")
	})

	// ensure transport won't stay open
	s.Transport().Close()

	utils.ClearTimeout(s.pingTimeoutTimer.Load())
}

// Called upon transport considered closed.
// Possible reasons: `ping timeout`, `client error`, `parse error`,
// `transport error`, `server close`, `transport close`
func (s *socket) OnClose(reason string, description ...any) {
	description = append(description, nil)
	if "closed" != s.ReadyState() {
		s.SetReadyState("closed")

		// clear timers
		utils.ClearTimeout(s.pingIntervalTimer.Load())

		utils.ClearTimeout(s.pingTimeoutTimer.Load())

		// clean writeBuffer in defer, so developers can still
		// grab the writeBuffer on 'close' event
		defer func() {
			s.writeBuffer.Clear()
		}()

		s.packetsFn.Clear()

		s.sentCallbackFn.Clear()

		s.clearTransport()
		s.Emit("close", reason, description[0])
	}
}

// Setup and manage send callback
func (s *socket) setupSendCallback() {
	// the message was sent successfully, execute the callback
	onDrain := func(...any) {
		if seqFn, err := s.sentCallbackFn.Shift(); err == nil {
			switch fns := seqFn.(type) {
			case func(transports.Transport):
				socket_log.Debug("executing send callback")
				fns(s.Transport())
			case []func(transports.Transport):
				socket_log.Debug("executing batch send callback")
				for _, fn := range fns {
					fn(s.Transport())
				}
			}
		}
	}

	s.Transport().On("drain", onDrain)

	s.cleanupFn.Push(func() {
		s.Transport().RemoveListener("drain", onDrain)
	})
}

// Sends a message packet.
func (s *socket) Send(
	data io.Reader,
	options *packet.Options,
	callback func(transports.Transport),
) Socket {
	s.sendPacket(packet.MESSAGE, data, options, callback)
	return s
}

// Alias of {@link send}.
func (s *socket) Write(
	data io.Reader,
	options *packet.Options,
	callback func(transports.Transport),
) Socket {
	s.sendPacket(packet.MESSAGE, data, options, callback)
	return s
}

// Sends a packet.
func (s *socket) sendPacket(
	packetType packet.Type,
	data io.Reader,
	options *packet.Options,
	callback func(transports.Transport),
) {

	if "closing" != s.ReadyState() && "closed" != s.ReadyState() {
		socket_log.Debug(`sending packet "%s" (%v)`, packetType, data)

		if options == nil {
			options = &packet.Options{Compress: true}
		}

		packet := &packet.Packet{
			Type:    packetType,
			Data:    data,
			Options: options,
		}

		// exports packetCreate event
		s.Emit("packetCreate", packet)

		s.writeBuffer.Push(packet)

		// add send callback to object, if defined
		if callback != nil {
			s.packetsFn.Push(callback)
		}

		s.flush()
	}
}

// Attempts to flush the packets buffer.
func (s *socket) flush() {
	if "closed" != s.ReadyState() && s.Transport().Writable() {
		if wbuf := s.writeBuffer.AllAndClear(); len(wbuf) > 0 {
			socket_log.Debug("flushing buffer to transport")
			s.Emit("flush", wbuf)
			s.server.Emit("flush", s, wbuf)
			if !s.Transport().SupportsFraming() {
				s.sentCallbackFn.Push(s.packetsFn.AllAndClear())
			} else {
				s.sentCallbackFn.Push((func(inputs []func(transports.Transport)) (outputs []any) {
					for _, fn := range inputs {
						outputs = append(outputs, fn)
					}
					return outputs
				})(s.packetsFn.AllAndClear())...)
			}

			s.Transport().Send(wbuf)
			s.Emit("drain")
			s.server.Emit("drain", s)
		}
	}
}

// Get available upgrades for this socket.
func (s *socket) getAvailableUpgrades() []string {
	availableUpgrades := []string{}
	for _, upg := range s.server.Upgrades(s.Transport().Name()).Keys() {
		if s.server.Opts().Transports().Has(upg) {
			availableUpgrades = append(availableUpgrades, upg)
		}
	}
	return availableUpgrades
}

// Closes the socket and underlying transport.
func (s *socket) Close(discard bool) {
	if "open" != s.ReadyState() {
		return
	}

	s.SetReadyState("closing")

	if length := s.writeBuffer.Len(); length > 0 {
		socket_log.Debug("there are %d remaining packets in the buffer, waiting for the 'drain' event", length)
		s.Once("drain", func(...any) {
			socket_log.Debug("all packets have been sent, closing the transport")
			s.closeTransport(discard)
		})
		return
	}

	socket_log.Debug("the buffer is empty, closing the transport right away (discard? %t)", discard)
	s.closeTransport(discard)
}

// Closes the underlying transport.
func (s *socket) closeTransport(discard bool) {
	socket_log.Debug("closing the transport (discard? %t)", discard)
	if discard {
		s.Transport().Discard()
	}
	s.Transport().Close(func() { s.OnClose("forced close") })
}
