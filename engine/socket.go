package engine

import (
	"encoding/json"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/transports"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"io"
	"strings"
	"sync"
	"time"
)

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
	sentCallbackFn      []interface{}
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

func (s *socket) Upgraded() bool {
	return s.upgraded
}

func (s *socket) Upgrading() bool {
	return s.upgrading
}

func (s *socket) Id() string {
	return s.id
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
	utils.Log().Debug("readyState updated from %s to %s", s.readyState, state)
	s.readyState = state
}

func NewSocket(id string, server Server, transport transports.Transport, ctx *types.HttpContext, protocol int) Socket {
	s := &socket{
		EventEmitter: events.New(),
	}
	return s.New(id, server, transport, ctx, protocol)
}

func (s *socket) New(id string, server Server, transport transports.Transport, ctx *types.HttpContext, protocol int) Socket {
	s.id = id

	s.server = server
	s.upgrading = false
	s.upgraded = false
	s.SetReadyState("opening")

	s.writeBuffer = []*packet.Packet{}
	s.packetsFn = []func(transports.Transport){}
	s.sentCallbackFn = []interface{}{}
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
func (s *socket) onOpen() {
	s.SetReadyState("open")

	// sends an `open` packet
	s.transport.SetSid(s.id)

	data, err := json.Marshal(map[string]interface{}{
		"sid":          s.id,
		"upgrades":     s.getAvailableUpgrades(),
		"pingInterval": int64(s.server.Opts().PingInterval() / time.Millisecond),
		"pingTimeout":  int64(s.server.Opts().PingTimeout() / time.Millisecond),
		"maxPayload":   s.server.Opts().MaxHttpBufferSize(),
	})

	if err != nil {
		utils.Log().Debug("json.Marshal err")
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
		utils.Log().Debug("packet received with closed socket")
		return
	}

	// export packet event
	utils.Log().Debug(`received packet %s`, data.Type)
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
		utils.Log().Debug("got ping")
		s.sendPacket(packet.PONG, nil, nil, nil)
		s.Emit("heartbeat")
		break

	case packet.PONG:
		if s.transport.Protocol() == 3 {
			s.onError("invalid heartbeat direction")
			return
		}
		utils.Log().Debug("got pong")
		s.schedulePing()
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

func (s *socket) onError(err interface{}) {
	utils.Log().Debug("transport error %v", err)
	s.OnClose("transport error", err)
}

func (s *socket) schedulePing() {
	utils.ClearTimeout(s.pingIntervalTimer)
	s.pingIntervalTimer = utils.SetTimeOut(func() {
		utils.Log().Debug("writing ping packet - expecting pong within %dms", int64(s.server.Opts().PingTimeout()/time.Millisecond))
		s.sendPacket(packet.PING, nil, nil, nil)
		s.resetPingTimeout(s.server.Opts().PingTimeout())
	}, s.server.Opts().PingInterval())
}

func (s *socket) resetPingTimeout(timeout time.Duration) {
	utils.ClearTimeout(s.pingTimeoutTimer)
	s.pingTimeoutTimer = utils.SetTimeOut(func() {
		if s.ReadyState() == "closed" {
			return
		}
		s.OnClose("ping timeout")
	}, timeout)
}

func (s *socket) setTransport(transport transports.Transport) {
	onError := func(err ...interface{}) {
		err = append(err, nil)
		s.onError(err[0])
	}
	onPacket := func(packets ...interface{}) {
		if len(packets) > 0 {
			s.onPacket(packets[0].(*packet.Packet))
		}
	}
	flush := func(...interface{}) { s.flush() }
	onClose := func(...interface{}) { s.OnClose("transport close") }

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

func (s *socket) MaybeUpgrade(transport transports.Transport) {
	utils.Log().Debug(`might upgrade socket transport from "%s" to "%s"`, s.transport.Name(), transport.Name())

	s.upgrading = true

	var check, cleanup func()
	var onPacket, onError, onTransportClose, onClose events.Listener

	onPacket = func(datas ...interface{}) {
		data := datas[0].(*packet.Packet)
		sb := new(strings.Builder)
		io.Copy(sb, data.Data)
		if packet.PING == data.Type && "probe" == sb.String() {
			utils.Log().Debug("got probe ping packet, sending pong")
			transport.Send([]*packet.Packet{&packet.Packet{Type: packet.PONG, Data: strings.NewReader("probe")}})
			s.Emit("upgrading", transport)
			utils.ClearInterval(s.checkIntervalTimer)
			s.checkIntervalTimer = utils.SetInterval(check, 100*time.Millisecond)
		} else if packet.UPGRADE == data.Type && s.ReadyState() != "closed" {
			utils.Log().Debug("got upgrade packet - upgrading")
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
			utils.Log().Debug("writing a noop packet to polling for fast upgrade")
			s.transport.Send([]*packet.Packet{&packet.Packet{Type: packet.NOOP}})
		}
	}

	cleanup = func() {
		s.upgrading = false

		utils.ClearInterval(s.checkIntervalTimer)
		s.checkIntervalTimer = nil

		utils.ClearTimeout(s.upgradeTimeoutTimer)
		s.upgradeTimeoutTimer = nil

		transport.RemoveListener("packet", onPacket)
		transport.RemoveListener("close", onTransportClose)
		transport.RemoveListener("error", onError)
		s.RemoveListener("close", onClose)
	}

	onError = func(err ...interface{}) {
		utils.Log().Debug("client did not complete upgrade - %v", err[0])
		cleanup()
		transport.Close()
		transport = nil
	}

	onTransportClose = func(...interface{}) {
		onError("transport closed")
	}

	onClose = func(...interface{}) {
		onError("socket closed")
	}

	// set transport upgrade timer
	s.upgradeTimeoutTimer = utils.SetTimeOut(func() {
		utils.Log().Debug("client did not complete upgrade - closing transport")
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

func (s *socket) clearTransport() {

	s.mucleanupFn.RLock()
	for _, cleanup := range s.cleanupFn {
		cleanup()
	}
	s.mucleanupFn.RUnlock()

	// silence further transport errors and prevent uncaught exceptions
	s.transport.On("error", func(...interface{}) {
		utils.Log().Debug("error triggered by discarded transport")
	})

	// ensure transport won't stay open
	s.transport.Close()

	utils.ClearTimeout(s.pingTimeoutTimer)
}

func (s *socket) OnClose(reason string, description ...interface{}) {
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
		s.sentCallbackFn = []interface{}{}
		s.musentCallbackFn.Unlock()

		s.clearTransport()
		s.Emit("close", reason, description[0])
	}
}

func (s *socket) setupSendCallback() {
	// the message was sent successfully, execute the callback
	onDrain := func(...interface{}) {
		s.musentCallbackFn.Lock()
		defer s.musentCallbackFn.Unlock()

		if len(s.sentCallbackFn) > 0 {
			seqFn := s.sentCallbackFn[0]
			s.sentCallbackFn = s.sentCallbackFn[1:]

			switch fns := seqFn.(type) {
			case func(transports.Transport):
				utils.Log().Debug("executing send callback")
				fns(s.transport)
			case []func(transports.Transport):
				utils.Log().Debug("executing batch send callback")
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

func (s *socket) Send(data io.Reader, options *packet.Options, callback func(transports.Transport)) Socket {
	s.sendPacket(packet.MESSAGE, data, options, callback)
	return s
}

func (s *socket) Write(data io.Reader, options *packet.Options, callback func(transports.Transport)) Socket {
	return s.Send(data, options, callback)
}

func (s *socket) sendPacket(packetType packet.Type, data io.Reader, options *packet.Options, callback func(transports.Transport)) {

	if "closing" != s.ReadyState() && "closed" != s.ReadyState() {
		utils.Log().Debug(`sending packet "%s" (%v)`, packetType, data)

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

func (s *socket) flush() {
	s.muwriteBuffer.RLock()
	wbuf := s.writeBuffer
	s.muwriteBuffer.RUnlock()

	if "closed" != s.ReadyState() && s.transport.Writable() && len(wbuf) > 0 {
		utils.Log().Debug("flushing buffer to transport")
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

func (s *socket) getAvailableUpgrades() (availableUpgrades []string) {
	for upg := range s.server.Upgrades(s.transport.Name()).All() {
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
	l := len(s.writeBuffer)
	s.muwriteBuffer.RUnlock()

	if l > 0 {
		s.Once("drain", func(...interface{}) {
			s.closeTransport(discard)
		})
		return
	}

	s.closeTransport(discard)
}

func (s *socket) closeTransport(discard bool) {
	if discard {
		s.transport.Discard()
	}
	s.transport.Close(func() { s.OnClose("forced close") })
}