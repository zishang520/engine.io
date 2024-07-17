package transports

import (
	"io"
	"sync"

	"github.com/zishang520/engine.io-go-parser/packet"
	_types "github.com/zishang520/engine.io-go-parser/types"
	"github.com/zishang520/engine.io/v2/log"
	"github.com/zishang520/engine.io/v2/types"
	"github.com/zishang520/engine.io/v2/webtransport"
)

var (
	wt_log = log.NewLog("engine:webtransport")

	BINARY_HEADER = []byte{0x36}
)

type webTransport struct {
	Transport

	session *types.WebTransportConn
	mu      sync.Mutex
}

// WebTransport transport
func MakeWebTransport() WebTransport {
	w := &webTransport{Transport: MakeTransport()}

	w.Prototype(w)

	return w
}

func NewWebTransport(ctx *types.HttpContext) WebTransport {
	w := MakeWebTransport()

	w.Construct(ctx)

	return w
}

func (w *webTransport) Construct(ctx *types.HttpContext) {
	w.Transport.Construct(ctx)

	w.session = ctx.WebTransport
	go w._init()

	w.SetWritable(true)
	w.SetPerMessageDeflate(nil)
}

// Transport name
func (w *webTransport) Name() string {
	return "webtransport"
}

// Advertise upgrade support.
func (w *webTransport) HandlesUpgrades() bool {
	return true
}

func (w *webTransport) _init() {
	for {
		mt, message, err := w.session.NextReader()
		if err != nil {
			if webtransport.IsUnexpectedCloseError(err) {
				w.OnClose()
			} else {
				w.OnError("Error reading data", err)
			}
			return
		}

		switch mt {
		case webtransport.BinaryMessage:
			read := _types.NewBytesBuffer(nil)
			if _, err := read.ReadFrom(message); err != nil {
				w.OnError("Error reading data", err)
			} else {
				w.onMessage(read)
			}
		case webtransport.TextMessage:
			read := _types.NewStringBuffer(nil)
			if _, err := read.ReadFrom(message); err != nil {
				w.OnError("Error reading data", err)
			} else {
				w.onMessage(read)
			}
		}
		if c, ok := message.(io.Closer); ok {
			c.Close()
		}
	}
}

func (w *webTransport) onMessage(data _types.BufferInterface) {
	wt_log.Debug(`webTransport received "%s"`, data)
	w.Transport.OnData(data)
}

// Writes a packet payload.
func (w *webTransport) Send(packets []*packet.Packet) {
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
				mt := webtransport.BinaryMessage
				if _, ok := packet.Options.WsPreEncodedFrame.(*_types.StringBuffer); ok {
					mt = webtransport.TextMessage
				}
				pm, err := webtransport.NewPreparedMessage(mt, packet.Options.WsPreEncodedFrame.Bytes())
				if err != nil {
					wt_log.Debug(`Send Error "%s"`, err.Error())
					w.OnError("write error", err)
					return
				}
				if err := w.session.WritePreparedMessage(pm); err != nil {
					wt_log.Debug(`Send Error "%s"`, err.Error())
					w.OnError("write error", err)
					return
				}
				return

			}
		}

		data, err := w.Parser().EncodePacket(packet, w.SupportsBinary())
		if err != nil {
			wt_log.Debug(`Send Error "%s"`, err.Error())
			w.OnError("write error", err)
			return
		}
		w.write(data, compress)
	}
}

func (w *webTransport) write(data _types.BufferInterface, compress bool) {
	if w.PerMessageDeflate() != nil {
		if data.Len() < w.PerMessageDeflate().Threshold {
			compress = false
		}
	}
	wt_log.Debug(`writing %#s`, data)

	// w.session.EnableWriteCompression(compress)
	mt := webtransport.BinaryMessage
	if _, ok := data.(*_types.StringBuffer); ok {
		mt = webtransport.TextMessage
	}
	write, err := w.session.NextWriter(mt)
	if err != nil {
		w.OnError("write error", err)
		return
	}
	defer func() {
		if err := write.Close(); err != nil {
			w.OnError("write error", err)
			return
		}
	}()
	if _, err := io.Copy(write, data); err != nil {
		w.OnError("write error", err)
		return
	}
}

// Closes the transport.
func (w *webTransport) DoClose(fn types.Callable) {
	wt_log.Debug(`closing WebTransport session`)
	w.session.CloseWithError(0, "")
	if fn != nil {
		fn()
	}
}
