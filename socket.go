package engineio

import (
	"encoding/json"
	"github.com/zishang520/engine.io/types"
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
	checkIntervalTimer  int
	upgradeTimeoutTimer int
	pingTimeoutTimer    int
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

	s.checkIntervalTimer = 0
	s.upgradeTimeoutTimer = 0
	s.pingTimeoutTimer = 0
	s.pingIntervalTimer = 0

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
	s.transport.sid = s.id
	// s.sendPacket(
	//   "open",
	//   JSON.stringify({
	//     sid: s.id,
	//     upgrades: s.getAvailableUpgrades(),
	//     pingInterval: s.server.opts.pingInterval,
	//     pingTimeout: s.server.opts.pingTimeout
	//   })
	// );

	if s.server.opts.initialPacket {
		s.sendPacket("message", s.server.opts.initialPacket)
	}

	s.emit("open")

	if s.protocol == 3 {
		// in protocol v3, the client sends a ping, and the server answers with a pong
		// s.resetPingTimeout(
		//   s.server.opts.pingInterval + s.server.opts.pingTimeout
		// );
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

func (s *socket) onPacket(packet) {
	if "open" == s.readyState {
		// export packet event
		debug("packet")
		s.emit("packet", packet)

		// Reset ping timeout on any packet, incoming data is a good sign of
		// other side's liveness
		// s.resetPingTimeout(
		//   s.server.opts.pingInterval + s.server.opts.pingTimeout
		// );

		switch packet.Type {
		case "ping":
			if s.transport.protocol != 3 {
				s.onError("invalid heartbeat direction")
				return
			}
			debug("got ping")
			s.sendPacket("pong")
			s.emit("heartbeat")
			break

		case "pong":
			if s.transport.protocol == 3 {
				s.onError("invalid heartbeat direction")
				return
			}
			debug("got pong")
			s.schedulePing()
			s.emit("heartbeat")
			break

		case "error":
			s.onClose("parse error")
			break

		case "message":
			s.emit("data", packet.data)
			s.emit("message", packet.data)
			break
		}
	} else {
		debug("packet received with closed socket")
	}
}

func (s *socket) onError(err) {
	debug("transport error")
	s.onClose("transport error", err)
}

func (s *socket) schedulePing() {
	clearTimeout(s.pingIntervalTimer)
	// s.pingIntervalTimer = setTimeout(() => {
	//   debug(
	//     "writing ping packet - expecting pong within %sms",
	//     s.server.opts.pingTimeout
	//   );
	//   s.sendPacket("ping");
	//   s.resetPingTimeout(s.server.opts.pingTimeout);
	// }, s.server.opts.pingInterval);
}

func (s *socket) resetPingTimeout(timeout) {
	clearTimeout(s.pingTimeoutTimer)
	// s.pingTimeoutTimer = setTimeout(() => {
	//   if (s.readyState == "closed") {
	//   	return;
	//   }
	//   s.onClose("ping timeout");
	// }, timeout);
}

func (s *socket) setTransport(transport) {
	const onError = s.onError.bind(s)
	const onPacket = s.onPacket.bind(s)
	const flush = s.flush.bind(s)
	const onClose = s.onClose.bind(s, "transport close")

	s.transport = transport
	s.transport.once("error", onError)
	s.transport.on("packet", onPacket)
	s.transport.on("drain", flush)
	s.transport.once("close", onClose)
	// s function will manage packet events (also message callbacks)
	s.setupSendCallback()

	// s.cleanupFn.push(function() {
	//   transport.removeListener("error", onError);
	//   transport.removeListener("packet", onPacket);
	//   transport.removeListener("drain", flush);
	//   transport.removeListener("close", onClose);
	// });
}

func (s *socket) maybeUpgrade(transport) {
	debug(`might upgrade socket transport from "%s" to "%s"`, s.transport.name, transport.name)

	s.upgrading = true

	// set transport upgrade timer
	// self.upgradeTimeoutTimer = setTimeout(function() {
	//   debug("client did not complete upgrade - closing transport");
	//   cleanup();
	//   if ("open" == transport.readyState) {
	//     transport.close();
	//   }
	// }, s.server.opts.upgradeTimeout);

	// function onPacket(packet) {
	//   if ("ping" == packet.type && "probe" == packet.data) {
	//     transport.send([{ type: "pong", data: "probe" }]);
	//     self.emit("upgrading", transport);
	//     clearInterval(self.checkIntervalTimer);
	//     self.checkIntervalTimer = setInterval(check, 100);
	//   } else if ("upgrade" == packet.type && self.readyState != "closed") {
	//     debug("got upgrade packet - upgrading");
	//     cleanup();
	//     self.transport.discard();
	//     self.upgraded = true;
	//     self.clearTransport();
	//     self.setTransport(transport);
	//     self.emit("upgrade", transport);
	//     self.flush();
	//     if (self.readyState == "closing") {
	//       transport.close(function() {
	//         self.onClose("forced close");
	//       });
	//     }
	//   } else {
	//     cleanup();
	//     transport.close();
	//   }
	// }

	// // we force a polling cycle to ensure a fast upgrade
	// function check() {
	//   if ("polling" == self.transport.name && self.transport.writable) {
	//     debug("writing a noop packet to polling for fast upgrade");
	//     self.transport.send([{ type: "noop" }]);
	//   }
	// }

	// function cleanup() {
	//   self.upgrading = false;

	//   clearInterval(self.checkIntervalTimer);
	//   self.checkIntervalTimer = null;

	//   clearTimeout(self.upgradeTimeoutTimer);
	//   self.upgradeTimeoutTimer = null;

	//   transport.removeListener("packet", onPacket);
	//   transport.removeListener("close", onTransportClose);
	//   transport.removeListener("error", onError);
	//   self.removeListener("close", onClose);
	// }

	// function onError(err) {
	//   debug("client did not complete upgrade - %s", err);
	//   cleanup();
	//   transport.close();
	//   transport = null;
	// }

	// function onTransportClose() {
	//   onError("transport closed");
	// }

	// function onClose() {
	//   onError("socket closed");
	// }

	transport.on("packet", onPacket)
	transport.once("close", onTransportClose)
	transport.once("error", onError)

	self.once("close", onClose)
}

/**
 * Clears listeners and timers associated with current transport.
 *
 * @api private
 */

func (s *socket) clearTransport() {
	// let cleanup;

	const toCleanUp = s.cleanupFn.length

	// for (let i = 0; i < toCleanUp; i++) {
	//   cleanup = s.cleanupFn.shift();
	//   cleanup();
	// }

	// silence further transport errors and prevent uncaught exceptions
	/* s.transport.on("error", function() {
	   debug("error triggered by discarded transport");
	 });*/

	// ensure transport won't stay open
	s.transport.close()

	clearTimeout(s.pingTimeoutTimer)
}

func (s *socket) onClose(reason, description) {
	if "closed" != s.readyState {
		s.readyState = "closed"

		// clear timers
		clearTimeout(s.pingIntervalTimer)
		clearTimeout(s.pingTimeoutTimer)

		clearInterval(s.checkIntervalTimer)
		s.checkIntervalTimer = null
		clearTimeout(s.upgradeTimeoutTimer)
		// const self = s;
		// clean writeBuffer in next tick, so developers can still
		// grab the writeBuffer on 'close' event
		// process.nextTick(function() {
		//   self.writeBuffer = [];
		// });
		s.packetsFn = []interface{}{}
		s.sentCallbackFn = []interface{}{}
		s.clearTransport()
		s.emit("close", reason, description)
	}
}

/**
 * Setup and manage send callback
 *
 * @api private
 */

func (s *socket) setupSendCallback() {
	const self = s
	s.transport.on("drain", onDrain)

	// s.cleanupFn.push(function() {
	//   self.transport.removeListener("drain", onDrain);
	// });

	// the message was sent successfully, execute the callback
	// function onDrain() {
	//   if (self.sentCallbackFn.length > 0) {
	//     const seqFn = self.sentCallbackFn.splice(0, 1)[0];
	//     if ("function" === typeof seqFn) {
	//       debug("executing send callback");
	//       seqFn(self.transport);
	//     } else if (Array.isArray(seqFn)) {
	//       debug("executing batch send callback");
	//       const l = seqFn.length;
	//       let i = 0;
	//       for (; i < l; i++) {
	//         if ("function" === typeof seqFn[i]) {
	//           seqFn[i](self.transport);
	//         }
	//       }
	//     }
	//   }
	// }
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

func (s *socket) sendPacket(packet_type string, data io.Reader, options interface{}, callback interface{}) {
	// if ("function" == typeof options) {
	//   callback = options;
	//   options = null;
	// }

	// options = options || {};
	// options.compress = false !== options.compress;

	// if ("closing" !== s.readyState && "closed" !== s.readyState) {
	//   debug('sending packet "%s" (%s)', type, data);

	//   const packet = {
	//     type: type,
	//     options: options
	//   };
	//   if (data) packet.data = data;

	//   // exports packetCreate event
	//   s.emit("packetCreate", packet);

	//   s.writeBuffer.push(packet);

	//   // add send callback to object, if defined
	//   if (callback) s.packetsFn.push(callback);

	//   s.flush();
	// }
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
	//   debug("flushing buffer to transport");
	//   s.emit("flush", s.writeBuffer);
	//   s.server.emit("flush", s, s.writeBuffer);
	//   const wbuf = s.writeBuffer;
	//   s.writeBuffer = [];
	//   if (!s.transport.supportsFraming) {
	//     s.sentCallbackFn.push(s.packetsFn);
	//   } else {
	//     s.sentCallbackFn.push.apply(s.sentCallbackFn, s.packetsFn);
	//   }
	//   s.packetsFn = [];
	//   s.transport.send(wbuf);
	//   s.emit("drain");
	//   s.server.emit("drain", s);
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
