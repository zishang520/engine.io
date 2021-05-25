package transports

import (
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
)

type websocket struct {
	*transport

	supportsBinary    bool
	writable          bool
	perMessageDeflate *types.PerMessageDeflate
	socket            *types.WebSocketConn
}

func NewWebSocket(ctx *types.HttpContext) *websocket {

	s := &websocket{}
	s.socket = ctx.Websocket
	go func() {
		for {
			mt, message, err := s.socket.NextReader()
			if err != nil {
				// log.Println("read:", err)
				break
			}
			defer message.Close()

			switch mt {
			case websocket.BinaryMessage:
				read := types.NewBytesBuffer(nil)
				read.ReadFrom(message)
				s.OnData(read)
			case websocket.TextMessage:
				read := types.NewStringBuffer(nil)
				read.ReadFrom(message)
				s.OnData(read)
			case websocket.CloseMessage:
				s.OnClose()
			}
		}
	}()
	s.socket.On("error", s.onError.bind(s))
	s.socket.On("headers", func(headers ...interface{}) {
		s.Emit("headers", headers...)
	})
	s.writable = true
	s.perMessageDeflate = nil
	s._DoClose(s.doClose)
	return s
}

func (w *websocket) Name() string {
	return "websocket"
}

func (w *websocket) SupportsFraming() bool {
	return true
}

func (w *websocket) OnData(data io.Reader) {
	utils.Log.Debug(`received "%s"`, data)
	w.transport.OnData(data)
}

func (w *websocket) Send(packets []*packet.Packet) {
	onEnd := func(err error) {
		if err != nil {
			return w.OnError("write error", err.Error())
		}
		w.writable = true
		w.Emit("drain")
	}

	send := func(data types.PacketBuffer, packet *packet.Packet) {
		utils.Log.Debug(`writing "%s"`, data)

		// always creates a new object since ws modifies it
		compress := false
		if packet.Options != nil {
			compress = packet.Options.Compress
		}
		if w.perMessageDeflate != nil {
			if data.Size() < p.perMessageDeflate.Threshold {
				compress = false
			}
		}
		w.writable = false
		w.socket.EnableWriteCompression(compress)
		mt := websocket.BinaryMessage
		if _, ok := data.(*types.StringBuffer); ok {
			mt = websocket.TextMessage
		}
		onEnd(w.socket.WriteMessage(mt, data.Bytes()))
	}

	for _, packet := range packets {
		if buf, err := w.Parser.EncodePacket(packet, w.supportsBinary); err != nil {
			utils.Log.Debug(`Send Error "%s"`, err)
			continue
		}
		send(buf, packet)
	}
}

func (w *websocket) doClose(fn types.Fn) {
	utils.Log.Debug(`closing`)
	w.socket.Close()
	fn()
}
