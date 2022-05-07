package transports

import (
	"compress/flate"
	"compress/gzip"
	"github.com/andybalholm/brotli"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type polling struct {
	*transport

	dataCtx     *types.HttpContext
	shouldClose types.Callable
}

func NewPolling(ctx *types.HttpContext) *polling {
	p := &polling{}
	return p.New(ctx)
}

func (p *polling) New(ctx *types.HttpContext) *polling {

	p.transport = &transport{}

	p.supportsFraming = false
	p.name = "polling"

	p.transport.New(ctx)

	p.onClose = p.PollingOnClose
	p.onData = p.PollingOnData
	p.doWrite = p.PollingDoWrite
	p.doClose = p.PollingDoClose
	p.send = p.PollingSend

	p.closeTimeout = 30 * 1000

	return p
}

func (p *polling) OnRequest(ctx *types.HttpContext) {
	method := ctx.Method()

	if http.MethodGet == method {
		p.onPollRequest(ctx)
	} else if http.MethodPost == method {
		p.onDataRequest(ctx)
	} else {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.Write(nil)
	}
}

func (p *polling) onPollRequest(ctx *types.HttpContext) {
	if p.req != nil {
		utils.Log().Debug("request overlap")
		// assert: p.res, '.req and .res should be (un)set together'
		p.OnError("overlap from client")
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.Write(nil)
		return
	}

	utils.Log().Debug("setting request")

	p.req = ctx

	var onClose events.Listener

	onClose = func(...interface{}) {
		p.OnError("poll connection closed prematurely")
	}

	ctx.Cleanup = func() {
		ctx.RemoveListener("close", onClose)
		p.req = nil
	}

	ctx.On("close", onClose)

	p.writable = true
	p.Emit("drain")

	// if we're still writable but had a pending close, trigger an empty send
	if p.writable && p.shouldClose != nil {
		utils.Log().Debug("triggering empty send to append close packet")
		p.Send([]*packet.Packet{
			&packet.Packet{
				Type: packet.NOOP,
			},
		})
	}
}

func (p *polling) onDataRequest(ctx *types.HttpContext) {
	if p.dataCtx != nil {
		// assert: p.dataRes, '.dataReq and .dataRes should be (un)set together'
		p.OnError("data request overlap from client")
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.Write(nil)
		return
	}

	isBinary := "application/octet-stream" == ctx.Headers().Get("Content-Type")

	if isBinary && p.protocol == 4 {
		p.OnError("invalid content")
		return
	}

	p.dataCtx = ctx

	var onClose events.Listener

	cleanup := func() {
		ctx.RemoveListener("close", onClose)
		p.dataCtx = nil
	}

	onClose = func(...interface{}) {
		cleanup()
		p.OnError("data request connection closed prematurely")
	}

	ctx.On("close", onClose)

	if ctx.Request().ContentLength > p.maxHttpBufferSize {
		ctx.SetStatusCode(http.StatusRequestEntityTooLarge)
		ctx.Write(nil)
		cleanup()
		return
	}

	var packet types.BufferInterface
	if isBinary {
		packet = types.NewBytesBuffer(nil)
	} else {
		packet = types.NewStringBuffer(nil)
	}
	if rc, ok := ctx.Request().Body.(io.ReadCloser); ok && rc != nil {
		packet.ReadFrom(rc)
		rc.Close()
	}
	p.OnData(packet)

	headers := map[string]string{
		// text/html is required instead of text/plain to avoid an
		// unwanted download dialog on certain user-agents (GH-43)
		"Content-Type":   "text/html",
		"Content-Length": "2",
	}
	for key, value := range p.Headers(ctx, headers) {
		ctx.Response().Header().Set(key, value)
	}
	ctx.SetStatusCode(http.StatusOK)
	io.WriteString(ctx, "ok")
	cleanup()
}

func (p *polling) PollingOnData(data types.BufferInterface) {
	utils.Log().Debug(`received "%s"`, data)

	for _, packetData := range p.parser.DecodePayload(data) {
		if packet.CLOSE == packetData.Type {
			utils.Log().Debug("got xhr close packet")
			p.OnClose()
			return
		}

		p.OnPacket(packetData)
	}
}

func (p *polling) PollingOnClose() {
	if p.writable {
		// close pending poll request
		p.Send([]*packet.Packet{
			&packet.Packet{
				Type: packet.NOOP,
			},
		})
	}
	p.TransportOnClose()
}

func (p *polling) PollingSend(packets []*packet.Packet) {
	p.writable = false

	if p.shouldClose != nil {
		utils.Log().Debug("appending close packet to payload")
		packets = append(packets, &packet.Packet{
			Type: packet.CLOSE,
		})
		p.shouldClose()
		p.shouldClose = nil
	}

	doWrite := func(data types.BufferInterface) {
		option := &packet.Options{false}
		for _, packetData := range packets {
			if packetData.Options != nil && packetData.Options.Compress {
				option.Compress = true
				break
			}
		}
		p.Write(data, option)
	}

	if p.protocol == 3 {
		data, _ := p.parser.EncodePayload(packets, p.supportsBinary)
		doWrite(data)
	} else {
		data, _ := p.parser.EncodePayload(packets)
		doWrite(data)
	}
}

func (p *polling) Write(data types.BufferInterface, options *packet.Options) {
	utils.Log().Debug(`writing "%s"`, data)
	p.DoWrite(data, options, func() { p.req.Cleanup() })
}

func (p *polling) PollingDoWrite(data types.BufferInterface, options *packet.Options, callback types.Callable) {
	contentType := "application/octet-stream"
	// explicit UTF-8 is required for pages not served under utf
	switch data.(type) {
	case *types.StringBuffer:
		contentType = "text/plain; charset=UTF-8"
	}

	headers := map[string]string{
		"Content-Type": contentType,
	}

	respond := func(data types.BufferInterface, length string) {
		headers["Content-Length"] = length
		for key, value := range p.Headers(p.req, headers) {
			p.req.Response().Header().Set(key, value)
		}
		p.req.SetStatusCode(http.StatusOK)
		io.Copy(p.req, data)
		callback()
	}

	if p.httpCompression == nil || options == nil || !options.Compress {
		respond(data, strconv.Itoa(data.Len()))
		return
	}

	if data.Len() < p.httpCompression.Threshold {
		respond(data, strconv.Itoa(data.Len()))
		return
	}

	encoding := utils.Contains(p.req.Headers().Get("Accept-Encoding"), []string{"gzip", "deflate", "br"})
	if encoding == "" {
		respond(data, strconv.Itoa(data.Len()))
		return
	}

	if buf, err := p.compress(data, encoding); err == nil {
		headers["Content-Encoding"] = encoding
		respond(buf, strconv.Itoa(buf.Len()))
	}
}

func (p *polling) compress(data types.BufferInterface, encoding string) (types.BufferInterface, error) {
	utils.Log().Debug("compressing")
	buf := types.NewBytesBuffer(nil)
	switch encoding {
	case "gzip":
		gz, err := gzip.NewWriterLevel(buf, 1)
		if err != nil {
			return buf, err
		}
		defer gz.Close()
		io.Copy(gz, data)
	case "deflate":
		fl, err := flate.NewWriter(buf, 1)
		if err != nil {
			return buf, err
		}
		defer fl.Close()
		io.Copy(fl, data)
	case "br":
		br := brotli.NewWriterLevel(buf, 1)
		defer br.Close()
		io.Copy(br, data)
	}
	return buf, nil
}

func (p *polling) PollingDoClose(fn ...types.Callable) {
	utils.Log().Debug("closing")

	var closeTimeoutTimer *utils.Timer = nil

	if p.dataCtx != nil {
		utils.Log().Debug("aborting ongoing data request")
		if h, ok := p.dataCtx.Response().(http.Hijacker); ok {
			if netConn, _, err := h.Hijack(); err == nil {
				netConn.Close()
			}
		}
	}

	onClose := func() {
		utils.ClearTimeout(closeTimeoutTimer)
		if len(fn) > 0 {
			(fn[0])()
		}
		p.OnClose()
	}

	if p.writable {
		utils.Log().Debug("transport writable - closing right away")
		p.Send([]*packet.Packet{
			&packet.Packet{
				Type: packet.CLOSE,
			},
		})
		onClose()
	} else if p.discarded {
		utils.Log().Debug("transport discarded - closing right away")
		onClose()
	} else {
		utils.Log().Debug("transport not writable - buffering orderly close")
		p.shouldClose = onClose
		closeTimeoutTimer = utils.SetTimeOut(onClose, p.closeTimeout*time.Millisecond)
	}
}

func (p *polling) Headers(ctx *types.HttpContext, headers ...map[string]string) map[string]string {
	headers = append(headers, map[string]string{})

	// prevent XSS warnings on IE
	// https://github.com/socketio/socket.io/pull/1333
	ua := ctx.UserAgent()
	if (len(ua) > 0) && ((strings.Index(ua, ";MSIE") > -1) || (strings.Index(ua, "Trident/") > -1)) {
		headers[0]["X-XSS-Protection"] = "0"
	}
	p.Emit("headers", headers[0], ctx)
	return headers[0]
}