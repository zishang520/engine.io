package transports

import (
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"io"
	"strings"
)

type Polling interface {
	Transport
}

type polling struct {
	*transport

	closeTimeout      int
	maxHttpBufferSize int
	httpCompression   int
	pollCtx           *types.HttpContext
	dataCtx           *types.HttpContext

	writable    bool
	shouldClose types.Fn
}

func NewPolling(ctx *types.HttpContext) Polling {
	p := &polling{
		transport:    NewTransport(ctx),
		closeTimeout: 30 * 1000,
		// maxHttpBufferSize = null;
		// httpCompression = null;
		writable: false,
	}
	return p
}

func (p *polling) Name() string {
	return "polling"
}

func (p *polling) HandlesUpgrades() bool {
	return false
}

func (p *polling) SupportsFraming() bool {
	return false
}

func (p *polling) OnRequest(ctx *types.HttpContext) {
	method := strings.ToUpper(ctx.Request.Method)
	if "GET" == method {
		p.OnPollRequest(ctx)
	} else if "POST" == method {
		p.OnDataRequest(ctx)
	} else {
		ctx.Response.WriteHeader(500)
		ctx.response.Write(nil)
	}
}

func (p *polling) OnPollRequest(ctx *types.HttpContext) {
	if p.pollCtx != nil {
		utils.Log.Debug("request overlap")
		// assert: p.res, '.req and .res should be (un)set together'
		p.OnError("overlap from client")
		ctx.Response.WriteHeader(500)
		ctx.response.Write(nil)
		return
	}

	utils.Log.Debug("setting request")

	p.pollCtx = ctx

	// function onClose() {
	//   p.OnError("poll connection closed prematurely");
	// }

	// function cleanup() {
	//   req.removeListener("close", onClose);
	//   self.req = self.res = null;
	// }

	// req.cleanup = cleanup;
	// req.on("close", onClose);

	p.writable = true
	p.Emit("drain")

	// if we're still writable but had a pending close, trigger an empty send
	if this.writable && this.shouldClose != nil {
		utils.Log.Debug("triggering empty send to append close packet")
		p.Send([]*packet.Packet{
			&packet.Packet{
				Type: packet.NOOP,
			},
		})
	}
}

func (p *polling) OnDataRequest(ctx *types.HttpContext) {
	if p.dataCtx {
		// assert: p.dataRes, '.dataReq and .dataRes should be (un)set together'
		p.OnError("data request overlap from client")
		ctx.Response.WriteHeader(500)
		ctx.response.Write(nil)
		return
	}

	isBinary := "application/octet-stream" == ctx.Request.Header.Get("Content-Type")

	if isBinary && this.Protocol == 4 {
		p.OnError("invalid content")
		return
	}

	this.dataCtx = ctx

	// let chunks = isBinary ? Buffer.concat([]) : "";
	// const self = this;

	// function cleanup() {
	//   req.removeListener("data", onData);
	//   req.removeListener("end", onEnd);
	//   req.removeListener("close", onClose);
	//   self.dataReq = self.dataRes = chunks = null;
	// }

	// function onClose() {
	//   cleanup();
	//   self.onError("data request connection closed prematurely");
	// }

	// function onData(data) {
	//   let contentLength;
	//   if (isBinary) {
	//     chunks = Buffer.concat([chunks, data]);
	//     contentLength = chunks.length;
	//   } else {
	//     chunks += data;
	//     contentLength = Buffer.byteLength(chunks);
	//   }

	//   if (contentLength > self.maxHttpBufferSize) {
	//     chunks = isBinary ? Buffer.concat([]) : "";
	//     req.connection.destroy();
	//   }
	// }

	// function onEnd() {
	//   self.onData(chunks);

	//   const headers = {
	//     // text/html is required instead of text/plain to avoid an
	//     // unwanted download dialog on certain user-agents (GH-43)
	//     "Content-Type": "text/html",
	//     "Content-Length": 2
	//   };

	//   res.writeHead(200, self.headers(req, headers));
	//   res.end("ok");
	//   cleanup();
	// }

	// req.on("close", onClose);
	// if (!isBinary) req.setEncoding("utf8");
	// req.on("data", onData);
	// req.on("end", onEnd);
}

