//go:build ignore

package vncproxy

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
)

// GorillaWebSocketAdapter adapts github.com/gorilla/websocket.Conn to the WebSocketAdapter interface
type GorillaWebSocketAdapter struct {
	conn    *websocket.Conn
	request *http.Request
	ctx     context.Context
}

// NewGorillaWebSocketAdapter creates a new adapter for github.com/gorilla/websocket.Conn
func NewGorillaWebSocketAdapter(conn *websocket.Conn, req *http.Request, ctx context.Context) *GorillaWebSocketAdapter {
	return &GorillaWebSocketAdapter{
		conn:    conn,
		request: req,
		ctx:     ctx,
	}
}

// Read implements WebSocketAdapter
func (a *GorillaWebSocketAdapter) Read(p []byte) (n int, err error) {
	// Gorilla WebSocket uses message-based reading, so we need to handle this differently
	messageType, data, err := a.conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	
	// For VNC, we expect binary messages
	if messageType != websocket.BinaryMessage {
		// Convert text to binary if needed
	}
	
	n = copy(p, data)
	return n, nil
}

// Write implements WebSocketAdapter
func (a *GorillaWebSocketAdapter) Write(p []byte) (n int, err error) {
	err = a.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close implements WebSocketAdapter
func (a *GorillaWebSocketAdapter) Close() error {
	return a.conn.Close()
}

// RemoteAddr implements WebSocketAdapter
func (a *GorillaWebSocketAdapter) RemoteAddr() string {
	return a.conn.RemoteAddr().String()
}

// Request implements WebSocketAdapter
func (a *GorillaWebSocketAdapter) Request() *http.Request {
	return a.request
}

// SetBinaryMode implements WebSocketAdapter
func (a *GorillaWebSocketAdapter) SetBinaryMode() error {
	// Gorilla WebSocket handles message types per message, not globally
	// So this is essentially a no-op
	return nil
}

// GorillaWebSocketUpgrader upgrades HTTP connections to WebSocket using github.com/gorilla/websocket
type GorillaWebSocketUpgrader struct {
	upgrader websocket.Upgrader
	handler  WebSocketHandler
}

// NewGorillaWebSocketUpgrader creates a new upgrader for github.com/gorilla/websocket
func NewGorillaWebSocketUpgrader(handler WebSocketHandler, upgrader websocket.Upgrader) *GorillaWebSocketUpgrader {
	return &GorillaWebSocketUpgrader{
		upgrader: upgrader,
		handler:  handler,
	}
}

// Upgrade implements WebSocketUpgrader
func (u *GorillaWebSocketUpgrader) Upgrade(w http.ResponseWriter, r *http.Request) (WebSocketAdapter, error) {
	conn, err := u.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	
	adapter := NewGorillaWebSocketAdapter(conn, r, r.Context())
	return adapter, nil
}

/*
Example usage with Gorilla WebSocket:

```go
package main

import (
	"net/http"

	"github.com/evangwt/go-vncproxy"
	"github.com/gorilla/websocket"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Create VNC proxy with Gorilla WebSocket adapter
	vncProxy := NewVNCProxyWithGorilla()
	
	r.GET("/ws", func(ctx *gin.Context) {
		handler := vncProxy.HTTPHandler()
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	})

	if err := r.Run(); err != nil {
		panic(err)
	}
}

func NewVNCProxyWithGorilla() *vncproxy.Proxy {
	// Create Gorilla WebSocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for demo
		},
	}
	
	// Create VNC proxy with Gorilla adapter
	proxy := vncproxy.New(&vncproxy.Config{
		LogLevel: vncproxy.DebugLevel,
		TokenHandler: func(r *http.Request) (addr string, err error) {
			return ":5901", nil
		},
	})
	
	// Create Gorilla WebSocket upgrader adapter
	wsUpgrader := vncproxy.NewGorillaWebSocketUpgrader(proxy.Handler(), upgrader)
	
	// Create new proxy with the custom adapter
	return vncproxy.New(&vncproxy.Config{
		LogLevel: vncproxy.DebugLevel,
		TokenHandler: func(r *http.Request) (addr string, err error) {
			return ":5901", nil
		},
		WebSocketAdapter: wsUpgrader,
	})
}
```
*/