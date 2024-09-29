package transports

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/zishang520/engine.io-go-parser/packet"
	_types "github.com/zishang520/engine.io-go-parser/types"
	"github.com/zishang520/engine.io/v2/events"
	"github.com/zishang520/engine.io/v2/log"
	"github.com/zishang520/engine.io/v2/types"
	"github.com/zishang520/engine.io/v2/utils"
)

var polling_log = log.NewLog("engine:polling")

type polling struct {
	Transport

	closeTimeout time.Duration

	req     atomic.Pointer[types.HttpContext]
	dataCtx atomic.Pointer[types.HttpContext]

	shouldClose atomic.Pointer[types.Callable]
	mu          sync.Mutex
}

// HTTP polling New.
func MakePolling() Polling {
	p := &polling{Transport: MakeTransport()}

	p.Prototype(p)

	return p
}

func NewPolling(ctx *types.HttpContext) Polling {
	p := MakePolling()

	p.Construct(ctx)

	return p
}

func (p *polling) Construct(ctx *types.HttpContext) {
	p.Transport.Construct(ctx)

	p.closeTimeout = 30 * 1000 * time.Millisecond
}

func (p *polling) Req() *types.HttpContext {
	return p.req.Load()
}

func (p *polling) SetReq(req *types.HttpContext) {
	p.req.Store(req)
}

func (p *polling) Name() string {
	return POLLING
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
	if p.Req() != nil {
		polling_log.Debug("request overlap")
		// assert: p.res, '.req should be (un)set together'
		p.OnError("overlap from client", nil)
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.Write(nil)
		return
	}

	polling_log.Debug("setting request")

	onClose := events.Listener(func(...any) {
		p.OnError("poll connection closed prematurely", nil)
	})

	p.SetReq(ctx)

	ctx.Cleanup = func() {
		ctx.RemoveListener("close", onClose)
		p.SetReq(nil)
	}

	ctx.Once("close", onClose)

	p.SetWritable(true)
	p.Emit("ready")

	// if we're still writable but had a pending close, trigger an empty send
	if p.Writable() && p.shouldClose.Load() != nil {
		polling_log.Debug("triggering empty send to append close packet")
		p.Send([]*packet.Packet{
			{
				Type: packet.NOOP,
			},
		})
	}
}

// The client sends a request with data.
func (p *polling) onDataRequest(ctx *types.HttpContext) {
	if p.dataCtx.Load() != nil {
		// assert: p.dataRes, '.dataCtx should be (un)set together'
		p.OnError("data request overlap from client", nil)
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.Write(nil)
		return
	}

	isBinary := "application/octet-stream" == ctx.Headers().Peek("Content-Type")

	if isBinary && p.Protocol() == 4 {
		p.OnError("invalid content", nil)
		return
	}

	p.dataCtx.Store(ctx)

	var onClose events.Listener

	cleanup := func() {
		ctx.RemoveListener("close", onClose)
		p.dataCtx.Store(nil)
	}

	onClose = func(...any) {
		cleanup()
		p.OnError("data request connection closed prematurely", nil)
	}

	ctx.Once("close", onClose)

	if ctx.Request().ContentLength > p.MaxHttpBufferSize() {
		ctx.SetStatusCode(http.StatusRequestEntityTooLarge)
		ctx.Write(nil)
		cleanup()
		return
	}

	var packet _types.BufferInterface
	if isBinary {
		packet = _types.NewBytesBuffer(nil)
	} else {
		packet = _types.NewStringBuffer(nil)
	}
	if rc, ok := ctx.Request().Body.(io.ReadCloser); ok && rc != nil {
		packet.ReadFrom(rc)
		rc.Close()
	}
	p.Proto().OnData(packet)

	headers := utils.NewParameterBag(map[string][]string{
		// text/html is required instead of text/plain to avoid an
		// unwanted download dialog on certain user-agents (GH-43)
		"Content-Type":   {"text/html"},
		"Content-Length": {"2"},
	})

	// After writing the data, close will be triggered, so it needs to be executed first.
	cleanup()

	// The following process in nodejs is asynchronous.
	ctx.ResponseHeaders.With(p.headers(ctx, headers).All())
	ctx.SetStatusCode(http.StatusOK)
	io.WriteString(ctx, "ok")
}

// Processes the incoming data payload.
func (p *polling) OnData(data _types.BufferInterface) {
	polling_log.Debug(`received "%s"`, data)

	packets, _ := p.Parser().DecodePayload(data)
	for _, packetData := range packets {
		if packet.CLOSE == packetData.Type {
			polling_log.Debug("got xhr close packet")
			p.OnClose()
			return
		}

		p.OnPacket(packetData)
	}
}

// Overrides onClose.
func (p *polling) OnClose() {
	if p.Writable() {
		// close pending poll request
		p.Send([]*packet.Packet{
			{
				Type: packet.NOOP,
			},
		})
	}
	p.Transport.OnClose()
}

