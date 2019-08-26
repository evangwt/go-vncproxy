# go-vncproxy
A tiny vnc websocket proxy written by golang supports [noVNC](https://github.com/novnc/noVNC) client.

# Feature

 * Token handler: like [websockify](https://github.com/novnc/websockify), you can customlize the token handler to multiple vnc backend by a single proxy instance.
 * Authentication: it depends on your vnc servers, since the proxy just copy the stream of both clients and servers.

# Usage
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
			return ":5901", nil
		},
	})
}
```
