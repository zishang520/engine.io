package engineio

import (
	"encoding/json"
	events "github.com/kataras/go-events"
)

type Socket struct {
	id                  string
	server              interface{}
	upgrading           bool
	upgraded            bool
	readyState          string
	writeBuffer         []byte
	packetsFn           []byte
	sentCallbackFn      []byte
	cleanupFn           []byte
	request             interface{}
	remoteAddress       string
	checkIntervalTimer  int
	upgradeTimeoutTimer int
	pingTimeoutTimer    int

	EventEmitter events.EventEmmiter
}

func NewSocket(id string, server interface{}, transport interface{}, req interface{}) *Socket {
	this := &Socket{}

	this.id = id
	this.server = server
	this.upgrading = false
	this.upgraded = false
	this.readyState = `opening`
	this.writeBuffer = []byte{}
	this.packetsFn = []byte{}
	this.sentCallbackFn = []byte{}
	this.cleanupFn = []byte{}
	this.request = req

	// Cache IP since it might not be in the req later
	// if req.websocket && req.websocket._socket {
	// 	this.remoteAddress = req.websocket._socket.remoteAddress
	// } else {
	// 	this.remoteAddress = req.connection.remoteAddress
	// }

	this.checkIntervalTimer = 0
	this.upgradeTimeoutTimer = 0
	this.pingTimeoutTimer = 0

	this.EventEmitter = events.New()

	this.setTransport(transport)
	this.onOpen()

	return this
}

/**
 * Called upon transport considered open.
 *
 * @api private
 */

func (this *Socket) onOpen() {
	this.readyState = `open`

	// sends an `open` packet
	this.transport.sid = this.id

	data, _ := json.Marshal(map[string]interface{}{
		"sid":          this.id,
		"upgrades":     this.getAvailableUpgrades(),
		"pingInterval": this.server.pingInterval,
		"pingTimeout":  this.server.pingTimeout,
	})
	this.sendPacket(`open`, data)

	if this.server.initialPacket != nil {
		this.sendPacket(`message`, this.server.initialPacket)
	}

	this.EventEmitter.Emit(`open`)
	this.setPingTimeout()
}

/**
 * Called upon transport packet.
 *
 * @param {Object} packet
 * @api private
 */

func (this *Socket) onPacket(packet Packet) {
	if `open` == this.readyState {
		// export packet event
		// debug(`packet`)
		this.EventEmitter.Emit(`packet`, packet)

		// Reset ping timeout on any packet, incoming data is a good sign of
		// other side`s liveness
		this.setPingTimeout()

		switch packet.Type {
		case `ping`:
			// debug(`got ping`)
			this.sendPacket(`pong`)
			this.EventEmitter.Emit(`heartbeat`)
			break

		case `error`:
			this.onClose(`parse error`)
			break

		case `message`:
			this.EventEmitter.Emit(`data`, packet.Data)
			this.EventEmitter.Emit(`message`, packet.Data)
			break
		}
	} else {
		// debug(`packet received with closed socket`)
	}
}

/**
 * Called upon transport error.
 *
 * @param {Error} error object
 * @api private
 */

func (this *Socket) onError(err error) {
	// debug(`transport error`);
	this.onClose(`transport error`, err)
}

/**
 * Sets and resets ping timeout timer based on client pings.
 *
 * @api private
 */

func (this *Socket) setPingTimeout() {
	this.pingTimeoutTimer.Stop()
	this.pingTimeoutTimer = time.NewTimer((this.server.pingInterval + this.server.pingTimeout) * time.Microsecond)

	go (func() {
		<-this.pingTimeoutTimer.C
		this.onClose(`ping timeout`)
	})()
}

/**
 * Attaches handlers for the given transport.
 *
 * @param {Transport} transport
 * @api private
 */

func (this *Socket) setTransport(transport interface{}) {
	onError := this.onError
	onPacket := this.onPacket
	flush := this.flush
	onClose := func() { this.onClose(`transport close`) }

	this.transport = transport
	this.transport.EventEmitter.Once(`error`, onError)
	this.transport.EventEmitter.On(`packet`, onPacket)
	this.transport.EventEmitter.On(`drain`, flush)
	this.transport.EventEmitter.Once(`close`, onClose)
	// this function will manage packet events (also message callbacks)
	this.setupSendCallback()

	this.cleanupFn.push(func() {
		transport.EventEmitter.removeListener(`error`, onError)
		transport.EventEmitter.removeListener(`packet`, onPacket)
		transport.EventEmitter.removeListener(`drain`, flush)
		transport.EventEmitter.removeListener(`close`, onClose)
	})
}

