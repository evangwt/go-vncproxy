# go-vncproxy

[![Go Report Card](https://goreportcard.com/badge/github.com/evangwt/go-vncproxy)](https://goreportcard.com/report/github.com/evangwt/go-vncproxy)[![GitHub release](https://img.shields.io/github/release/evangwt/go-vncproxy.svg)](https://github.com/evangwt/go-vncproxy/releases/)

A tiny vnc websocket proxy written by golang supports [noVNC](https://github.com/novnc/noVNC) client.

# Feature

 * Token handler: like [websockify](https://github.com/novnc/websockify), you can customlize the token handler to access multiple vnc backends by a single proxy instance.
 * Authentication: it depends on your vnc servers, since the proxy just copy the stream of both clients and servers.
 * **WebSocket Adapter Interface**: Abstraction layer for different WebSocket libraries, supporting `golang.org/x/net/websocket`, `github.com/gorilla/websocket`, `github.com/coder/websocket`, and custom implementations.

# Usage

## Basic Usage (Backward Compatible)

```go
package main

import (
	"net/http"

	"github.com/evangwt/go-vncproxy"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

func main() {
	r := gin.Default()

	vncProxy := NewVNCProxy()
	r.GET("/ws", func(ctx *gin.Context) {
		h := websocket.Handler(vncProxy.ServeWS)
		h.ServeHTTP(ctx.Writer, ctx.Request)
	})

	if err := r.Run(); err != nil {
		panic(err)
	}
}

func NewVNCProxy() *vncproxy.Proxy {
	return vncproxy.New(&vncproxy.Config{
		LogLevel: vncproxy.DebugLevel,
		TokenHandler: func(r *http.Request) (addr string, err error) {
			// validate token and get forward vnc addr
			// ...
			addr = ":5901"
			return
		},
	})
}
```

## Using WebSocket Adapter Interface

The new adapter interface allows you to use different WebSocket libraries:

```go
package main

import (
	"net/http"

	"github.com/evangwt/go-vncproxy"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Using the new adapter interface with default adapter
	vncProxy := NewVNCProxyWithAdapter()
	r.GET("/ws", func(ctx *gin.Context) {
		handler := vncProxy.HTTPHandler()
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	})

	if err := r.Run(); err != nil {
		panic(err)
	}
}

func NewVNCProxyWithAdapter() *vncproxy.Proxy {
	return vncproxy.New(&vncproxy.Config{
		LogLevel: vncproxy.DebugLevel,
		TokenHandler: func(r *http.Request) (addr string, err error) {
			return ":5901", nil
		},
		// WebSocketAdapter: customAdapter, // Optional: inject custom adapter
	})
}
```

## Custom WebSocket Adapter

You can implement your own WebSocket adapter for different libraries:

```go
// Example adapter for github.com/gorilla/websocket
type GorillaWebSocketAdapter struct {
	conn    *websocket.Conn
	request *http.Request
}

func (a *GorillaWebSocketAdapter) Read(p []byte) (n int, err error) {
	_, data, err := a.conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	n = copy(p, data)
	return n, nil
}

func (a *GorillaWebSocketAdapter) Write(p []byte) (n int, err error) {
	err = a.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (a *GorillaWebSocketAdapter) Close() error {
	return a.conn.Close()
}

func (a *GorillaWebSocketAdapter) RemoteAddr() string {
	return a.conn.RemoteAddr().String()
}

func (a *GorillaWebSocketAdapter) Request() *http.Request {
	return a.request
}

func (a *GorillaWebSocketAdapter) SetBinaryMode() error {
	// Gorilla WebSocket handles message types per message
	return nil
}

// Create an upgrader that implements WebSocketUpgrader interface
type GorillaWebSocketUpgrader struct {
	upgrader websocket.Upgrader
}

func (u *GorillaWebSocketUpgrader) Upgrade(w http.ResponseWriter, r *http.Request) (vncproxy.WebSocketAdapter, error) {
	conn, err := u.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &GorillaWebSocketAdapter{conn: conn, request: r}, nil
}

// Use with VNC proxy
func NewVNCProxyWithGorilla() *vncproxy.Proxy {
	upgrader := &GorillaWebSocketUpgrader{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	
	return vncproxy.New(&vncproxy.Config{
		LogLevel: vncproxy.DebugLevel,
		TokenHandler: func(r *http.Request) (addr string, err error) {
			return ":5901", nil
		},
		WebSocketAdapter: upgrader,
	})
}
```

## WebSocket Adapter Interface

The `WebSocketAdapter` interface provides a unified way to work with different WebSocket libraries:

```go
type WebSocketAdapter interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
	RemoteAddr() string
	Request() *http.Request
	SetBinaryMode() error
}

type WebSocketUpgrader interface {
	Upgrade(w http.ResponseWriter, r *http.Request) (WebSocketAdapter, error)
}
```

This allows go-vncproxy to work with:
- `golang.org/x/net/websocket` (default)
- `github.com/gorilla/websocket`
- `github.com/coder/websocket`
- Any custom WebSocket implementation
