package engineio

import (
	"encoding/json"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
)

type socket struct {
	events.EventEmmiter

	id                  string
	server              interface{}
	upgrading           bool
	upgraded            bool
	readyState          string
	writeBuffer         types.BytesBuffer
	packetsFn           []interface{}
	sentCallbackFn      []interface{}
	cleanupFn           []interface{}
	request             interface{}
	protocol            int
	remoteAddress       string
	checkIntervalTimer  *utils.Timer
	upgradeTimeoutTimer *utils.Timer
	pingTimeoutTimer    *utils.Timer
	pingIntervalTimer   *utils.Timer

	transport types.Transport
}

/**
 * Client class (abstract).
 *
 * @api private
 */
func NewSocket(id string, server *server, transport types.Transport, ctx *types.HttpContext, protocol int) types.Socket {
	s := &socket{}
	s.id = id
	s.server = server
	s.upgrading = false
	s.upgraded = false
	s.readyState = "opening"
	s.writeBuffer = types.NewBytesBuffer(nil)
	s.packetsFn = []interface{}{}
	s.sentCallbackFn = []interface{}{}
	s.cleanupFn = []interface{}{}
	s.request = req
	s.protocol = protocol

	// Cache IP since it might not be in the req later
	// if req.websocket && req.websocket._socket {
	// 	s.remoteAddress = req.websocket._socket.remoteAddress
	// } else {
	// 	s.remoteAddress = req.connection.remoteAddress
	// }

	s.checkIntervalTimer = nil
	s.upgradeTimeoutTimer = nil
	s.pingTimeoutTimer = nil
	s.pingIntervalTimer = nil

	s.setTransport(transport)
	s.onOpen()

	return s
}

/**
 * Called upon transport considered open.
 *
 * @api private
 */

func (s *socket) onOpen() {
	s.readyState = "open"

	// sends an `open` packet
	s.transport.Sid(s.id)
	s.sendPacket(
		packet.OPEN,
		map[string]interface{}{
			"sid":          s.id,
			"upgrades":     s.getAvailableUpgrades(),
			"pingInterval": s.server.Opts.PingInterval,
			"pingTimeout":  s.server.Opts.PingTimeout,
		},
	)

	if s.server.Opts.InitialPacket != nil {
		s.sendPacket(packet.MESSAGE, s.server.Opts.InitialPacket)
	}

	s.Emit("open")

	if s.protocol == 3 {
		// in protocol v3, the client sends a ping, and the server answers with a pong
		s.resetPingTimeout(s.server.Opts.PingInterval + s.server.Opts.PingTimeout)
	} else {
		// in protocol v4, the server sends a ping, and the client answers with a pong
		s.schedulePing()
	}
}

/**
 * Called upon transport packet.
 *
 * @param {Object} packet
 * @api private
 */

func (s *socket) onPacket(data packet.Packet) {
	if "open" == s.readyState {
		// export packet event
		utils.Log.Debug("packet")
		s.Emit("packet", data)

		// Reset ping timeout on any packet, incoming data is a good sign of
		// other side's liveness
		s.resetPingTimeout(s.server.Opts.PingInterval + s.server.Opts.PingTimeout)

		switch data.Type {
		case "ping":
			if s.transport.protocol != 3 {
				s.onError("invalid heartbeat direction")
				return
			}
			utils.Log.Debug("got ping")
			s.sendPacket(packet.PONG, nil)
			s.Emit("heartbeat")
			break

		case "pong":
			if s.transport.protocol == 3 {
				s.onError("invalid heartbeat direction")
				return
			}
			utils.Log.Debug("got pong")
			s.schedulePing()
			s.Emit("heartbeat")
			break

		case "error":
			s.onClose("parse error")
			break

		case "message":
			s.Emit("data", data.Data)
			s.Emit("message", data.Data)
			break
		}
	} else {
		utils.Log.Debug("packet received with closed socket")
	}
}

func (s *socket) onError(err string) {
	utils.Log.Debug("transport error")
	s.onClose("transport error", err)
}

func (s *socket) schedulePing() {
	if s.pingIntervalTimer != nil {
		utils.ClearTimeOut(s.pingIntervalTimer)
	}
	s.pingIntervalTimer = utils.SetTimeOut(func() {
		utils.Log.Debug("writing ping packet - expecting pong within %sms", s.server.opts.pingTimeout)
		s.sendPacket(packet.PING)
		s.resetPingTimeout(s.server.Opts.PingTimeout)
	}, s.server.Opts.PingInterval)
}

