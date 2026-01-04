package vncproxy

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type TokenHandler func(r *http.Request) (addr string, err error)

// Config represents vnc proxy config
type Config struct {
	LogLevel       uint32
	Logger         Logger
	DialTimeout    time.Duration
	TokenHandler   TokenHandler
	WebSocketAdapter WebSocketUpgrader // Optional: custom WebSocket adapter
}

// Proxy represents vnc proxy
type Proxy struct {
	logLevel         uint32
	logger           *logger
	dialTimeout      time.Duration // Timeout for connecting to each target vnc server
	peers            map[*peer]struct{}
	l                sync.RWMutex
	tokenHandler     TokenHandler
	webSocketAdapter WebSocketUpgrader
}

// New returns a vnc proxy
// If token handler is nil, vnc backend address will always be :5901
func New(conf *Config) *Proxy {
	if conf.TokenHandler == nil {
		conf.TokenHandler = func(r *http.Request) (addr string, err error) {
			return ":5901", nil
		}
	}

	// Create default WebSocket adapter if none provided
	var wsAdapter WebSocketUpgrader
	if conf.WebSocketAdapter != nil {
		wsAdapter = conf.WebSocketAdapter
	}

	return &Proxy{
		logLevel:         conf.LogLevel,
		logger:           NewLogger(conf.LogLevel, conf.Logger),
		dialTimeout:      conf.DialTimeout,
		peers:            make(map[*peer]struct{}),
		l:                sync.RWMutex{},
		tokenHandler:     conf.TokenHandler,
		webSocketAdapter: wsAdapter,
	}
}

// ServeWS provides websocket handler for golang.org/x/net/websocket (backward compatibility)
func (p *Proxy) ServeWS(ws *websocket.Conn) {
	adapter := NewNetWebSocketAdapter(ws)
	p.ServeWSAdapter(adapter)
}

// ServeWSAdapter provides websocket handler using the WebSocketAdapter interface
func (p *Proxy) ServeWSAdapter(ws WebSocketAdapter) {
	p.logger.Debugf("ServeWSAdapter")
	
	// Set binary mode for VNC protocol
	if err := ws.SetBinaryMode(); err != nil {
		p.logger.Infof("failed to set binary mode: %v", err)
		return
	}

	r := ws.Request()
	p.logger.Debugf("request url: %v", r.URL)

	// get vnc backend server addr
	addr, err := p.tokenHandler(r)
	if err != nil {
		p.logger.Infof("get vnc backend failed: %v", err)
		return
	}

	peer, err := NewPeer(ws, addr, p.dialTimeout)
	if err != nil {
		p.logger.Infof("new vnc peer failed: %v", err)
		return
	}

	p.addPeer(peer)
	defer func() {
		p.logger.Info("close peer")
		p.deletePeer(peer)
	}()

	go func() {
		if err := peer.ReadTarget(); err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			p.logger.Info(err)
			return
		}
	}()

	if err = peer.ReadSource(); err != nil {
		if strings.Contains(err.Error(), "use of closed network connection") {
			return
		}
		p.logger.Info(err)
		return
	}
}

func (p *Proxy) addPeer(peer *peer) {
	p.l.Lock()
	p.peers[peer] = struct{}{}
	p.l.Unlock()
}

func (p *Proxy) deletePeer(peer *peer) {
	p.l.Lock()
	delete(p.peers, peer)
	peer.Close()
	p.l.Unlock()
}

// Handler returns an HTTP handler that can work with custom WebSocket adapters
// This allows users to plug in different WebSocket libraries
func (p *Proxy) Handler() WebSocketHandler {
	return p.ServeWSAdapter
}

// HTTPHandler returns an HTTP handler using the configured WebSocket adapter
// If no adapter is configured, it uses the default golang.org/x/net/websocket adapter
func (p *Proxy) HTTPHandler() http.Handler {
	if p.webSocketAdapter != nil {
		// Return a handler that uses the custom adapter
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ws, err := p.webSocketAdapter.Upgrade(w, r)
			if err != nil {
				p.logger.Infof("WebSocket upgrade failed: %v", err)
				return
			}
			p.ServeWSAdapter(ws)
		})
	}
	
	// Use default golang.org/x/net/websocket adapter
	upgrader := NewNetWebSocketUpgrader(p.Handler())
	return upgrader
}

func (p *Proxy) Peers() map[*peer]struct{} {
	p.l.RLock()
	defer p.l.RUnlock()
	return p.peers
}
