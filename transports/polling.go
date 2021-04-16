package transports

import (
	"compress/gzip"
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
	httpCompression   interface{}
	pollCtx           *types.HttpContext
	dataCtx           *types.HttpContext

	writable       bool
	shouldClose    types.Fn
	supportsBinary bool
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
	if this.writable {
		// close pending poll request
		p.Send([]*packet.Packet{
			&packet.Packet{
				Type: packet.NOOP,
			},
		})
	}
	p.transport.OnClose()
}

func (p *polling) Send(packets []*packet.Packet) {
	p.writable = false

	if p.shouldClose != nil {
		utils.Log.Debug("appending close packet to payload")
		packets = append(packets, &packet.Packet{
			Type: packet.CLOSE,
		})
		p.shouldClose()
		p.shouldClose = nil
	}

	// const doWrite = data => {
	//   const compress = packets.some(packet => {
	//     return packet.options && packet.options.compress;
	//   });
	//   this.write(data, { compress });
	// };

	if p.Protocol == 3 {
		data, _ := this.parser.EncodePayload(packets, p.supportsBinary)
		p.Write(data, nil)
	} else {
		data, _ := this.parser.EncodePayload(packets)
		p.Write(data, nil)
	}
}

func (p *polling) Write(data io.Reader, options interface{}) {
	utils.Log.Debug(`writing "%s"`, data)
	p.DoWrite(data, options, func() {
		//   self.req.cleanup();
	})
}

func (p *polling) DoWrite(data io.Reader, options interface{}, callback types.Fn) {
	if c, ok := data.(io.Closer); ok {
		defer c.Close()
	}

	// explicit UTF-8 is required for pages not served under utf
	_, isString := packet.Data.(*types.StringBuffer)
	contentType := "application/octet-stream"
	if isString {
		contentType = "text/plain; charset=UTF-8"
	}

	headers := map[string]string{
		"Content-Type": contentType,
	}

	respond := func(data io.Reader) {
		ctx.Response.WriteHeader(200)
		_data := bufio.NewReader(data)
		headers["Content-Length"] = strconv.Itoa(_data.Size())
		for key, value := range p.Headers(p.ctx, headers) {
			ctx.Response.Header().Add(key, value)
		}
		io.Copy(ctx.Response, _data)
		callback()
	}

	// if p.httpCompression == nil || !options.compress {
	// 	respond(data)
	// 	return
	// }

	// _data := bufio.NewReader(data)
	// if _data.Size() < p.httpCompression.threshold {
	// 	respond(data)
	// 	return
	// }

	// const encoding = accepts(this.req).encodings(["gzip", "deflate"]);
	// if (!encoding) {
	//   respond(data);
	//   return;
	// }

	// p.Compress(data, encoding, func(err, data) {
	//   if (err) {
	//     self.res.writeHead(500);
	//     self.res.end();
	//     callback(err);
	//     return;
	//   }

	//   headers["Content-Encoding"] = encoding;
	//   respond(data);
	// });
}

func (p *polling) Compress(data, encoding, callback) {
	utils.Log.Debug("compressing")

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

func (p *polling) DoClose(fn types.Fn) {
	utils.Log.Debug("closing")

	closeTimeoutTimer := make(chan struct{})

	if p.dataCtx {
		utils.Log.Debug("aborting ongoing data request")
		p.dataCtx.destroy()
	}

	onClose := func() {
		closeTimeoutTimer <- struct{}{}
		fn()
		p.OnClose()
	}

	if p.writable {
		utils.Log.Debug("transport writable - closing right away")
		p.Send([]*packet.Packet{
			&packet.Packet{
				Type: packet.CLOSE,
			},
		})
		onClose()
	} else if p.discarded {
		utils.Log.Debug("transport discarded - closing right away")
		onClose()
	} else {
		utils.Log.Debug("transport not writable - buffering orderly close")
		p.shouldClose = onClose
		go func() {
			select {
			case <-time.After(p.closeTimeout * time.Millisecond):
				onClose()
			case <-closeTimeoutTimer:
			}
		}()
	}
}

func (p *polling) Headers(ctx *types.HttpContext, headers ...map[string]string) map[string]string {
	headers = append(headers, map[string]string{})

	// prevent XSS warnings on IE
	// https://github.com/socketio/socket.io/pull/1333
	ua := ctx.Request.UserAgent()
	if len(ua) > 0 && ((strings.Index(ua, ";MSIE") > -1) || (strings.Index(ua, "Trident/") > -1)) {
		headers[0]["X-XSS-Protection"] = "0"
	}
	x.Emit("headers", headers[0])
	return headers[0]
}
