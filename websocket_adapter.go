package vncproxy

import (
	"context"
	"net/http"
)

// WebSocketAdapter provides an abstraction layer for different WebSocket libraries.
// This interface allows go-vncproxy to work with various WebSocket implementations
// such as golang.org/x/net/websocket, github.com/gorilla/websocket, or github.com/coder/websocket.
type WebSocketAdapter interface {
	// Read reads data from the WebSocket connection
	Read(p []byte) (n int, err error)
	
	// Write writes data to the WebSocket connection
	Write(p []byte) (n int, err error)
	
	// Close closes the WebSocket connection
	Close() error
	
	// RemoteAddr returns the remote network address
	RemoteAddr() string
	
	// Request returns the HTTP request that initiated the WebSocket connection
	Request() *http.Request
	
	// SetBinaryMode sets the WebSocket to handle binary frames
	SetBinaryMode() error
}

// WebSocketUpgrader defines how to upgrade an HTTP connection to a WebSocket
type WebSocketUpgrader interface {
	// Upgrade upgrades an HTTP connection to a WebSocket connection
	Upgrade(w http.ResponseWriter, r *http.Request) (WebSocketAdapter, error)
}

// WebSocketHandler is a function that handles WebSocket connections using the adapter interface
type WebSocketHandler func(adapter WebSocketAdapter)

// AdapterConfig allows configuration of WebSocket adapters
type AdapterConfig struct {
	// Context for cancellation support
	Context context.Context
	
	// Additional adapter-specific configuration can be added here
}