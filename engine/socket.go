package engine

import (
	"encoding/json"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/log"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/transports"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
)

var socket_log = log.NewLog("engine:socket")

type socket struct {
	events.EventEmitter

	protocol      int
	request       *types.HttpContext
	remoteAddress string

	readyState string
	transport  transports.Transport

	id                  string
	server              Server
	upgrading           bool
	upgraded            bool
	writeBuffer         []*packet.Packet
	packetsFn           []func(transports.Transport)
	sentCallbackFn      []any
	cleanupFn           []types.Callable
	checkIntervalTimer  *utils.Timer
	upgradeTimeoutTimer *utils.Timer
	pingTimeoutTimer    *utils.Timer
	pingIntervalTimer   *utils.Timer

	muwriteBuffer    sync.RWMutex
	mupacketsFn      sync.RWMutex
	musentCallbackFn sync.RWMutex
	mucleanupFn      sync.RWMutex
}

func (s *socket) Protocol() int {
	return s.protocol
}

func (s *socket) Upgraded() bool {
	return s.upgraded
}

func (s *socket) Upgrading() bool {
	return s.upgrading
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
	return s.transport
}

func (s *socket) Server() Server {
	return s.server
}

func (s *socket) ReadyState() string {
	return s.readyState
}

func (s *socket) SetReadyState(state string) {
	socket_log.Debug("readyState updated from %s to %s", s.readyState, state)
	s.readyState = state
}

// Client class.
func NewSocket(id string, server Server, transport transports.Transport, ctx *types.HttpContext, protocol int) Socket {
	s := &socket{
		EventEmitter: events.New(),
	}
	return s.New(id, server, transport, ctx, protocol)
}

// Client class.
func (s *socket) New(id string, server Server, transport transports.Transport, ctx *types.HttpContext, protocol int) Socket {
	s.id = id

	s.server = server
	s.upgrading = false
	s.upgraded = false
	s.SetReadyState("opening")

	s.writeBuffer = []*packet.Packet{}
	s.packetsFn = []func(transports.Transport){}
	s.sentCallbackFn = []any{}
	s.cleanupFn = []types.Callable{}
	s.request = ctx
	s.protocol = protocol

	// Cache IP since it might not be in the req later
	if ctx.Websocket != nil && ctx.Websocket.Conn != nil {
		s.remoteAddress = ctx.Websocket.Conn.RemoteAddr().String()
	} else {
		s.remoteAddress = ctx.Request().RemoteAddr
	}

	s.checkIntervalTimer = nil
	s.upgradeTimeoutTimer = nil
	s.pingTimeoutTimer = nil
	s.pingIntervalTimer = nil

	s.setTransport(transport)
	s.onOpen()

	return s
}

// Called upon transport considered open.
func (s *socket) onOpen() {
	s.SetReadyState("open")

	// sends an `open` packet
	s.transport.SetSid(s.id)

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
		types.NewStringBuffer(data),
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
		if s.transport.Protocol() != 3 {
			s.onError("invalid heartbeat direction")
			return
		}
		socket_log.Debug("got ping")
		s.sendPacket(packet.PONG, nil, nil, nil)
		s.Emit("heartbeat")
		break

	case packet.PONG:
		if s.transport.Protocol() == 3 {
			s.onError("invalid heartbeat direction")
			return
		}
		socket_log.Debug("got pong")
		s.pingIntervalTimer.Refresh()
		s.Emit("heartbeat")
		break

	case packet.ERROR:
		s.OnClose("parse error")
		break

	case packet.MESSAGE:
		s.Emit("data", data.Data)
		s.Emit("message", data.Data)
		break
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
	s.pingIntervalTimer = utils.SetTimeOut(func() {
		socket_log.Debug("writing ping packet - expecting pong within %dms", int64(s.server.Opts().PingTimeout()/time.Millisecond))
		s.sendPacket(packet.PING, nil, nil, nil)
		s.resetPingTimeout(s.server.Opts().PingTimeout())
	}, s.server.Opts().PingInterval())
}

