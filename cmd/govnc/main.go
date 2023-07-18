package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/codio/go-vncproxy/pkg/go-vncproxy"

	"github.com/akamensky/argparse"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

var indexHTML string

func main() {
	parser := argparse.NewParser("govnc", "VNCProxy for novnc")
	vncPort := parser.Int("s", "vncport", &argparse.Options{Required: false, Default: 5900, Help: "VNC port"})
	port := parser.Int("p", "port", &argparse.Options{Required: false, Default: 8080, Help: "VNC port"})
	version := parser.String("v", "version", &argparse.Options{Required: false, Default: "v1.0.0-a.7-dev", Help: "NOVNC client version"})
	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Print(parser.Usage(err))
		panic(err)
	}
	indexHTML = loadIndex(*version)
	runProxy(*port, *vncPort)
}

func loadIndex(version string) string {
	resp, err := http.Get(fmt.Sprintf(`https://static-assets.codio.com/noVNC/%s/vnc.html`, version))
	if err != nil {
		panic(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(body)
}

func runProxy(port int, vncPort int) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	vncProxy := NewVNCProxy(vncPort)
	router.GET("/websockify", func(ctx *gin.Context) {
		h := websocket.Handler(vncProxy.ServeWS)
		h.ServeHTTP(ctx.Writer, ctx.Request)
	})

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.GET("/index.html", serveIndex)
	router.GET("/", serveIndex)

	if os.Getenv("LISTEN_PID") == strconv.Itoa(os.Getpid()) {
		// systemd socket activation
		f := os.NewFile(3, "from systemd")
		listener, err := net.FileListener(f)
		if err != nil {
			panic(err)
		}
		if err := router.RunListener(listener); err != nil {
			panic(err)
		}
	} else {
		// cli activation
		if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
			panic(err)
		}
	}
}

func serveIndex(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(indexHTML))
}

func NewVNCProxy(vncport int) *vncproxy.Proxy {
	return vncproxy.New(&vncproxy.Config{
		LogLevel: vncproxy.DebugLevel,
		// DialTimeout: 10 * time.Second, // customer DialTimeout
		TokenHandler: func(r *http.Request) (addr string, err error) {
			return fmt.Sprintf(`:%d`, vncport), nil
		},
	})
}
