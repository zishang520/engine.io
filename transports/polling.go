package transports

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/log"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
)

var polling_log = log.NewLog("engine:polling")

type polling struct {
	*transport

	dataCtx    *types.HttpContext
	mu_dataCtx sync.RWMutex

	shouldClose types.Callable
}

// HTTP polling New.
func NewPolling(ctx *types.HttpContext) *polling {
	p := &polling{}
	return p.New(ctx)
}

func (p *polling) New(ctx *types.HttpContext) *polling {

	p.transport = &transport{}

	p.supportsFraming = false

	// Transport name
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

// Overrides onRequest.
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

// The client sends a request awaiting for us to send data.
func (p *polling) onPollRequest(ctx *types.HttpContext) {
	p.mu_req.RLock()
	if p.req != nil {
		defer p.mu_req.RUnlock()
		polling_log.Debug("request overlap")
		// assert: p.res, '.req and .res should be (un)set together'
		p.OnError("overlap from client")
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.Write(nil)
		return
	}
	p.mu_req.RUnlock()

	polling_log.Debug("setting request")

	onClose := events.Listener(func(...any) {
		p.OnError("poll connection closed prematurely")
	})

	ctx.Cleanup = func() {
		ctx.RemoveListener("close", onClose)
		p.mu_req.Lock()
		p.req = nil
		p.mu_req.Unlock()
	}

	ctx.On("close", onClose)

	p.mu_req.Lock()
	p.req = ctx
	p.mu_req.Unlock()

	p.SetWritable(true)
	p.Emit("drain")

	// if we're still writable but had a pending close, trigger an empty send
	if p.Writable() && p.shouldClose != nil {
		polling_log.Debug("triggering empty send to append close packet")
		p.Send([]*packet.Packet{
			&packet.Packet{
				Type: packet.NOOP,
			},
		})
	}
}

// The client sends a request with data.
func (p *polling) onDataRequest(ctx *types.HttpContext) {
	p.mu_dataCtx.RLock()
	if p.dataCtx != nil {
		defer p.mu_dataCtx.RUnlock()
		// assert: p.dataRes, '.dataReq and .dataRes should be (un)set together'
		p.OnError("data request overlap from client")
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.Write(nil)
		return
	}
	p.mu_dataCtx.RUnlock()

	isBinary := "application/octet-stream" == ctx.Headers().Peek("Content-Type")

	if isBinary && p.protocol == 4 {
		p.OnError("invalid content")
		return
	}

	p.mu_dataCtx.Lock()
	p.dataCtx = ctx
	p.mu_dataCtx.Unlock()

	var onClose events.Listener

	cleanup := func() {
		p.dataCtx.RemoveListener("close", onClose)
		p.mu_dataCtx.Lock()
		p.dataCtx = nil
		p.mu_dataCtx.Unlock()
	}

	onClose = func(...any) {
		cleanup()
		p.OnError("data request connection closed prematurely")
	}

	p.dataCtx.On("close", onClose)

	if p.dataCtx.Request().ContentLength > p.maxHttpBufferSize {
		p.dataCtx.SetStatusCode(http.StatusRequestEntityTooLarge)
		p.dataCtx.Write(nil)
		cleanup()
		return
	}

	var packet types.BufferInterface
	if isBinary {
		packet = types.NewBytesBuffer(nil)
	} else {
		packet = types.NewStringBuffer(nil)
	}
	if rc, ok := p.dataCtx.Request().Body.(io.ReadCloser); ok && rc != nil {
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
	var mu sync.Mutex
	for key, value := range p.Headers(p.dataCtx, headers) {
		mu.Lock()
		p.dataCtx.Response().Header().Set(key, value)
		mu.Unlock()
	}
	p.dataCtx.SetStatusCode(http.StatusOK)
	io.WriteString(p.dataCtx, "ok")
	cleanup()
}

// Processes the incoming data payload.
func (p *polling) PollingOnData(data types.BufferInterface) {
	polling_log.Debug(`received "%s"`, data)

	for _, packetData := range p.parser.DecodePayload(data) {
		if packet.CLOSE == packetData.Type {
			polling_log.Debug("got xhr close packet")
			p.OnClose()
			return
		}

		p.OnPacket(packetData)
	}
}

// Overrides onClose.
func (p *polling) PollingOnClose() {
	if p.Writable() {
		// close pending poll request
		p.Send([]*packet.Packet{
			&packet.Packet{
				Type: packet.NOOP,
			},
		})
	}
	p.TransportOnClose()
}

// Writes a packet payload.
func (p *polling) PollingSend(packets []*packet.Packet) {
	p.SetWritable(false)

	if p.shouldClose != nil {
		polling_log.Debug("appending close packet to payload")
		packets = append(packets, &packet.Packet{
			Type: packet.CLOSE,
		})
		p.shouldClose()
		p.shouldClose = nil
	}

	option := &packet.Options{false}
	for _, packetData := range packets {
		if packetData.Options != nil && packetData.Options.Compress {
			option.Compress = true
			break
		}
	}

	if p.protocol == 3 {
		data, _ := p.parser.EncodePayload(packets, p.supportsBinary)
		p.Write(data, option)
	} else {
		data, _ := p.parser.EncodePayload(packets)
		p.Write(data, option)
	}
}

// Writes data as response to poll request.
func (p *polling) Write(data types.BufferInterface, options *packet.Options) {
	polling_log.Debug(`writing "%s"`, data)
	p.DoWrite(data, options, func(ctx *types.HttpContext) { ctx.Cleanup() })
}

// Performs the write.
func (p *polling) PollingDoWrite(data types.BufferInterface, options *packet.Options, callback func(*types.HttpContext)) {
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
		p.mu_req.RLock()
		ctx := p.req
		p.mu_req.RUnlock()

		var mu sync.Mutex
		headers["Content-Length"] = length
		for key, value := range p.Headers(ctx, headers) {
			mu.Lock()
			ctx.Response().Header().Set(key, value)
			mu.Unlock()
		}
		ctx.SetStatusCode(http.StatusOK)
		io.Copy(ctx, data)
		callback(ctx)
	}

	if p.httpCompression == nil || options == nil || !options.Compress {
		respond(data, strconv.Itoa(data.Len()))
		return
	}

	if data.Len() < p.httpCompression.Threshold {
		respond(data, strconv.Itoa(data.Len()))
		return
	}

	encoding := utils.Contains(p.req.Headers().Peek("Accept-Encoding"), []string{"gzip", "deflate", "br"})
	if encoding == "" {
		respond(data, strconv.Itoa(data.Len()))
		return
	}

	if buf, err := p.compress(data, encoding); err == nil {
		headers["Content-Encoding"] = encoding
		respond(buf, strconv.Itoa(buf.Len()))
	}
}

// Compresses data.
func (p *polling) compress(data types.BufferInterface, encoding string) (types.BufferInterface, error) {
	polling_log.Debug("compressing")
	buf := types.NewBytesBuffer(nil)
	switch encoding {
	case "gzip":
		gz, err := gzip.NewWriterLevel(buf, 1)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		if _, err := io.Copy(gz, data); err != nil {
			return nil, err
		}
	case "deflate":
		fl, err := flate.NewWriter(buf, 1)
		if err != nil {
			return nil, err
		}
		defer fl.Close()
		if _, err := io.Copy(fl, data); err != nil {
			return nil, err
		}
	case "br":
		br := brotli.NewWriterLevel(buf, 1)
		defer br.Close()
		if _, err := io.Copy(br, data); err != nil {
			return nil, err
		}
	}
	return buf, nil
}

// Closes the transport.
func (p *polling) PollingDoClose(fn ...types.Callable) {
	polling_log.Debug("closing")

	p.mu_dataCtx.RLock()
	if p.dataCtx != nil && !p.dataCtx.IsWroteHeader() {
		polling_log.Debug("aborting ongoing data request")
		if h, ok := p.dataCtx.Response().(http.Hijacker); ok {
			if netConn, _, err := h.Hijack(); err == nil {
				if netConn.Close() == nil && !p.dataCtx.IsDone() {
					p.dataCtx.Flush()
				}
			}
		}
	}
	p.mu_dataCtx.RUnlock()

	var closeTimeoutTimer *utils.Timer = nil

	onClose := func() {
		utils.ClearTimeout(closeTimeoutTimer)
		if len(fn) > 0 {
			(fn[0])()
		}
		p.OnClose()
	}

	if p.Writable() {
		polling_log.Debug("transport writable - closing right away")
		p.Send([]*packet.Packet{
			&packet.Packet{
				Type: packet.CLOSE,
			},
		})
		onClose()
	} else if p.GetDiscarded() {
		polling_log.Debug("transport discarded - closing right away")
		onClose()
	} else {
		polling_log.Debug("transport not writable - buffering orderly close")
		p.shouldClose = onClose
		closeTimeoutTimer = utils.SetTimeOut(onClose, p.closeTimeout*time.Millisecond)
	}
}

// Returns headers for a response.
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