/**
 * Upgrades socket to the given transport
 *
 * @param {Transport} transport
 * @api private
 */

func (this *Socket) maybeUpgrade(transport interface{}) {
	// debug(`might upgrade socket transport from "%s" to "%s"` , this.transport.name, transport.name);
	this.upgrading = true

	onError := func(err) {
		// debug(`client did not complete upgrade - %s`, err)
		cleanup()
		transport.close()
		transport = null
	}

	onTransportClose := func() {
		onError(`transport closed`)
	}

	onClose := func() {
		onError(`socket closed`)
	}

	cleanup := func() {
		this.upgrading = false

		clearInterval(this.checkIntervalTimer)
		this.checkIntervalTimer = null

		clearTimeout(this.upgradeTimeoutTimer)
		this.upgradeTimeoutTimer = null

		transport.EventEmitter.removeListener(`packet`, onPacket)
		transport.EventEmitter.removeListener(`close`, onTransportClose)
		transport.EventEmitter.removeListener(`error`, onError)
		this.EventEmitter.removeListener(`close`, onClose)
	}

	onPacket := func(packet Packet) {
		if `ping` == packet.Type && `probe` == packet.Data {
			transport.send([]interface{}{"type": `pong`, "data": `probe`})
			this.EventEmitter.Emit(`upgrading`, transport)
			clearInterval(this.checkIntervalTimer)
			this.checkIntervalTimer = setInterval(check, 100)
		} else if `upgrade` == packet.Type && this.readyState != `closed` {
			// debug(`got upgrade packet - upgrading`)
			cleanup()
			this.transport.discard()
			this.upgraded = true
			this.clearTransport()
			this.setTransport(transport)
			this.EventEmitter.Emit(`upgrade`, transport)
			this.setPingTimeout()
			this.flush()
			if this.readyState == `closing` {
				transport.close(func() {
					this.onClose(`forced close`)
				})
			}
		} else {
			cleanup()
			transport.close()
		}
	}

	// we force a polling cycle to ensure a fast upgrade
	check := func() {
		if `polling` == this.transport.name && this.transport.writable {
			// debug(`writing a noop packet to polling for fast upgrade`);
			this.transport.send([]Packet{Type: `noop`})
		}
	}

	// set transport upgrade timer
	this.upgradeTimeoutTimer = time.NewTimer(this.server.upgradeTimeout * time.Microsecond)

	go (func() {
		<-this.pingTimeoutTimer.C
		// debug(`client did not complete upgrade - closing transport`);
		cleanup()
		if `open` == transport.readyState {
			transport.close()
		}
	})()

	transport.EventEmitter.On(`packet`, onPacket)
	transport.EventEmitter.Once(`close`, onTransportClose)
	transport.EventEmitter.Once(`error`, onError)

	this.EventEmitter.Once(`close`, onClose)
}

/**
 * Clears listeners and timers associated with current transport.
 *
 * @api private
 */

func (this *Socket) clearTransport() {
	for cleanup := range toCleanUp {
		cleanup()
	}
	// silence further transport errors and prevent uncaught exceptions
	this.transport.EventEmitter.On(`error`, func() {
		// debug(`error triggered by discarded transport`)
	})

	// ensure transport won't stay open
	this.transport.close()

	this.pingTimeoutTimer.Stop()
}

/**
 * Called upon transport considered closed.
 * Possible reasons: `ping timeout`, `client error`, `parse error`,
 * `transport error`, `server close`, `transport close`
 */

func (this *Socket) onClose(reason string, description string) {
	if `closed` != this.readyState {
		this.readyState = `closed`
		clearTimeout(this.pingTimeoutTimer)
		clearInterval(this.checkIntervalTimer)
		this.checkIntervalTimer = null
		clearTimeout(this.upgradeTimeoutTimer)
		// clean writeBuffer in next tick, so developers can still
		// grab the writeBuffer on `close` event
		defer (func() {
			this.writeBuffer = []byte{}
		})()
		this.packetsFn = []byte{}
		this.sentCallbackFn = []byte{}
		this.clearTransport()
		this.EventEmitter.Emit(`close`, reason, description)
	}
}

