package vncproxy

import (
	"net"

	"github.com/evangwt/go-bufcopy"

	"github.com/pkg/errors"

	"golang.org/x/net/websocket"
)

var (
	bcopy = bufcopy.New()
)

// peer represents a vnc proxy peer
// with a websocket connection and a vnc backend connection
type peer struct {
	source *websocket.Conn
	target net.Conn
}

func NewPeer(ws *websocket.Conn, addr string) (*peer, error) {
	if ws == nil {
		return nil, errors.New("websocket connection is nil")
	}

	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "cannot connect to vnc backend")
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