func (s *socket) resetPingTimeout(timeout time.Duration) {
	if s.pingTimeoutTimer != nil {
		utils.ClearTimeOut(s.pingTimeoutTimer)
	}
	s.pingTimeoutTimer = utils.SetTimeOut(func() {
		if s.readyState == "closed" {
			return
		}
		s.onClose("ping timeout")
	}, timeout)
}

func (s *socket) setTransport(transport Transport) {
	onError := s.onError
	onPacket := s.onPacket
	flush := s.flush
	onClose := s.onClose.bind(s, "transport close")

	s.transport = transport
	s.transport.Once("error", onError)
	s.transport.On("packet", onPacket)
	s.transport.On("drain", flush)
	s.transport.Once("close", onClose)
	// s function will manage packet events (also message callbacks)
	s.setupSendCallback()

	s.cleanupFn = append(s.cleanupFn, func() {
		transport.RemoveListener("error", onError)
		transport.RemoveListener("packet", onPacket)
		transport.RemoveListener("drain", flush)
		transport.RemoveListener("close", onClose)
	})
}

func (s *socket) maybeUpgrade(transport) {
	utils.Log.Debug(`might upgrade socket transport from "%s" to "%s"`, s.transport.name, transport.name)

	s.upgrading = true

	cleanup := func() {
		s.upgrading = false

		if s.checkIntervalTimer != nil {
			utils.ClearInterval(s.checkIntervalTimer)
		}

		if s.upgradeTimeoutTimer != nil {
			utils.ClearTimeout(s.upgradeTimeoutTimer)
		}

		transport.RemoveListener("packet", onPacket)
		transport.RemoveListener("close", onTransportClose)
		transport.RemoveListener("error", onError)
		s.RemoveListener("close", onClose)
	}

	onPacket := func(data *packet.Packet) {
		var sb = strings.Builder
		io.Copy(sb, data.Data)
		if "ping" == data.Type && "probe" == sb.String() {
			transport.Send([]*packet.Packet{data})
			s.Emit("upgrading", transport)
			if s.checkIntervalTimer != nil {
				utils.ClearInterval(s.checkIntervalTimer)
			}
			// we force a polling cycle to ensure a fast upgrade
			s.checkIntervalTimer = utils.SetInterval(func() {
				if "polling" == s.transport.Name && s.transport.writable {
					utils.Log.Debug("writing a noop packet to polling for fast upgrade")
					s.transport.send([]*packet.Packet{
						&packet.Packet{
							Type: packet.NOOP,
						},
					})
				}
			}, 100*time.Millisecond)
		} else if "upgrade" == data.Type && s.readyState != "closed" {
			utils.Log.Debug("got upgrade packet - upgrading")
			cleanup()
			s.transport.discard()
			s.upgraded = true
			s.clearTransport()
			s.setTransport(transport)
			s.Emit("upgrade", transport)
			s.flush()
			if s.readyState == "closing" {
				transport.close(func() {
					s.onClose("forced close")
				})
			}
		} else {
			cleanup()
			transport.close()
		}
	}

	onError := func(err) {
		utils.Log.Debug("client did not complete upgrade - %s", err)
		cleanup()
		transport.Close()
		transport = nil
	}

	onTransportClose := func() {
		onError("transport closed")
	}

	onClose := func() {
		onError("socket closed")
	}

	// set transport upgrade timer
	s.upgradeTimeoutTimer = utils.SetTimeOut(func() {
		utils.Log.Debug("client did not complete upgrade - closing transport")
		cleanup()
		if "open" == transport.ReadyState() {
			transport.Close()
		}
	}, s.server.Opts.UpgradeTimeout)

	transport.On("packet", onPacket)
	transport.Once("close", onTransportClose)
	transport.Once("error", onError)

	s.Once("close", onClose)
}

/**
 * Clears listeners and timers associated with current transport.
 *
 * @api private
 */

func (s *socket) clearTransport() {
	for _, cleanup := range s.cleanupFn {
		cleanup()
	}

	// silence further transport errors and prevent uncaught exceptions
	s.transport.on("error", func() {
		utils.Log.Debug("error triggered by discarded transport")
	})

	// ensure transport won't stay open
	s.transport.Close()

	if s.pingTimeoutTimer != nil {
		utils.ClearTimeout(s.pingTimeoutTimer)
	}
}

func (s *socket) onClose(reason string, description string) {
	if "closed" != s.readyState {
		s.readyState = "closed"

		// clear timers
		if s.pingIntervalTimer != nil {
			utils.ClearTimeout(s.pingIntervalTimer)
		}
		if s.pingTimeoutTimer != nil {
			utils.ClearTimeout(s.pingTimeoutTimer)
		}

		if s.checkIntervalTimer != nil {
			utils.ClearInterval(s.checkIntervalTimer)
		}
		if s.upgradeTimeoutTimer != nil {
			utils.ClearTimeout(s.upgradeTimeoutTimer)
		}
		// clean writeBuffer in next tick, so developers can still
		// grab the writeBuffer on 'close' event
		defer func() {
			s.writeBuffer = []interface{}{}
		}()
		s.packetsFn = []interface{}{}
		s.sentCallbackFn = []interface{}{}
		s.clearTransport()
		s.Emit("close", reason, description)
	}
}

