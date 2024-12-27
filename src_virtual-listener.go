package hahosp

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
	"unsafe"
)

type VirtualListener struct {
	net.Listener
	TLSConf *tls.Config
	Server  *http.Server

	acceptChan chan net.Conn
	closeChan  chan struct{}
}

func NewVisualListener(l net.Listener, config *tls.Config, Server *http.Server) *VirtualListener {
	return &VirtualListener{
		Listener: l,
		TLSConf:  config,
		Server:   Server,
	}
}

func (vl *VirtualListener) Accept() (net.Conn, error) {
	if vl.acceptChan == nil {
		// init
		vl.acceptChan = make(chan net.Conn)
		vl.closeChan = make(chan struct{})
		go vl.serve()
	}

	select {
	case <-vl.closeChan:
		return nil, net.ErrClosed
	case c := <-vl.acceptChan:
		return c, nil
	}
}

func (vl *VirtualListener) serve() {
	for {
		c, err := vl.Listener.Accept()
		if err != nil {
			// An error is returned when the listener is closed only.
			select {
			case <-vl.closeChan:
				// closed
			default:
				close(vl.closeChan)
			}
			return
		}
		go vl.conn(c)
	}
}

func (vl *VirtualListener) conn(c net.Conn) {
	crb := &connReadBuffer{
		Conn: c,
		buf:  make([]byte, 576),
	}

	if vl.Server.ReadHeaderTimeout != 0 {
		c.SetReadDeadline(time.Now().Add(vl.Server.ReadHeaderTimeout))
	} else if vl.Server.ReadTimeout != 0 {
		c.SetReadDeadline(time.Now().Add(vl.Server.ReadTimeout))
	}

	for {
		n, err := c.Read(crb.buf)
		if err != nil {
			c.Close()
			return
		}
		if n != 0 {
			if n != len(crb.buf) {
				crb.buf = crb.buf[:n]
			}
			break
		}
	}

	switch crb.buf[0] {
	case 22: // recordTypeHandshake
		// TLS
		tc := tls.Server(crb, vl.TLSConf)
		c = tc
		crb.higher = (*conn)(unsafe.Pointer(tc))

	case 'G', // GET
		'H', // HEAD
		'P', // POST PUT PATCH
		'O', // OPTIONS
		'D', // DELETE
		'C', // CONNECT
		'T': // TRACE
		// HTTP
		crb.higher = &conn{crb}
		c = crb.higher

	default:
		// unknown
		c.Close()
		return
	}

	c.SetReadDeadline(time.Time{})

	select {
	case <-vl.closeChan:
	case vl.acceptChan <- c:
	}
}