func (p *polling) OnData(data io.Reader) {
	utils.Log.Debug(`received "%s"`, data)

	// const callback = function(packet) {
	//   if ("close" === packet.type) {
	//     debug("got xhr close packet");
	//     self.onClose();
	//     return false;
	//   }

	//   self.onPacket(packet);
	// };

	for packet := range p.Parser.DecodePayload(data) {
		if "close" == packet.Type {
			utils.Log.Debug("got xhr close packet")
			p.OnClose()
			return
		}

		p.OnPacket(packet)
	}
}

func (p *polling) OnClose() {
	// if (this.writable) {
	//   // close pending poll request
	//   this.send([{ type: "noop" }]);
	// }
	// super.onClose();
}

func (p *polling) Send(packets []*packet.Packet) {
	// this.writable = false;

	// if (this.shouldClose) {
	//   debug("appending close packet to payload");
	//   packets.push({ type: "close" });
	//   this.shouldClose();
	//   this.shouldClose = null;
	// }

	// const doWrite = data => {
	//   const compress = packets.some(packet => {
	//     return packet.options && packet.options.compress;
	//   });
	//   this.write(data, { compress });
	// };

	// if (this.Protocol === 3) {
	//   this.parser.encodePayload(packets, this.supportsBinary, doWrite);
	// } else {
	//   this.parser.encodePayload(packets, doWrite);
	// }
}

func (p *polling) Write(data, options) {
	// debug('writing "%s"', data);
	// const self = this;
	// this.doWrite(data, options, function() {
	//   self.req.cleanup();
	// });
}

func (p *polling) DoWrite(data, options, callback) {
	// const self = this;

	// // explicit UTF-8 is required for pages not served under utf
	// const isString = typeof data === "string";
	// const contentType = isString
	//   ? "text/plain; charset=UTF-8"
	//   : "application/octet-stream";

	// const headers = {
	//   "Content-Type": contentType
	// };

	// if (!this.httpCompression || !options.compress) {
	//   respond(data);
	//   return;
	// }

	// const len = isString ? Buffer.byteLength(data) : data.length;
	// if (len < this.httpCompression.threshold) {
	//   respond(data);
	//   return;
	// }

	// const encoding = accepts(this.req).encodings(["gzip", "deflate"]);
	// if (!encoding) {
	//   respond(data);
	//   return;
	// }

	// this.compress(data, encoding, function(err, data) {
	//   if (err) {
	//     self.res.writeHead(500);
	//     self.res.end();
	//     callback(err);
	//     return;
	//   }

	//   headers["Content-Encoding"] = encoding;
	//   respond(data);
	// });

	// function respond(data) {
	//   headers["Content-Length"] =
	//     "string" === typeof data ? Buffer.byteLength(data) : data.length;
	//   self.res.writeHead(200, self.headers(self.req, headers));
	//   self.res.end(data);
	//   callback();
	// }
}

func (p *polling) Compress(data, encoding, callback) {
	// debug("compressing");

	// const buffers = [];
	// let nread = 0;

	// compressionMethods[encoding](this.httpCompression)
	//   .on("error", callback)
	//   .on("data", function(chunk) {
	//     buffers.push(chunk);
	//     nread += chunk.length;
	//   })
	//   .on("end", function() {
	//     callback(null, Buffer.concat(buffers, nread));
	//   })
	//   .end(data);
}

func (p *polling) DoClose(fn) {
	// debug("closing");

	// const self = this;
	// let closeTimeoutTimer;

	// if (this.dataReq) {
	//   debug("aborting ongoing data request");
	//   this.dataReq.destroy();
	// }

	// if (this.writable) {
	//   debug("transport writable - closing right away");
	//   this.send([{ type: "close" }]);
	//   onClose();
	// } else if (this.discarded) {
	//   debug("transport discarded - closing right away");
	//   onClose();
	// } else {
	//   debug("transport not writable - buffering orderly close");
	//   this.shouldClose = onClose;
	//   closeTimeoutTimer = setTimeout(onClose, this.closeTimeout);
	// }

	// function onClose() {
	//   clearTimeout(closeTimeoutTimer);
	//   fn();
	//   self.onClose();
	// }
}

func (p *polling) Headers(req, headers) {
	// headers = headers || {};

	// // prevent XSS warnings on IE
	// // https://github.com/LearnBoost/socket.io/pull/1333
	// const ua = req.headers["user-agent"];
	// if (ua && (~ua.indexOf(";MSIE") || ~ua.indexOf("Trident/"))) {
	//   headers["X-XSS-Protection"] = "0";
	// }

	// this.emit("headers", headers);
	// return headers;
}
