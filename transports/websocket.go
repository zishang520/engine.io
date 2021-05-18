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
	// s.socket.On("message", s.onData.bind(s))
	go func() {
		for {
			mt, message, err := s.socket.NextReader()
			if err != nil {
				// log.Println("read:", err)
				break
			}
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
			// log.Printf("recv: %s", message)
			// err = c.WriteMessage(mt, message)
			// if err != nil {
			// 	log.Println("write:", err)
			// 	break
			// }
		}
	}()
	// s.socket.Once("close", s.OnClose.bind(s))
	s.socket.On("error", s.onError.bind(s))
	s.socket.On("headers", func(headers ...interface{}) {
		s.Emit("headers", headers...)
	})
	s.writable = true
	s.perMessageDeflate = nil
	s.DoClose(s.doClose)
	return s
}

func (w *websocket) Name() string {
	return "websocket"
}

func (w *websocket) HandlesUpgrades() bool {
	return true
}

func (w *websocket) SupportsFraming() bool {
	return true
}

func (p *polling) UpgradesTo() *types.Set {
	return &types.Set{}
}

func (w *websocket) OnData(data io.Reader) {
	utils.Log.Debug(`received "%s"`, data)
	w.transport.OnData(data)
}

func (w *websocket) Send(packets []*packet.Packet) {
	for _, packet := range packets {
		if buf, err := w.Parser.EncodePacket(packet, w.supportsBinary); err != nil {
			utils.Log.Debug(`Send Error "%s"`, err)
			continue
		}
		send(buf, packet)
	}

	onEnd := func(err ...interface{}) {
		if len(err) > 0 {
			return w.OnError("write error", err[0].Error())
		}
		w.writable = true
		w.Emit("drain")
	}

	send := func(data io.Reader, packet *packet.Packet) {
		utils.Log.Debug(`writing "%s"`, data)

		// always creates a new object since ws modifies it
		opts := &packet.Option{false}
		if packet.Options != nil {
			opts.Compress = packet.Options.Compress
		}
		if w.perMessageDeflate != nil {
			if data.Size() < p.perMessageDeflate.Threshold {
				opts.compress = false
			}
		}
		w.writable = false
		w.socket.EnableWriteCompression(opts.compress)
		var mt int
		w.socket.WriteMessage()
		// err = c.WriteMessage(mt, message)
		// if err != nil {
		// 	log.Println("write:", err)
		// 	break
		// }
		w.socket.Send(data, opts, onEnd)
	}

}

func (w *websocket) doClose(fn types.Fn) {
	utils.Log.Debug(`closing`)
	w.socket.Close()
	fn()
}
