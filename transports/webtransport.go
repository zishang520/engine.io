package transports

import (
	"context"
	"io"

	"github.com/quic-go/webtransport-go"
	"github.com/zishang520/engine.io-go-parser/packet"
	_types "github.com/zishang520/engine.io-go-parser/types"
	"github.com/zishang520/engine.io/log"
	"github.com/zishang520/engine.io/types"
)

var wt_log = log.NewLog("engine:webtransport")

var BINARY_HEADER = []byte{0x36}

type webTransport struct {
	*transport

	session *types.WebTransportConn
	stream  webtransport.Stream
}

// WebTransport transport
func NewWebTransport(ctx *types.HttpContext) *webTransport {
	w := &webTransport{}
	return w.New(ctx)
}

func (w *webTransport) New(ctx *types.HttpContext) *webTransport {
	w.transport = &transport{}

	// Advertise framing support.
	w.supportsFraming = true

	// Advertise upgrade support.
	w.handlesUpgrades = true

	// Transport name
	w.name = "webtransport"

	w.transport.New(ctx)

	w.session = ctx.WebTransport
	w.SetWritable(true)
	w.perMessageDeflate = nil

	w.doClose = w.WebTransportDoClose
	w.send = w.WebTransportSend

	go w._init(ctx)

	w.session.On("error", func(errors ...any) {
		w.OnError("webTransport error", errors[0].(error))
	})
	w.session.On("close", func(...any) {
		w.OnClose()
	})

	return w
}

func (w *webTransport) _init(ctx *types.HttpContext) {
	var err error
	// Wait for incoming bidi stream
	w.stream, err = w.session.AcceptStream(context.Background())
	if err != nil {
		w.OnError("Error reading data", err)
		return
	}

	defer w.stream.Close()

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
			n, err := w.stream.Read(buf)
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
				break LOOP
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
		w.WebTransportOnData(data)
		binaryFlag = false
	}
}

func (w *webTransport) WebTransportOnData(data _types.BufferInterface) {
	wt_log.Debug(`webTransport received "%s"`, data)
	w.TransportOnData(data)
}

// Writes a packet payload.
func (w *webTransport) WebTransportSend(packets []*packet.Packet) {
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
	data, err := w.parser.EncodePacket(packet, w.supportsBinary)
	if err != nil {
		wt_log.Debug(`Send Error "%s"`, err)
		return
	}

	if _, ok := data.(*_types.BytesBuffer); ok {
		wt_log.Debug("writing binary header")
		if _, err := w.stream.Write(BINARY_HEADER); err != nil {
			w.OnError("write error", err)
			return
		}
	}

	wt_log.Debug(`writing chunk: "%s"`, data)
	if _, err := io.Copy(w.stream, data); err != nil {
		w.OnError("write error", err)
		return
	}
}

// Closes the transport.
func (w *webTransport) WebTransportDoClose(fn ...types.Callable) {
	wt_log.Debug(`closing WebTransport session`)
	w.session.CloseWithError(0, "")
	if len(fn) > 0 {
		(fn[0])()
	}
}
