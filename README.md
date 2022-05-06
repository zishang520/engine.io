
# Engine.IO: the realtime engine by golang

`Engine.IO` is the implementation of transport-based
cross-browser/cross-device bi-directional communication layer for
[Socket.IO for golang](http://github.com/zishang520/socket.io).

## How to use

### Server

#### (A) Listening on a port

```go
package main

import (
    "github.com/zishang520/engine.io/config"
    "github.com/zishang520/engine.io/engine"
    "github.com/zishang520/engine.io/types"
    "github.com/zishang520/engine.io/utils"
    "os"
    "os/signal"
    "strings"
    "syscall"
)

func main() {
    utils.Log().DEBUG = true

    serverOptions := &config.ServerOptions{}
    serverOptions.SetAllowEIO3(true)
    serverOptions.SetCors(&types.Cors{
        Origin:      "*",
        Credentials: true,
    })

    engineServer := engine.Listen("127.0.0.1:4444", serverOptions, nil)
    engineServer.On("connection", func(sockets ...interface{}) {
        socket := sockets[0].(engine.Socket)
        socket.Send(strings.NewReader("utf 8 string"), nil, nil)
        socket.Send(types.NewBytesBuffer([]byte{0, 1, 2, 3, 4, 5}), nil, nil)
        socket.Send(types.NewBytesBufferString("BufferString by string"), nil, nil)
        socket.Send(types.NewStringBuffer([]byte("StringBuffer by byte")), nil, nil)
        socket.Send(types.NewStringBufferString("StringBuffer by string"), nil, nil)
        socket.On("message", func(...interface{}) {
            // socket.Send(strings.NewReader("utf 8 string"), nil, nil)
        })
    })
    utils.Log().Println("%v", engineServer)

    exit := make(chan struct{})
    SignalC := make(chan os.Signal)

    signal.Notify(SignalC, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
    go func() {
        for s := range SignalC {
            switch s {
            case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
                close(exit)
            }
        }
    }()

    <-exit
    os.Exit(0)
}

```

#### (B) Intercepting requests for a types.HttpServer

```go
package main

import (
    "github.com/zishang520/engine.io/config"
    "github.com/zishang520/engine.io/engine"
    "github.com/zishang520/engine.io/types"
    "github.com/zishang520/engine.io/utils"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    utils.Log().DEBUG = true

    serverOptions := &config.ServerOptions{}
    serverOptions.SetAllowEIO3(true)
    serverOptions.SetCors(&types.Cors{
        Origin:      "*",
        Credentials: true,
    })

    http := types.CreateServer(nil).Listen("127.0.0.1:4444", nil)

    engineServer := engine.Attach(http, serverOptions)

    engineServer.On("connection", func(sockets ...interface{}) {
        socket := sockets[0].(engine.Socket)
        socket.On("message", func(...interface{}) {
        })
        socket.On("close", func(...interface{}) {
            utils.Log().Println("client close.")
        })
    })
    utils.Log().Println("%v", engineServer)

    exit := make(chan struct{})
    SignalC := make(chan os.Signal)

    signal.Notify(SignalC, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
    go func() {
        for s := range SignalC {
            switch s {
            case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
                close(exit)
            }
        }
    }()

    <-exit
    os.Exit(0)
}

```

#### (C) Passing in requests

```go
package main

import (
    "github.com/gorilla/websocket"
    "github.com/zishang520/engine.io/config"
    "github.com/zishang520/engine.io/engine"
    "github.com/zishang520/engine.io/types"
    "github.com/zishang520/engine.io/utils"
    "net/http"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    utils.Log().DEBUG = true

    serverOptions := &config.ServerOptions{}
    serverOptions.SetAllowEIO3(true)
    serverOptions.SetCors(&types.Cors{
        Origin:      "*",
        Credentials: true,
    })

    httpServer := types.CreateServer(nil).Listen("127.0.0.1:4444", nil)

    engineServer := engine.New(serverOptions)

    httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := types.NewHttpContext(w, r)
        if !websocket.IsWebSocketUpgrade(r) {
            engineServer.HandleRequest(ctx)
        } else {
            engineServer.HandleUpgrade(ctx)
        }
    })

    engineServer.On("connection", func(sockets ...interface{}) {
        socket := sockets[0].(engine.Socket)
        socket.On("message", func(...interface{}) {
        })
        socket.On("close", func(...interface{}) {
            utils.Log().Println("client close.")
        })
    })
    utils.Log().Println("%v", engineServer)

    exit := make(chan struct{})
    SignalC := make(chan os.Signal)

    signal.Notify(SignalC, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
    go func() {
        for s := range SignalC {
            switch s {
            case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
                close(exit)
            }
        }
    }()

    <-exit
    os.Exit(0)
}
```

### Client

```html
<script src="engine.io.js"></script>
<script>
  const socket = new eio.Socket('ws://localhost:4444');
  socket.on('open', () => {
    socket.on('message', data => {});
    socket.on('close', () => {});
  });
</script>
```

For more information on the client refer to the
[engine-client](http://github.com/socketio/engine.io-client) repository.

## What features does it have?

- **Support engine-client 3+**

## API

### Server

<hr><br>

#### Top-level

These are exposed by `import github.com/zishang520/engine.io/engine"`:

##### Events

- `flush`
    - Called when a socket buffer is being flushed.
    - **Arguments**
      - `engine.Socket`: socket being flushed
      - `[]*packet.Packet`: write buffer
- `drain`
    - Called when a socket buffer is drained
    - **Arguments**
      - `engine.Socket`: socket being flushed

##### Properties

- `Protocol` _(int)_: protocol revision number
- `Server`: Server class constructor
- `Socket`: Socket class constructor

##### Methods

- `()`
    - Returns a new `Server` instance. If the first argument is an `types.HttpServer` then the
      new `Server` instance will be attached to it. Otherwise, the arguments are passed
      directly to the `Server` constructor.
    - **Parameters**
      - `*types.HttpServer`: optional, server to attach to.
      - `interface{}`: optional, options object (see `*config.Server#constructor` api docs below)

  The following are identical ways to instantiate a server and then attach it.

```go
import github.com/zishang520/engine.io/engine"
import github.com/zishang520/engine.io/config"

var httpServer *types.HttpServer // previously created with `types.CreateServer(nil);`.
var eioServer engine.Server

// create a server first, and then attach
eioServer = engine.NewServer(nil)
eioServer.Attach(httpServer)

// or call the module as a function to get `Server`
eioServer = engine.New(nil)
eioServer.Attach(httpServer)

// immediately attach
eioServer = engine.New(httpServer)

// with custom options
c := &config.ServerOptions{}
c.SetMaxHttpBufferSize(1e3)
eioServer = engine.New(httpServer, c)

```

- `listen`
    - Creates an `types.HttpServer` which listens on the given port and attaches WS
      to it. It returns `501 Not Implemented` for regular http requests.
    - **Parameters**
      - `string`: address to listen on.
      - `interface{}`: optional, options object
      - `func()`: callback for `listen`.
    - **Options**
      - All options from `engine.Server.Attach` method, documented below.
      - **Additionally** See Server `New` below for options you can pass for creating the new Server
    - **Returns** `engine.Server`

```go
import "github.com/zishang520/engine.io/engine"
import "github.com/zishang520/engine.io/config"

c := &config.ServerOptions{}
c.SetPingTimeout(2000)
c.SetPingInterval(10000)

const server = engine.Listen("127.0.0.1:3000", c);

server.On('connection', func(...interface{}) {});
```

- `Attach`
    - Captures `upgrade` requests for a `types.HttpServer`. In other words, makes
      a regular http.Server WebSocket-compatible.
    - **Parameters**
      - `*types.HttpServer`: server to attach to.
      - `interface{}`: optional, options object
    - **Options**
      - All options from `engine.Server.attach` method, documented below.
      - **Additionally** See Server `New` below for options you can pass for creating the new Server
    - **Returns** `engine.Server` a new Server instance.

#### Server

The main server/manager. _Inherits from EventEmitter_.

##### Events

- `connection`
    - Fired when a new connection is established.
    - **Arguments**
      - `engine.Socket`: a Socket object

- `initial_headers`
    - Fired on the first request of the connection, before writing the response headers
    - **Arguments**
      - `headers` (`map[string]string`): a hash of headers
      - `ctx` (`*types.HttpContext`): the request

- `headers`
    - Fired on the all requests of the connection, before writing the response headers
    - **Arguments**
      - `headers` (`map[string]string`): a hash of headers
      - `ctx` (`*types.HttpContext`): the request

- `connection_error`
    - Fired when an error occurs when establishing the connection.
    - **Arguments**
      - `types.ErrorMessage`: an object with following properties:
        - `req` (`*types.HttpContext`): the request that was dropped
        - `code` (`int`): one of `Server.errors`
        - `message` (`string`): one of `Server.errorMessages`
        - `context` (`map[string]interface{}`): extra info about the error

| Code | Message |
| ---- | ------- |
| -1 | "Ok"
| 0 | "Transport unknown"
| 1 | "Session ID unknown"
| 2 | "Bad handshake method"
| 3 | "Bad request"
| 4 | "Forbidden"
| 5 | "Unsupported protocol version"

##### Methods

- `Clients()` _(*sync.Map)_: hash of connected clients by id.
- `ClientsCount()` _(uint64)_: number of connected clients.

...
