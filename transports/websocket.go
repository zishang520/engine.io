package transports

import (
	"io"

	ws "github.com/gorilla/websocket"
	"github.com/zishang520/engine.io/log"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/types"
)

var ws_log = log.NewLog("engine:ws")

type websocket struct {
	*transport

	socket *types.WebSocketConn
}

// WebSocket transport
func NewWebSocket(ctx *types.HttpContext) *websocket {
	w := &websocket{}
	return w.New(ctx)
}

func (w *websocket) New(ctx *types.HttpContext) *websocket {
	w.transport = &transport{}

	// Advertise framing support.
	w.supportsFraming = true

	// Advertise upgrade support.
	w.handlesUpgrades = true

	// Transport name
	w.name = "websocket"

	w.transport.New(ctx)

	w.socket = ctx.Websocket
	w.writable = true
	w.perMessageDeflate = nil

	w.doClose = w.WebSocketDoClose
	w.send = w.WebSocketSend

	go func() {
		for {
			mt, message, err := w.socket.NextReader()
			if err != nil {
				if ws.IsUnexpectedCloseError(err) {
					w.OnClose()
				} else {
					w.OnError(err.Error())
				}
				break
			}

			if c, ok := message.(io.Closer); ok {
				defer c.Close()
			}

			switch mt {
			case ws.BinaryMessage:
				read := types.NewBytesBuffer(nil)
				if _, err := read.ReadFrom(message); err != nil {
					w.OnError(err.Error())
					break
				} else {
					w.WebSocketOnData(read)
				}
			case ws.TextMessage:
				read := types.NewStringBuffer(nil)
				if _, err := read.ReadFrom(message); err != nil {
					w.OnError(err.Error())
					break
				} else {
					w.WebSocketOnData(read)
				}
			case ws.CloseMessage:
				w.OnClose()
				break
			case ws.PingMessage:
			case ws.PongMessage:
			}
		}
	}()
	w.socket.On("error", func(errors ...interface{}) {
		w.OnError(errors[0].(error).Error())
	})
	w.socket.On("close", func(...interface{}) {
		w.OnClose()
	})
	return w
}

func (w *websocket) WebSocketOnData(data types.BufferInterface) {
	ws_log.Debug(`websocket received "%s"`, data)
	w.TransportOnData(data)
}

// Writes a packet payload.
func (w *websocket) WebSocketSend(packets []*packet.Packet) {
	onEnd := func(err error) {
		if err != nil {
			w.OnError("write error", err.Error())
			return
		}
		w.writable = true
		w.Emit("drain")
	}

	send := func(packet *packet.Packet) {
		data, err := w.parser.EncodePacket(packet, w.supportsBinary)
		if err != nil {
			ws_log.Debug(`Send Error "%s"`, err)
			return
		}

		ws_log.Debug(`writing "%s"`, data)

		// always creates a new object since ws modifies it
		compress := false
		if packet.Options != nil {
			compress = packet.Options.Compress
		}
		if w.perMessageDeflate != nil {
			if data.Len() < w.perMessageDeflate.Threshold {
				compress = false
			}
		}
		w.writable = false
		w.socket.EnableWriteCompression(compress)
		mt := ws.BinaryMessage
		if _, ok := data.(*types.StringBuffer); ok {
			mt = ws.TextMessage
		}
		write, err := w.socket.NextWriter(mt)
		if err != nil {
			onEnd(err)
			return
		}
		if _, err := io.Copy(write, data); err != nil {
			onEnd(err)
			return
		}
		onEnd(write.Close())
	}

	for _, packet := range packets {
		send(packet)
	}
}

// Closes the transport.
func (w *websocket) WebSocketDoClose(fn ...types.Callable) {
	ws_log.Debug(`closing`)
	w.socket.Close()
	if len(fn) > 0 {
		(fn[0])()
	}
}
