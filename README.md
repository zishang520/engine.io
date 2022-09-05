# Engine.IO: the realtime engine for golang

[![Build Status](https://github.com/zishang520/engine.io/workflows/Go/badge.svg?branch=master)](https://github.com/zishang520/engine.io/actions)
[![GoDoc](https://pkg.go.dev/badge/github.com/zishang520/engine.io?utm_source=godoc)](https://pkg.go.dev/github.com/zishang520/engine.io)

`Engine.IO` is the implementation of transport-based
cross-browser/cross-device bi-directional communication layer for
[Socket.IO for golang](http://github.com/zishang520/socket.io).

## How to use

### Server
If you need to print the debug log, please set the environment variable `DEBUG=*`

#### (A) Listening on a port

```go
package main

import (
    "os"
    "os/signal"
    "strings"
    "syscall"

    "github.com/zishang520/engine.io/config"
    "github.com/zishang520/engine.io/engine"
    "github.com/zishang520/engine.io/types"
    "github.com/zishang520/engine.io/utils"
)

func main() {
    serverOptions := &config.ServerOptions{}
    serverOptions.SetAllowEIO3(true)
    serverOptions.SetCors(&types.Cors{
        Origin:      "*",
        Credentials: true,
    })

    engineServer := engine.Listen("127.0.0.1:4444", serverOptions, nil)
    engineServer.On("connection", func(sockets ...any) {
        socket := sockets[0].(engine.Socket)
        socket.Send(strings.NewReader("utf 8 string"), nil, nil)
        socket.Send(types.NewBytesBuffer([]byte{0, 1, 2, 3, 4, 5}), nil, nil)
        socket.Send(types.NewBytesBufferString("BufferString by string"), nil, nil)
        socket.Send(types.NewStringBuffer([]byte("StringBuffer by byte")), nil, nil)
        socket.Send(types.NewStringBufferString("StringBuffer by string"), nil, nil)
        socket.On("message", func(...any) {
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
                return
            }
        }
    }()

    <-exit
    os.Exit(0)
}

```

#### (B) Intercepting requests for a `*types.HttpServer`

```go
package main

import (
    "os"
    "os/signal"
    "syscall"

    "github.com/zishang520/engine.io/config"
    "github.com/zishang520/engine.io/engine"
    "github.com/zishang520/engine.io/types"
    "github.com/zishang520/engine.io/utils"
)

func main() {
    serverOptions := &config.ServerOptions{}
    serverOptions.SetAllowEIO3(true)
    serverOptions.SetCors(&types.Cors{
        Origin:      "*",
        Credentials: true,
    })

    http := types.CreateServer(nil).Listen("127.0.0.1:4444", nil)

    engineServer := engine.Attach(http, serverOptions)

    engineServer.On("connection", func(sockets ...any) {
        socket := sockets[0].(engine.Socket)
        socket.On("message", func(...any) {
        })
        socket.On("close", func(...any) {
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
                return
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
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "github.com/gorilla/websocket"
    "github.com/zishang520/engine.io/config"
    "github.com/zishang520/engine.io/engine"
    "github.com/zishang520/engine.io/types"
    "github.com/zishang520/engine.io/utils"
)

func main() {
    serverOptions := &config.ServerOptions{}
    serverOptions.SetAllowEIO3(true)
    serverOptions.SetCors(&types.Cors{
        Origin:      "*",
        Credentials: true,
    })

    httpServer := types.CreateServer(nil).Listen("127.0.0.1:4444", nil)

    engineServer := engine.New(serverOptions)

    httpServer.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if !websocket.IsWebSocketUpgrade(r) {
            engineServer.HandleRequest(types.NewHttpContext(w, r))
        } else if engineServer.Opts().Transports().Has("websocket") {
            engineServer.HandleUpgrade(types.NewHttpContext(w, r))
        } else {
            httpServer.DefaultHandler.ServeHTTP(w, r)
        }
    })

    engineServer.On("connection", func(sockets ...any) {
        socket := sockets[0].(engine.Socket)
        socket.On("message", func(...any) {
        })
        socket.On("close", func(...any) {
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
                return
            }
        }
    }()

    <-exit
    httpServer.Close(nil)
    os.Exit(0)
}
```

#### (D) Passing in requests (http.Handler interface)

```go
package main

import (
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "github.com/zishang520/engine.io/config"
    "github.com/zishang520/engine.io/engine"
    "github.com/zishang520/engine.io/types"
    "github.com/zishang520/engine.io/utils"
)

func main() {
    serverOptions := &config.ServerOptions{}
    serverOptions.SetAllowEIO3(true)
    serverOptions.SetCors(&types.Cors{
        Origin:      "*",
        Credentials: true,
    })

    engineServer := engine.New(serverOptions)

    engineServer.On("connection", func(sockets ...any) {
        socket := sockets[0].(engine.Socket)
        socket.On("message", func(...any) {
        })
        socket.On("close", func(...any) {
            utils.Log().Println("client close.")
        })
    })

    http.Handle("/engine.io/", engineServer)
    go http.ListenAndServe(":8090", nil)

    utils.Log().Println("%v", engineServer)

    exit := make(chan struct{})
    SignalC := make(chan os.Signal)

    signal.Notify(SignalC, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
    go func() {
        for s := range SignalC {
            switch s {
            case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
                close(exit)
                return
            }
        }
    }()

    <-exit

    // Need to handle server shutdown disconnecting client connections.
    engineServer.Close()
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

These are exposed by `import "github.com/zishang520/engine.io/engine"`:

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
- `Server`: Server struct
- `Socket`: Socket struct

##### Methods

- `New`
    - Returns a new `Server` instance. If the first argument is an `*types.HttpServer`() then the
      new `Server` instance will be attached to it. Otherwise, the arguments are passed
      directly to the `Server` constructor.
    - **Parameters**
      - `*types.HttpServer`: optional, server to attach to.
      - `any`: optional, options object (see `*config.Server#constructor` api docs below)

  The following are identical ways to instantiate a server and then attach it.

```go
import "github.com/zishang520/engine.io/config"
import "github.com/zishang520/engine.io/engine"
import "github.com/zishang520/engine.io/types"

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

- `Listen`
    - Creates an `*types.HttpServer` which listens on the given port and attaches WS
      to it. It returns `501 Not Implemented` for regular http requests.
    - **Parameters**
      - `string`: address to listen on.
      - `any`: optional, options object
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

server.On('connection', func(...any) {});
```

- `Attach`
    - Captures `upgrade` requests for a `*types.HttpServer`. In other words, makes
      a regular `*types.HttpServer` WebSocket-compatible.
    - **Parameters**
      - `*types.HttpServer`: server to attach to.
      - `any`: `config.ServerOptionsInterface`: can be nil, interface config.ServerOptionsInterface or config.AttachOptionsInterface
    - **Options**
      - All options from `engine.Server.attach` method, documented below.
      - **Additionally** See Server `New` below for options you can pass for creating the new Server
    - **Returns** `engine.Server` a new Server instance.

#### Server

The main server/manager. _Inherits from events.EventEmitter_.

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
        - `context` (`map[string]any`): extra info about the error

| Code | Message |
| ---- | ------- |
| -1 | "Ok"
| 0 | "Transport unknown"
| 1 | "Session ID unknown"
| 2 | "Bad handshake method"
| 3 | "Bad request"
| 4 | "Forbidden"
| 5 | "Unsupported protocol version"

##### Properties

**Important**: if you plan to use Engine.IO in a scalable way, please
keep in mind the properties below will only reflect the clients connected
to a single process.

- `Clients()` _(*sync.Map)_: hash of connected clients by id.
- `ClientsCount()` _(uint64)_: number of connected clients.

##### Methods

- **New**
    - Initializes the server
    - **Parameters**
      - `config.ServerOptionsInterface`: can be nil, interface config.ServerOptionsInterface
    - **Options**
      - `SetPingTimeout(time.Duration)`: how many ms without a pong packet to
        consider the connection closed (`20000 * time.Millisecond`)
      - `SetPingInterval(time.Duration)`: how many ms before sending a new ping
        packet (`25000 * time.Millisecond`)
      - `SetUpgradeTimeout(time.Duration)`: how many ms before an uncompleted transport upgrade is cancelled (`10000 * time.Millisecond`)
      - `SetMaxHttpBufferSize(int64)`: how many bytes or characters a message
        can be, before closing the session (to avoid DoS). Default
        value is `1E6`.
      - `SetAllowRequest(config.AllowRequest)`: A function that receives a given handshake or upgrade request as its first argument and can decide whether to continue. error is not empty to indicate that the request was rejected.
      - `SetTransports(*types.Set[string])`: transports to allow connections
        to (`['polling', 'websocket']`)
      - `SetAllowUpgrades(bool)`: whether to allow transport upgrades
        (`true`)
      - `SetPerMessageDeflate(*types.PerMessageDeflate)`: parameters of the WebSocket permessage-deflate extension
        - `Threshold` (`int`): data is compressed only if the byte size is above this value (`1024`)
      - `SetHttpCompression(*types.HttpCompression)`: parameters of the http compression for the polling transports
        - `Threshold` (`int`): data is compressed only if the byte size is above this value (`1024`)
      - `SetCookie(*http.Cookie)`: configuration of the cookie that
        contains the client sid to send as part of handshake response
        headers. This cookie might be used for sticky-session. Defaults to not sending any cookie (`nil`).
      - `SetCors(*types.Cors)`: the options that will be forwarded to the cors module. See [there](https://pkg.go.dev/github.com/zishang520/engine.io/types#Cors) for all available options. Defaults to no CORS allowed.
      - `SetInitialPacket(io.Reader)`: an optional packet which will be concatenated to the handshake packet emitted by Engine.IO.
      - `SetAllowEIO3(bool)`: whether to support v3 Engine.IO clients (defaults to `false`)
- `Close`
    - Closes all clients
    - **Returns** `engine.Server` for chaining
- `HandleRequest`
    - Called internally when a `Engine` request is intercepted.
    - **Parameters**
      - `*types.HttpContext`: a node request context
- `HandleUpgrade`
    - Called internally when a `Engine` ws upgrade is intercepted.
    - **Parameters**
      - `*types.HttpContext`: a node request context
- `Attach`
    - Attach this Server instance to an `*types.HttpServer`
    - Captures `upgrade` requests for a `*types.HttpServer`. In other words, makes
      a regular *types.HttpServer WebSocket-compatible.
    - **Parameters**
      - `*types.HttpServer`: server to attach to.
      - `any`: can be nil, interface config.AttachOptionsInterface
    - **Options**
      - `SetPath(string)`: name of the path to capture (`/engine.io`).
      - `SetDestroyUpgrade(bool)`: destroy unhandled upgrade requests (`true`)
      - `SetDestroyUpgradeTimeout(time.Duration)`: milliseconds after which unhandled requests are ended (`1000 * time.Millisecond`)
- `GenerateId`
    - Generate a socket id.
    - Overwrite this method to generate your custom socket id.
    - **Parameters**
      - `*types.HttpContext`: a node request context
  - **Returns** A socket id for connected client.

<hr><br>

...