/**
 * Setup and manage send callback
 *
 * @api private
 */

func (this *Socket) setupSendCallback() {
	// the message was sent successfully, execute the callback
	onDrain := func() {
		if len(this.sentCallbackFn) > 0 {
			seqFn := this.sentCallbackFn[0:1][0]
			switch seq := seqFn.(type) {
			case xx:
				// debug(`executing send callback`);
				seq(this.transport)
			case []xx:
				debug(`executing batch send callback`)
				for seqFn_i := range seq {
					if seqFn, ok := seqFn_i.(xx); ok {
						seqFn(self.transport)
					}
				}
			}
		}
	}

	this.transport.on(`drain`, onDrain)

	this.cleanupFn.push(func() {
		this.transport.removeListener(`drain`, onDrain)
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

func (this *Socket) Send(data interface{}, options interface{}, callback interface{}) *Socket {
	this.sendPacket(`message`, data, options, callback)
	return this
}
func (this *Socket) Write(data interface{}, options interface{}, callback interface{}) *Socket {
	return this.Send(`message`, data, options, callback)
}

/**
 * Sends a packet.
 *
 * @param {String} packet type
 * @param {String} optional, data
 * @param {Object} options
 * @api private
 */

func (this *Socket) sendPacket(packet_type string, data string, options interface{}, callback func()) {
	// if (`function` === typeof options) {
	//   callback = options;
	//   options = null;
	// }

	// options = options || {};
	// options.compress = (false != options.compress);

	if `closing` != this.readyState && `closed` != this.readyState {
		// debug(`sending packet "%s" (%s)`, packet_type, data)

		packet := Packet{
			Type:    packet_type,
			Options: options,
		}
		if data {
			packet.Data = data
		}

		// exports packetCreate event
		this.EventEmitter.Emit(`packetCreate`, packet)

		this.writeBuffer.push(packet)

		// add send callback to object, if defined
		if callback {
			this.packetsFn.push(callback)
		}

		this.flush()
	}
}

/**
 * Attempts to flush the packets buffer.
 *
 * @api private
 */

func (this *Socket) flush() {
	if `closed` != this.readyState && this.transport.writable && this.writeBuffer.length {
		debug(`flushing buffer to transport`)
		this.EventEmitter.Emit(`flush`, this.writeBuffer)
		this.server.EventEmitter.Emit(`flush`, this, this.writeBuffer)
		var wbuf = this.writeBuffer
		this.writeBuffer = []byte{}
		if !this.transport.supportsFraming {
			this.sentCallbackFn.push(this.packetsFn)
		} else {
			this.sentCallbackFn.push.apply(this.sentCallbackFn, this.packetsFn)
		}
		this.packetsFn = []byte{}
		this.transport.send(wbuf)
		this.EventEmitter.Emit(`drain`)
		this.server.EventEmitter.Emit(`drain`, this)
	}
}

/**
 * Get available upgrades for this socket.
 *
 * @api private
 */

func (this *Socket) getAvailableUpgrades() {
	availableUpgrades := []byte{}
	allUpgrades = this.server.upgrades(this.transport.name)

	for upg := range allUpgrades {
		if this.server.transports.indexOf(upg) != -1 {
			availableUpgrades.push(upg)
		}
	}
	return availableUpgrades
}

/**
 * Closes the socket and underlying transport.
 *
 * @param {Boolean} optional, discard
 * @return {Socket} for chaining
 * @api public
 */

func (this *Socket) Close(discard bool) {
	if `open` != this.readyState {
		return
	}

	this.readyState = `closing`

	if this.writeBuffer.length {
		this.EventEmitter.Once(`drain`, func() {
			this.closeTransport(discard)
		})
		return
	}

	this.closeTransport(discard)
}

/**
 * Closes the underlying transport.
 *
 * @param {Boolean} discard
 * @api private
 */

func (this *Socket) closeTransport(discard bool) {
	if discard {
		this.transport.discard()
	}
	this.transport.close(func() {
		this.onClose(`forced close`)
	})
}
