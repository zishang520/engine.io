package transports

import (
	"github.com/zishang520/engine.io-go-parser/packet"
	"github.com/zishang520/engine.io-go-parser/parser"
	"github.com/zishang520/engine.io/v2/events"
	"github.com/zishang520/engine.io/v2/types"
)

type (
	Transport interface {
		// #extends

		events.EventEmitter

		// #prototype

		Prototype(Transport)
		Proto() Transport

		// #setters

		SetSid(string)
		SetWritable(bool)

		SetSupportsBinary(bool)
		SetReadyState(string)
		SetHttpCompression(*types.HttpCompression)
		SetPerMessageDeflate(*types.PerMessageDeflate)
		SetMaxHttpBufferSize(int64)

		// #getters

		// The session ID.
		Sid() string
		// Whether the transport is currently ready to send packets.
		Writable() bool
		// The revision of the protocol:
		//
		// - 3 is used in Engine.IO v3 / Socket.IO v2
		// - 4 is used in Engine.IO v4 and above / Socket.IO v3 and above
		//
		// It is found in the `EIO` query parameters of the HTTP requests.
		//
		// @see https://github.com/socketio/engine.io-protocol
		Protocol() int
		// Whether the transport is discarded and can be safely closed (used during upgrade).
		//
		// @protected
		Discarded() bool
		// The parser to use (depends on the revision of the {@link Transport#protocol}.
		//
		// @protected
		Parser() parser.Parser
		// Whether the transport supports binary payloads (else it will be base64-encoded)
		//
		// @protected
		SupportsBinary() bool
		// The current state of the transport.
		//
		// @protected
		ReadyState() string
		HttpCompression() *types.HttpCompression
		PerMessageDeflate() *types.PerMessageDeflate
		MaxHttpBufferSize() int64
		// @abstract
		HandlesUpgrades() bool
		// @abstract
		Name() string

		// #methods

		// Construct() should be called after calling Prototype()
		Construct(*types.HttpContext)
		// @private
		// Flags the transport as discarded.
		Discard()
		// @protected
		// Called with an incoming HTTP request.
		OnRequest(*types.HttpContext)
		// @private
		// Closes the transport.
		Close(...types.Callable)
		// @protected
		// Called with a transport error.
		OnError(string, error)
		// @protected
		// Called with parsed out a packets from the data stream.
		OnPacket(*packet.Packet)
		// @protected
		// Called with the encoded packet data.
		OnData(types.BufferInterface)
		// @protected
		// Called upon transport close.
		OnClose()
		// @protected
		// @abstract
		// Writes a packet payload.
		Send([]*packet.Packet)
		// @protected
		// @abstract
		// Closes the transport.
		DoClose(types.Callable)
	}

	Polling interface {
		// #extends

		Transport

		// #methods
		SetReq(*types.HttpContext)

		// @protected
		Req() *types.HttpContext

		DoWrite(*types.HttpContext, types.BufferInterface, *packet.Options, func(*types.HttpContext))
	}

	Jsonp interface {
		// #extends

		Polling
	}

	Websocket interface {
		// #extends

		Transport
	}

	WebTransport interface {
		// #extends

		Transport
	}
)
