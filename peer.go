package vncproxy

import (
	"net"
	"time"

	"github.com/evangwt/go-bufcopy"

	"github.com/pkg/errors"
)

const (
	defaultDialTimeout = 5 * time.Second
)

var (
	bcopy = bufcopy.New()
)

// peer represents a vnc proxy peer
// with a websocket connection and a vnc backend connection
type peer struct {
	source WebSocketAdapter
	target net.Conn
}

func NewPeer(ws WebSocketAdapter, addr string, dialTimeout time.Duration) (*peer, error) {
	if ws == nil {
		return nil, errors.New("websocket connection is nil")
	}

	if len(addr) == 0 {
		return nil, errors.New("addr is empty")
	}

	if dialTimeout <= 0 {
		dialTimeout = defaultDialTimeout
	}
	c, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "cannot connect to vnc backend")
	}

	err = c.(*net.TCPConn).SetKeepAlive(true)
	if err != nil {
		return nil, errors.Wrap(err, "enable vnc backend connection keepalive failed")
	}

	err = c.(*net.TCPConn).SetKeepAlivePeriod(30 * time.Second)
	if err != nil {
		return nil, errors.Wrap(err, "set vnc backend connection keepalive period failed")
	}

	return &peer{
		source: ws,
		target: c,
	}, nil
}

// ReadSource copy source stream to target connection
func (p *peer) ReadSource() error {
	if _, err := bcopy.Copy(p.target, p.source); err != nil {
		return errors.Wrapf(err, "copy source(%v) => target(%v) failed", p.source.RemoteAddr(), p.target.RemoteAddr())
	}
	return nil
}

// ReadTarget copys target stream to source connection
func (p *peer) ReadTarget() error {
	if _, err := bcopy.Copy(p.source, p.target); err != nil {
		return errors.Wrapf(err, "copy target(%v) => source(%v) failed", p.target.RemoteAddr(), p.source.RemoteAddr())
	}
	return nil
}

// Close close the websocket connection and the vnc backend connection
func (p *peer) Close() {
	p.source.Close()
	p.target.Close()
}
