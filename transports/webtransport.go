package transports

import (
	"io"
	"sync"

	"github.com/zishang520/engine.io-go-parser/packet"
	_types "github.com/zishang520/engine.io-go-parser/types"
	"github.com/zishang520/engine.io/log"
	"github.com/zishang520/engine.io/types"
)

var (
	wt_log = log.NewLog("engine:webtransport")

	BINARY_HEADER = []byte{0x36}
)

type webTransport struct {
	Transport

	session *types.WebTransportConn
	musend  sync.Mutex
}

// WebTransport transport
func MakeWebTransport() WebTransport {
	w := &webTransport{Transport: MakeTransport()}

	w.Prototype(w)

	return w
}

func NewWebTransport(ctx *types.HttpContext) WebTransport {
	w := MakeWebSocket()

	w.Construct(ctx)

	return w
}

func (w *webTransport) Construct(ctx *types.HttpContext) {
	w.Transport.Construct(ctx)

	w.session = ctx.WebTransport
	go w._init(ctx)

	w.session.On("error", func(errors ...any) {
		w.OnError("webTransport error", errors[0].(error))
	})
	w.session.On("close", func(...any) {
		w.OnClose()
	})
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

// Advertise framing support.
func (w *webTransport) SupportsFraming() bool {
	return true
}

func (w *webTransport) _init(ctx *types.HttpContext) {
	defer w.session.Stream.Close()

	binaryFlag := false
LOOP:
	for {
		var data _types.BufferInterface
		if binaryFlag {
			data = _types.NewBytesBuffer(nil)
		} else {
			data = _types.NewStringBuffer(nil)
		}
		// Read data from the stream
		buf := make([]byte, 1024)
		for {
			n, err := w.session.Stream.Read(buf)
			if err != nil {
				if err == io.EOF {
					w.OnClose()
					wt_log.Debug(`session is closed`)
				} else {
					w.OnError("Error reading data", err)
				}
				return
			}
			wt_log.Debug("received chunk: %v", buf[:n])

			if !binaryFlag && n == 1 && buf[0] == byte(0x36) {
				binaryFlag = true
				data.Reset() // clean
				goto LOOP
			}

			if _, err := data.Write(buf[:n]); err != nil {
				w.OnError("Error reading data", err)
				return
			}
			buf = buf[:0]
			if n < 1024 {
				break
			}
		}
		w.onMessage(data)
		binaryFlag = false
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
		w.SetWritable(true)
		w.Emit("drain")
	}()

	w.musend.Lock()
	defer w.musend.Unlock()

	for _, packet := range packets {
		w.write(packet)
	}
}

func (w *webTransport) write(packet *packet.Packet) {
	data, err := w.Parser().EncodePacket(packet, w.SupportsBinary())
	if err != nil {
		wt_log.Debug(`Send Error "%s"`, err)
		return
	}

	if _, ok := data.(*_types.BytesBuffer); ok {
		wt_log.Debug("writing binary header")
		if _, err := w.session.Stream.Write(BINARY_HEADER); err != nil {
			w.OnError("write error", err)
			return
		}
	}

	wt_log.Debug(`writing chunk: "%s"`, data)
	if _, err := io.Copy(w.session.Stream, data); err != nil {
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
