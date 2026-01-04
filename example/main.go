package main

import (
	"net/http"

	"github.com/evangwt/go-vncproxy"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

func main() {
	r := gin.Default()

	// Example 1: Traditional usage (backward compatible)
	vncProxy := NewVNCProxy()
	r.GET("/ws", func(ctx *gin.Context) {
		h := websocket.Handler(vncProxy.ServeWS)
		h.ServeHTTP(ctx.Writer, ctx.Request)
	})
	
	// Example 2: Using the new adapter interface with default adapter
	vncProxyAdapter := NewVNCProxyWithAdapter()
	r.GET("/ws-adapter", func(ctx *gin.Context) {
		handler := vncProxyAdapter.HTTPHandler()
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	})

	if err := r.Run(); err != nil {
		panic(err)
	}
}

// Traditional approach - backward compatible
func NewVNCProxy() *vncproxy.Proxy {
	return vncproxy.New(&vncproxy.Config{
		LogLevel: vncproxy.DebugLevel,
		// Logger: customerLogger,	// inject a custom logger
		// DialTimeout: 10 * time.Second, // customer DialTimeout
		TokenHandler: func(r *http.Request) (addr string, err error) {
			return ":5901", nil
		},
	})
}

// New approach using adapter interface
func NewVNCProxyWithAdapter() *vncproxy.Proxy {
	return vncproxy.New(&vncproxy.Config{
		LogLevel: vncproxy.DebugLevel,
		TokenHandler: func(r *http.Request) (addr string, err error) {
			return ":5901", nil
		},
		// WebSocketAdapter: customAdapter, // You can inject a custom adapter here
	})
}