// Writes a packet payload.
func (p *polling) Send(packets []*packet.Packet) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ctx := p.Req()

	if ctx == nil {
		return
	}

	p.SetWritable(false)
	if shouldClose := p.shouldClose.Load(); shouldClose != nil {
		polling_log.Debug("appending close packet to payload")
		packets = append(packets, &packet.Packet{
			Type: packet.CLOSE,
		})
		(*shouldClose)()
		p.shouldClose.Store(nil)
	}

	option := &packet.Options{Compress: false}
	for _, packetData := range packets {
		if packetData.Options != nil && packetData.Options.Compress {
			option.Compress = true
			break
		}
	}

	if p.Protocol() == 3 {
		data, _ := p.Parser().EncodePayload(packets, p.SupportsBinary())
		p.write(ctx, data, option)
	} else {
		data, _ := p.Parser().EncodePayload(packets)
		p.write(ctx, data, option)
	}
}

// Writes data as response to poll request.
func (p *polling) write(ctx *types.HttpContext, data _types.BufferInterface, options *packet.Options) {
	polling_log.Debug(`writing %#v`, data)
	// Assert that the prototype is Polling.
	p.Proto().(Polling).DoWrite(ctx, data, options, func(ctx *types.HttpContext) {
		ctx.Cleanup()
		p.Emit("drain")
	})
}

// Performs the write.
func (p *polling) DoWrite(ctx *types.HttpContext, data _types.BufferInterface, options *packet.Options, callback func(*types.HttpContext)) {
	contentType := "application/octet-stream"
	// explicit UTF-8 is required for pages not served under utf
	switch data.(type) {
	case *_types.StringBuffer:
		contentType = "text/plain; charset=UTF-8"
	}

	headers := utils.NewParameterBag(map[string][]string{
		"Content-Type": {contentType},
	})

	respond := func(data _types.BufferInterface, length string) {
		headers.Set("Content-Length", length)
		ctx.ResponseHeaders.With(p.headers(ctx, headers).All())
		ctx.SetStatusCode(http.StatusOK)
		io.Copy(ctx, data)
		callback(ctx)
	}

	if p.HttpCompression() == nil || options == nil || !options.Compress {
		respond(data, strconv.Itoa(data.Len()))
		return
	}

	if data.Len() < p.HttpCompression().Threshold {
		respond(data, strconv.Itoa(data.Len()))
		return
	}

	encoding := utils.Contains(ctx.Headers().Peek("Accept-Encoding"), []string{"gzip", "deflate", "br"})
	if encoding == "" {
		respond(data, strconv.Itoa(data.Len()))
		return
	}

	if buf, err := p.compress(data, encoding); err == nil {
		headers.Set("Content-Encoding", encoding)
		respond(buf, strconv.Itoa(buf.Len()))
	}
}

// Compresses data.
func (p *polling) compress(data _types.BufferInterface, encoding string) (_types.BufferInterface, error) {
	polling_log.Debug("compressing")
	buf := _types.NewBytesBuffer(nil)
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
func (p *polling) DoClose(fn types.Callable) {
	polling_log.Debug("closing")

	if dataCtx := p.dataCtx.Load(); dataCtx != nil && !dataCtx.IsDone() {
		polling_log.Debug("aborting ongoing data request")
		dataCtx.ResponseHeaders.Set("Connection", "close")
		dataCtx.SetStatusCode(http.StatusTooManyRequests)
		dataCtx.Write(nil)
	}

	onClose := func() {
		if fn != nil {
			fn()
		}
		p.OnClose()
	}

	if p.Writable() {
		polling_log.Debug("transport writable - closing right away")
		p.Send([]*packet.Packet{
			{
				Type: packet.CLOSE,
			},
		})
		onClose()
	} else if p.Discarded() {
		polling_log.Debug("transport discarded - closing right away")
		onClose()
	} else {
		polling_log.Debug("transport not writable - buffering orderly close")
		closeTimeoutTimer := utils.SetTimeout(onClose, p.closeTimeout)
		shouldClose := types.Callable(func() {
			utils.ClearTimeout(closeTimeoutTimer)
			onClose()
		})
		p.shouldClose.Store(&shouldClose)
	}
}

// Returns headers for a response.
func (p *polling) headers(ctx *types.HttpContext, headers *utils.ParameterBag) *utils.ParameterBag {
	// prevent XSS warnings on IE
	// https://github.com/socketio/socket.io/pull/1333
	if ua := ctx.UserAgent(); (len(ua) > 0) && ((strings.Index(ua, ";MSIE") > -1) || (strings.Index(ua, "Trident/") > -1)) {
		headers.Set("X-XSS-Protection", "0")
	}
	headers.Set("Cache-Control", "no-store")
	p.Emit("headers", headers, ctx)
	return headers
}
