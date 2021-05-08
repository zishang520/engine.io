package transports

import (
	"compress/flate"
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

type polling struct {
	*transport

	closeTimeout int
	// maxHttpBufferSize int
	httpCompression *types.HttpCompression
	pollCtx         bool
	dataCtx         *types.HttpContext

	writable       bool
	shouldClose    types.Fn
	supportsBinary bool
}

func NewPolling(ctx *types.HttpContext) *polling {
	p := &polling{
		transport:    NewTransport(ctx),
		closeTimeout: 30 * 1000,
		// maxHttpBufferSize: 0,
		httpCompression: &types.HttpCompression{Threshold: 1024},
		writable:        false,
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

func (p *polling) UpgradesTo() *types.Set {
	return &types.Set{"websocket": types.NULL}
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
	if p.pollCtx {
		utils.Log.Debug("request overlap")
		// assert: p.res, '.req and .res should be (un)set together'
		p.OnError("overlap from client")
		ctx.Response.WriteHeader(500)
		ctx.response.Write(nil)
		return
	}

	utils.Log.Debug("setting request")

	p.pollCtx = true

	_RemoveListener := make(chan struct{})
	p.ctx.Cleanup = func() {
		_RemoveListener <- struct{}{}
		p.ctx = nil
	}
	go func() {
		select {
		case <-p.ctx.Request.(http.CloseNotifier).CloseNotify():
			p.OnError("poll connection closed prematurely")
		case <-_RemoveListener:
		}
	}()

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
	if p.dataCtx != nil {
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

	go func() {
		select {
		case <-p.ctx.Request.(http.CloseNotifier).CloseNotify():
			p.OnError("data request connection closed prematurely")
		}
	}()
	// text/html is required instead of text/plain to avoid an
	// unwanted download dialog on certain user-agents (GH-43)
	headers := map[string]string{
		"Content-Type":   "text/html",
		"Content-Length": "2",
	}
	ctx.Response.WriteHeader(200)
	for key, value := range p.Headers(p.ctx, headers) {
		ctx.Response.Header().Set(key, value)
	}
	ctx.Response.WriteString("OK")
	p.dataCtx = nil
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

	doWrite := func(data io.Reader) {
		option := &packet.Option{false}
		for _, packet := range packets {
			if packet.Options != nil && packet.Options.Compress {
				option.Compress = true
				break
			}
		}
		p.Write(data, option)
	}

	if p.Protocol == 3 {
		data, _ := this.Parser.EncodePayload(packets, p.supportsBinary)
		doWrite(data)
	} else {
		data, _ := this.Parser.EncodePayload(packets)
		doWrite(data)
	}
}

func (p *polling) Write(data io.Reader, options *packet.Option) {
	utils.Log.Debug(`writing "%s"`, data)
	p.DoWrite(data, options, func() {
		p.ctx.Cleanup()
	})
}

func (p *polling) DoWrite(data io.Reader, options *packet.Option, callback types.Fn) {
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

	_data := bufio.NewReader(data)

	respond := func(data io.Reader, length string) {
		if c, ok := data.(io.Closer); ok {
			defer c.Close()
		}
		ctx.Response.WriteHeader(200)
		headers["Content-Length"] = length
		for key, value := range p.Headers(p.ctx, headers) {
			ctx.Response.Header().Set(key, value)
		}
		io.Copy(ctx.Response, data)
		callback()
	}

	if p.httpCompression == nil || options == nil || !options.Compress {
		respond(_data, strconv.Itoa(_data.Size()))
		return
	}

	if _data.Size() < p.httpCompression.Threshold {
		respond(_data, strconv.Itoa(_data.Size()))
		return
	}

	encoding := utils.Contains(ctx.Request.Header.Get("Accept-Encoding"), []string{"gzip", "deflate"})
	if encoding != "" {
		respond(_data, strconv.Itoa(_data.Size()))
		return
	}

	if buf, err := p.Compress(data, encoding); err == nil {
		headers["Content-Encoding"] = encoding
		respond(buf, strconv.Itoa(buf.Size()))
	}
}

func (p *polling) Compress(data io.Reader, encoding string) *bufio.Reader {
	if c, ok := data.(io.Closer); ok {
		defer c.Close()
	}
	utils.Log.Debug("compressing")
	buf := bufio.NewReader(nil)
	switch encoding {
	case "gzip":
		gz, err := gzip.NewWriterLevel(buf, 1)
		if err != nil {
			return buf, err
		}
		defer gz.Close()
		io.Copy(gz, data)
		break
	case "deflate":
		flate, err := flate.NewWriter(buf, 1)
		if err != nil {
			return buf, err
		}
		defer flate.Close()
		io.Copy(flate, data)
		break
	}
	return buf, nil
}

func (p *polling) DoClose(fn types.Fn) {
	utils.Log.Debug("closing")

	closeTimeoutTimer := make(chan struct{})

	if p.dataCtx {
		utils.Log.Debug("aborting ongoing data request")
		p.xtx.Request.Close = true
		// p.dataCtx.destroy()
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
	if (len(ua) > 0) && ((strings.Index(ua, ";MSIE") > -1) || (strings.Index(ua, "Trident/") > -1)) {
		headers[0]["X-XSS-Protection"] = "0"
	}
	x.Emit("headers", headers[0])
	return headers[0]
}
