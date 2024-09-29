package transports

import (
	"errors"
	"io"
	"net"
	"sync"

	ws "github.com/gorilla/websocket"
	"github.com/zishang520/engine.io-go-parser/packet"
	_types "github.com/zishang520/engine.io-go-parser/types"
	"github.com/zishang520/engine.io/v2/log"
	"github.com/zishang520/engine.io/v2/types"
)

var ws_log = log.NewLog("engine:ws")

type websocket struct {
	Transport

	socket *types.WebSocketConn
	mu     sync.Mutex
}

// WebSocket transport
func MakeWebSocket() Websocket {
	w := &websocket{Transport: MakeTransport()}

	w.Prototype(w)

	return w
}

func NewWebSocket(ctx *types.HttpContext) Websocket {
	w := MakeWebSocket()

	w.Construct(ctx)

	return w
}

func (w *websocket) Construct(ctx *types.HttpContext) {
	w.Transport.Construct(ctx)

	w.socket = ctx.Websocket

	go w._init()

	w.socket.On("error", func(errs ...any) {
		w.OnError("websocket error", errs[0].(error))
	})
	w.socket.Once("close", func(...any) {
		w.OnClose()
	})
	w.SetWritable(true)
	w.SetPerMessageDeflate(nil)
}

// Transport name
func (w *websocket) Name() string {
	return WEBSOCKET
}

// Advertise upgrade support.
func (w *websocket) HandlesUpgrades() bool {
	return true
}

func (w *websocket) _init() {
	for {
		mt, message, err := w.socket.NextReader()
		if err != nil {
			if ws.IsUnexpectedCloseError(err) || errors.Is(err, net.ErrClosed) {
				w.socket.Emit("close")
			} else {
				w.socket.Emit("error", err)
			}
			return
		}

		switch mt {
		case ws.BinaryMessage:
			read := _types.NewBytesBuffer(nil)
			if _, err := read.ReadFrom(message); err != nil {
				if errors.Is(err, net.ErrClosed) {
					w.socket.Emit("close")
				} else {
					w.socket.Emit("error", err)
				}
			} else {
				w.onMessage(read)
			}
		case ws.TextMessage:
			read := _types.NewStringBuffer(nil)
			if _, err := read.ReadFrom(message); err != nil {
				if errors.Is(err, net.ErrClosed) {
					w.socket.Emit("close")
				} else {
					w.socket.Emit("error", err)
				}
			} else {
				w.onMessage(read)
			}
		case ws.CloseMessage:
			w.socket.Emit("close")
			if c, ok := message.(io.Closer); ok {
				c.Close()
			}
			return
		case ws.PingMessage:
		case ws.PongMessage:
		}
		if c, ok := message.(io.Closer); ok {
			c.Close()
		}
	}
}

func (w *websocket) onMessage(data _types.BufferInterface) {
	ws_log.Debug(`websocket received "%s"`, data)
	w.Transport.OnData(data)
}

// Writes a packet payload.
func (w *websocket) Send(packets []*packet.Packet) {
	w.SetWritable(false)
	defer func() {
		w.Emit("drain")
		w.SetWritable(true)
		w.Emit("ready")
	}()

	w.mu.Lock()
	defer w.mu.Unlock()

	for _, packet := range packets {
		// always creates a new object since ws modifies it
		compress := false
		if packet.Options != nil {
			compress = packet.Options.Compress

			if w.PerMessageDeflate() == nil && packet.Options.WsPreEncodedFrame != nil {
				mt := ws.BinaryMessage
				if _, ok := packet.Options.WsPreEncodedFrame.(*_types.StringBuffer); ok {
					mt = ws.TextMessage
				}
				pm, err := ws.NewPreparedMessage(mt, packet.Options.WsPreEncodedFrame.Bytes())
				if err != nil {
					ws_log.Debug(`Send Error "%s"`, err.Error())
					if errors.Is(err, net.ErrClosed) {
						w.socket.Emit("close")
					} else {
						w.socket.Emit("error", err)
					}
					return
				}
				if err := w.socket.WritePreparedMessage(pm); err != nil {
					ws_log.Debug(`Send Error "%s"`, err.Error())
					if errors.Is(err, net.ErrClosed) {
						w.socket.Emit("close")
					} else {
						w.socket.Emit("error", err)
					}
					return
				}
				return

			}
		}

		data, err := w.Parser().EncodePacket(packet, w.SupportsBinary())
		if err != nil {
			ws_log.Debug(`Send Error "%s"`, err.Error())
			if errors.Is(err, net.ErrClosed) {
				w.socket.Emit("close")
			} else {
				w.socket.Emit("error", err)
			}
			return
		}
		w.write(data, compress)
	}
}

func (w *websocket) write(data _types.BufferInterface, compress bool) {
	if w.PerMessageDeflate() != nil {
		if data.Len() < w.PerMessageDeflate().Threshold {
			compress = false
		}
	}
	ws_log.Debug(`writing %#v`, data)

	w.socket.EnableWriteCompression(compress)
	mt := ws.BinaryMessage
	if _, ok := data.(*_types.StringBuffer); ok {
		mt = ws.TextMessage
	}
	write, err := w.socket.NextWriter(mt)
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			w.socket.Emit("close")
		} else {
			w.socket.Emit("error", err)
		}
		return
	}
	defer func() {
		if err := write.Close(); err != nil {
			if errors.Is(err, net.ErrClosed) {
				w.socket.Emit("close")
			} else {
				w.socket.Emit("error", err)
			}
			return
		}
	}()
	if _, err := io.Copy(write, data); err != nil {
		if errors.Is(err, net.ErrClosed) {
			w.socket.Emit("close")
		} else {
			w.socket.Emit("error", err)
		}
		return
	}
}

// Closes the transport.
func (w *websocket) DoClose(fn types.Callable) {
	ws_log.Debug(`closing`)
	defer w.socket.Close()
	if fn != nil {
		fn()
	}
}
