package vncproxy

import (
	"net/http"

	"golang.org/x/net/websocket"
)

// NetWebSocketAdapter adapts golang.org/x/net/websocket.Conn to the WebSocketAdapter interface
type NetWebSocketAdapter struct {
	conn *websocket.Conn
}

// NewNetWebSocketAdapter creates a new adapter for golang.org/x/net/websocket.Conn
func NewNetWebSocketAdapter(conn *websocket.Conn) *NetWebSocketAdapter {
	return &NetWebSocketAdapter{
		conn: conn,
	}
}

// Read implements WebSocketAdapter
func (a *NetWebSocketAdapter) Read(p []byte) (n int, err error) {
	return a.conn.Read(p)
}

// Write implements WebSocketAdapter
func (a *NetWebSocketAdapter) Write(p []byte) (n int, err error) {
	return a.conn.Write(p)
}

// Close implements WebSocketAdapter
func (a *NetWebSocketAdapter) Close() error {
	return a.conn.Close()
}

// RemoteAddr implements WebSocketAdapter
func (a *NetWebSocketAdapter) RemoteAddr() string {
	if a.conn.RemoteAddr() != nil {
		return a.conn.RemoteAddr().String()
	}
	return ""
}

// Request implements WebSocketAdapter
func (a *NetWebSocketAdapter) Request() *http.Request {
	return a.conn.Request()
}

// SetBinaryMode implements WebSocketAdapter
func (a *NetWebSocketAdapter) SetBinaryMode() error {
	a.conn.PayloadType = websocket.BinaryFrame
	return nil
}

// NetWebSocketUpgrader upgrades HTTP connections to WebSocket using golang.org/x/net/websocket
type NetWebSocketUpgrader struct {
	handler websocket.Handler
}

// NewNetWebSocketUpgrader creates a new upgrader for golang.org/x/net/websocket
func NewNetWebSocketUpgrader(handler WebSocketHandler) *NetWebSocketUpgrader {
	wsHandler := func(ws *websocket.Conn) {
		adapter := NewNetWebSocketAdapter(ws)
		handler(adapter)
	}
	
	return &NetWebSocketUpgrader{
		handler: websocket.Handler(wsHandler),
	}
}

// Upgrade implements WebSocketUpgrader
func (u *NetWebSocketUpgrader) Upgrade(w http.ResponseWriter, r *http.Request) (WebSocketAdapter, error) {
	// This method is used for direct upgrade scenarios
	// For the handler pattern, use ServeHTTP on the underlying handler
	panic("NetWebSocketUpgrader.Upgrade not implemented - use ServeHTTP on the handler instead")
}

// ServeHTTP serves WebSocket connections using the golang.org/x/net/websocket handler
func (u *NetWebSocketUpgrader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u.handler.ServeHTTP(w, r)
}