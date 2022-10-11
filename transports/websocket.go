package transports

import (
	"io"
	"sync"

	ws "github.com/gorilla/websocket"
	"github.com/zishang520/engine.io/log"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/types"
)

var ws_log = log.NewLog("engine:ws")

type websocket struct {
	*transport

	socket *types.WebSocketConn
	mu     sync.Mutex
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
	w.SetWritable(true)
	w.perMessageDeflate = nil

	w.doClose = w.WebSocketDoClose
	w.send = w.WebSocketSend

	go w._init()

	w.socket.On("error", func(errors ...any) {
		w.OnError(errors[0].(error).Error())
	})
	w.socket.On("close", func(...any) {
		w.OnClose()
	})

	return w
}

func (w *websocket) _init() {
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

		switch mt {
		case ws.BinaryMessage:
			read := types.NewBytesBuffer(nil)
			if _, err := read.ReadFrom(message); err != nil {
				w.OnError(err.Error())
			} else {
				w.WebSocketOnData(read)
			}
		case ws.TextMessage:
			read := types.NewStringBuffer(nil)
			if _, err := read.ReadFrom(message); err != nil {
				w.OnError(err.Error())
			} else {
				w.WebSocketOnData(read)
			}
		case ws.CloseMessage:
			w.OnClose()
			break
		case ws.PingMessage:
		case ws.PongMessage:
		}
		if c, ok := message.(io.Closer); ok {
			c.Close()
		}
	}
}

func (w *websocket) WebSocketOnData(data types.BufferInterface) {
	ws_log.Debug(`websocket received "%s"`, data)
	w.TransportOnData(data)
}

// Writes a packet payload.
func (w *websocket) WebSocketSend(packets []*packet.Packet) {
	w.musend.Lock()
	for _, packet := range packets {
		w._Send(packet)
	}
	w.musend.Unlock()

	w.SetWritable(true)
	w.Emit("drain")
}

func (w *websocket) _Send(packet *packet.Packet) {
	var data types.BufferInterface

	if packet.WsPreEncoded != nil {
		data = packet.WsPreEncoded
	} else {
		var err error
		data, err = w.parser.EncodePacket(packet, w.supportsBinary)
		if err != nil {
			ws_log.Debug(`Send Error "%s"`, err)
			return
		}
	}

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
	ws_log.Debug(`writing "%s"`, data)
	w.SetWritable(false)

	w._send(data, compress)
}

func (w *websocket) _send(data types.BufferInterface, compress bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.socket.EnableWriteCompression(compress)
	mt := ws.BinaryMessage
	if _, ok := data.(*types.StringBuffer); ok {
		mt = ws.TextMessage
	}
	write, err := w.socket.NextWriter(mt)
	if err != nil {
		w.OnError("write error", err.Error())
		return
	}
	defer func() {
		if err := write.Close(); err != nil {
			w.OnError("write error", err.Error())
			return
		}
	}()
	if _, err := io.Copy(write, data); err != nil {
		w.OnError("write error", err.Error())
		return
	}
}

// Closes the transport.
func (w *websocket) WebSocketDoClose(fn ...types.Callable) {
	ws_log.Debug(`closing`)
	if len(fn) > 0 {
		(fn[0])()
	}
	w.socket.Close()
}