/**
 * Setup and manage send callback
 *
 * @api private
 */

func (s *socket) setupSendCallback() {
	// the message was sent successfully, execute the callback
	onDrain := func() {
		if len(s.sentCallbackFn) > 0 {
			seqFn := s.sentCallbackFn[0:1][0]
			switch fn := seqFn.(type) {
			case func():
				utils.Log.Debug("executing send callback")
				fn(s.transport)
			case []func():
				utils.Log.Debug("executing batch send callback")
				for _, _fn := range fn {
					_fn(s.transport)
				}
			}
		}
	}

	s.transport.On("drain", onDrain)

	s.cleanupFn = append(s.cleanupFn, func() {
		s.transport.RemoveListener("drain", onDrain)
	})
}

/**
 * Sends a message packet.
 *
 * @param {String} message
 * @param {Object} options
 * @param {Function} callback
 * @return {Socket} for chaining
 * @api public
 */

func (s *socket) Send(data interface{}, options interface{}, callback interface{}) *socket {
	s.sendPacket(`message`, data, options, callback)
	return s
}

func (s *socket) Write(data interface{}, options interface{}, callback interface{}) *socket {
	return s.Send(`message`, data, options, callback)
}

/**
 * Sends a packet.
 *
 * @param {String} packet type
 * @param {String} optional, data
 * @param {Object} options
 * @api private
 */

func (s *socket) sendPacket(packet_type packet.Type, data io.Reader, options *packet.Option, callback interface{}) {
	if "closing" != s.readyState && "closed" != s.readyState {
		utils.Log.Debug(`sending packet "%s" (%s)`, packet_type, data)

		packet := &packet.Packet{
			Type:    packet_type,
			Options: options,
			Data:    data,
		}

		// exports packetCreate event
		s.Emit("packetCreate", packet)

		s.writeBuffer = append(s.writeBuffer, packet)

		// add send callback to object, if defined
		if callback != nil {
			s.packetsFn = append(s.packetsFn, callback)
		}

		s.flush()
	}
}

/**
 * Attempts to flush the packets buffer.
 *
 * @api private
 */

func (s *socket) flush() {
	// if (
	//   "closed" != s.readyState &&
	//   s.transport.writable &&
	//   s.writeBuffer.length
	// ) {
	//   utils.Log.Debug("flushing buffer to transport");
	//   s.Emit("flush", s.writeBuffer);
	//   s.server.Emit("flush", s, s.writeBuffer);
	//   const wbuf = s.writeBuffer;
	//   s.writeBuffer = [];
	//   if (!s.transport.supportsFraming) {
	//     s.sentCallbackFn.push(s.packetsFn);
	//   } else {
	//     s.sentCallbackFn.push.apply(s.sentCallbackFn, s.packetsFn);
	//   }
	//   s.packetsFn = [];
	//   s.transport.send(wbuf);
	//   s.Emit("drain");
	//   s.server.Emit("drain", s);
	// }
}

/**
 * Get available upgrades for s socket.
 *
 * @api private
 */

func (s *socket) getAvailableUpgrades() {
	// const availableUpgrades = [];
	// const allUpgrades = s.server.upgrades(s.transport.name);
	// let i = 0;
	// const l = allUpgrades.length;
	// for (; i < l; ++i) {
	//   const upg = allUpgrades[i];
	//   if (s.server.opts.transports.indexOf(upg) !== -1) {
	//     availableUpgrades.push(upg);
	//   }
	// }
	// return availableUpgrades;
}

/**
 * Closes the socket and underlying transport.
 *
 * @param {Boolean} optional, discard
 * @return {Socket} for chaining
 * @api public
 */

func (s *socket) close(discard) {
	// if ("open" !== s.readyState) return;

	// s.readyState = "closing";

	// if (s.writeBuffer.length) {
	//   s.once("drain", s.closeTransport.bind(s, discard));
	//   return;
	// }

	// s.closeTransport(discard);
}

/**
 * Closes the underlying transport.
 *
 * @param {Boolean} discard
 * @api private
 */

func (s *socket) closeTransport(discard) {
	if discard {
		s.transport.discard()
	}
	s.transport.close(s.onClose.bind(s, "forced close"))
}