// Resets ping timeout.
func (s *socket) resetPingTimeout(timeout time.Duration) {
	utils.ClearTimeout(s.pingTimeoutTimer)
	s.pingTimeoutTimer = utils.SetTimeOut(func() {
		if s.ReadyState() == "closed" {
			return
		}
		s.OnClose("ping timeout")
	}, timeout)
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

	s.transport = transport
	s.transport.Once("error", onError)
	s.transport.On("packet", onPacket)
	s.transport.On("drain", flush)
	s.transport.Once("close", onClose)
	// s function will manage packet events (also message callbacks)
	s.setupSendCallback()

	s.mucleanupFn.Lock()
	s.cleanupFn = append(s.cleanupFn, func() {
		transport.RemoveListener("error", onError)
		transport.RemoveListener("packet", onPacket)
		transport.RemoveListener("drain", flush)
		transport.RemoveListener("close", onClose)
	})
	s.mucleanupFn.Unlock()
}

// Upgrades socket to the given transport
func (s *socket) MaybeUpgrade(transport transports.Transport) {
	socket_log.Debug(`might upgrade socket transport from "%s" to "%s"`, s.transport.Name(), transport.Name())

	s.upgrading = true

	var check, cleanup func()
	var onPacket, onError, onTransportClose, onClose events.Listener

	onPacket = func(datas ...any) {
		data := datas[0].(*packet.Packet)
		sb := new(strings.Builder)
		io.Copy(sb, data.Data)
		if packet.PING == data.Type && "probe" == sb.String() {
			socket_log.Debug("got probe ping packet, sending pong")
			transport.Send([]*packet.Packet{&packet.Packet{Type: packet.PONG, Data: strings.NewReader("probe")}})
			s.Emit("upgrading", transport)
			utils.ClearInterval(s.checkIntervalTimer)
			s.checkIntervalTimer = utils.SetInterval(check, 100*time.Millisecond)
		} else if packet.UPGRADE == data.Type && s.ReadyState() != "closed" {
			socket_log.Debug("got upgrade packet - upgrading")
			cleanup()
			s.transport.Discard()
			s.upgraded = true
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
		if "polling" == s.transport.Name() && s.transport.Writable() {
			socket_log.Debug("writing a noop packet to polling for fast upgrade")
			s.transport.Send([]*packet.Packet{&packet.Packet{Type: packet.NOOP}})
		}
	}

	cleanup = func() {
		s.upgrading = false

		utils.ClearInterval(s.checkIntervalTimer)
		s.checkIntervalTimer = nil

		utils.ClearTimeout(s.upgradeTimeoutTimer)
		s.upgradeTimeoutTimer = nil

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
	s.upgradeTimeoutTimer = utils.SetTimeOut(func() {
		socket_log.Debug("client did not complete upgrade - closing transport")
		cleanup()
		if "open" == transport.ReadyState() {
			transport.Close()
		}
	}, s.server.Opts().UpgradeTimeout())

	transport.On("packet", onPacket)
	transport.Once("close", onTransportClose)
	transport.Once("error", onError)

	s.Once("close", onClose)
}

// Clears listeners and timers associated with current transport.
func (s *socket) clearTransport() {

	s.mucleanupFn.RLock()
	for _, cleanup := range s.cleanupFn {
		cleanup()
	}
	s.mucleanupFn.RUnlock()

	// silence further transport errors and prevent uncaught exceptions
	s.transport.On("error", func(...any) {
		socket_log.Debug("error triggered by discarded transport")
	})

	// ensure transport won't stay open
	s.transport.Close()

	utils.ClearTimeout(s.pingTimeoutTimer)
}

// Called upon transport considered closed.
// Possible reasons: `ping timeout`, `client error`, `parse error`,
// `transport error`, `server close`, `transport close`
func (s *socket) OnClose(reason string, description ...any) {
	description = append(description, nil)
	if "closed" != s.ReadyState() {
		s.SetReadyState("closed")

		// clear timers
		utils.ClearTimeout(s.pingIntervalTimer)
		utils.ClearTimeout(s.pingTimeoutTimer)

		utils.ClearInterval(s.checkIntervalTimer)
		s.checkIntervalTimer = nil
		utils.ClearTimeout(s.upgradeTimeoutTimer)
		// clean writeBuffer in next tick, so developers can still
		// grab the writeBuffer on 'close' event
		defer func() {
			s.muwriteBuffer.Lock()
			s.writeBuffer = []*packet.Packet{}
			s.muwriteBuffer.Unlock()
		}()

		s.mupacketsFn.Lock()
		s.packetsFn = []func(transports.Transport){}
		s.mupacketsFn.Unlock()

		s.musentCallbackFn.Lock()
		s.sentCallbackFn = []any{}
		s.musentCallbackFn.Unlock()

		s.clearTransport()
		s.Emit("close", reason, description[0])
	}
}

// Setup and manage send callback
func (s *socket) setupSendCallback() {
	// the message was sent successfully, execute the callback
	onDrain := func(...any) {
		s.musentCallbackFn.Lock()
		defer s.musentCallbackFn.Unlock()

		if len(s.sentCallbackFn) > 0 {
			seqFn := s.sentCallbackFn[0]
			s.sentCallbackFn = s.sentCallbackFn[1:]

			switch fns := seqFn.(type) {
			case func(transports.Transport):
				socket_log.Debug("executing send callback")
				fns(s.transport)
			case []func(transports.Transport):
				socket_log.Debug("executing batch send callback")
				for _, fn := range fns {
					fn(s.transport)
				}
			}
		}
	}

	s.transport.On("drain", onDrain)

	s.mucleanupFn.Lock()
	s.cleanupFn = append(s.cleanupFn, func() {
		s.transport.RemoveListener("drain", onDrain)
	})
	s.mucleanupFn.Unlock()
}

// Sends a message packet.
func (s *socket) Send(data io.Reader, options *packet.Options, callback func(transports.Transport)) Socket {
	s.sendPacket(packet.MESSAGE, data, options, callback)
	return s
}

func (s *socket) Write(data io.Reader, options *packet.Options, callback func(transports.Transport)) Socket {
	s.sendPacket(packet.MESSAGE, data, options, callback)
	return s
}

// Sends a packet.
func (s *socket) sendPacket(packetType packet.Type, data io.Reader, options *packet.Options, callback func(transports.Transport)) {

	if "closing" != s.ReadyState() && "closed" != s.ReadyState() {
		socket_log.Debug(`sending packet "%s" (%v)`, packetType, data)

		packet := &packet.Packet{
			Type:    packetType,
			Data:    data,
			Options: options,
		}

		// exports packetCreate event
		s.Emit("packetCreate", packet)

		s.muwriteBuffer.Lock()
		s.writeBuffer = append(s.writeBuffer, packet)
		s.muwriteBuffer.Unlock()

		// add send callback to object, if defined
		if callback != nil {
			s.mupacketsFn.Lock()
			s.packetsFn = append(s.packetsFn, callback)
			s.mupacketsFn.Unlock()
		}

		s.flush()
	}
}

// Attempts to flush the packets buffer.
func (s *socket) flush() {
	s.muwriteBuffer.RLock()
	wbuf := append([]*packet.Packet{}, s.writeBuffer...)
	s.muwriteBuffer.RUnlock()

	if "closed" != s.ReadyState() && s.transport.Writable() && len(wbuf) > 0 {
		socket_log.Debug("flushing buffer to transport")
		s.Emit("flush", wbuf)
		s.server.Emit("flush", s, wbuf)

		s.muwriteBuffer.Lock()
		s.writeBuffer = []*packet.Packet{}
		s.muwriteBuffer.Unlock()

		if !s.transport.SupportsFraming() {
			s.musentCallbackFn.Lock()
			s.mupacketsFn.RLock()
			s.sentCallbackFn = append(s.sentCallbackFn, s.packetsFn)
			s.mupacketsFn.RUnlock()
			s.musentCallbackFn.Unlock()

		} else {
			s.musentCallbackFn.Lock()
			s.mupacketsFn.RLock()
			for _, fn := range s.packetsFn {
				s.sentCallbackFn = append(s.sentCallbackFn, fn)
			}
			s.mupacketsFn.RUnlock()
			s.musentCallbackFn.Unlock()
		}
		s.mupacketsFn.Lock()
		s.packetsFn = []func(transports.Transport){}
		s.mupacketsFn.Unlock()

		s.transport.Send(wbuf)
		s.Emit("drain")
		s.server.Emit("drain", s)
	}
}

// Get available upgrades for this socket.
func (s *socket) getAvailableUpgrades() (availableUpgrades []string) {
	for _, upg := range s.server.Upgrades(s.transport.Name()).Keys() {
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

	s.muwriteBuffer.RLock()
	writeBufferLength := len(s.writeBuffer)
	s.muwriteBuffer.RUnlock()

	if writeBufferLength > 0 {
		s.Once("drain", func(...any) {
			s.closeTransport(discard)
		})
		return
	}

	s.closeTransport(discard)
}

// Closes the underlying transport.
func (s *socket) closeTransport(discard bool) {
	if discard {
		s.transport.Discard()
	}
	s.transport.Close(func() { s.OnClose("forced close") })
}
