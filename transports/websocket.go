package transports

import (
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
)

type WebSocket interface {
	Transport
}

type websocket struct {
	*transport

	supportsBinary    bool
	writable          bool
	perMessageDeflate *types.PerMessageDeflate
}

func NewWebSocket(ctx *types.HttpContext) *websocket {

	s := &websocket{}
	//  s.socket = req.websocket;
	// s.socket.on("message", s.onData.bind(s));
	// s.socket.once("close", s.onClose.bind(s));
	// s.socket.on("error", s.onError.bind(s));
	// s.socket.on("headers", headers => {
	//   s.emit("headers", headers);
	// });
	s.writable = true
	s.perMessageDeflate = nil
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

	send := func(data *types.StringBuffer, packet *packet.Packet) {
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
		w.socket.Send(data, opts, onEnd)
	}

}

func (w *websocket) DoClose(fn ...types.Fn) {
	utils.Log.Debug(`closing`)
	// w.socket.Close()
	if len(fn) > 0 {
		fn[0]()
	}
}
