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
}

func NewWebSocket(req interface{}) WebSocket {

	s := &websocket{}
	//  s.socket = req.websocket;
	// s.socket.on("message", s.onData.bind(s));
	// s.socket.once("close", s.onClose.bind(s));
	// s.socket.on("error", s.onError.bind(s));
	// s.socket.on("headers", headers => {
	//   s.emit("headers", headers);
	// });
	// s.writable = true;
	// s.perMessageDeflate = null;
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

func (w *websocket) OnData(data io.Reader) {
	// debug('received "%s"', data);
	w.transport.OnData(data)
}

func (w *websocket) Send(packets []*packet.Packet) {
	// for (var i = 0; i < packets.length; i++) {
	//   var packet = packets[i];
	//   w.parser.EncodePacket(packet, self.supportsBinary, send);
	// }
	// buf, err := parser.ParserV4.EncodePayload([]*packet.Packet{

	// function send(data) {
	//   debug('writing "%s"', data);

	//   // always creates a new object since ws modifies it
	//   var opts = {};
	//   if (packet.options) {
	//     opts.compress = packet.options.compress;
	//   }

	//   if (self.perMessageDeflate) {
	//     var len =
	//       "string" === typeof data ? Buffer.byteLength(data) : data.length;
	//     if (len < self.perMessageDeflate.threshold) {
	//       opts.compress = false;
	//     }
	//   }

	//   self.writable = false;
	//   self.socket.send(data, opts, onEnd);
	// }

	// function onEnd(err) {
	//   if (err) return self.onError("write error", err.stack);
	//   self.writable = true;
	//   self.Emit("drain");
	// }
}

func (w *websocket) DoClose(fn) {
	// debug("closing")
	// w.socket.Close()
	// fn && fn()
}
