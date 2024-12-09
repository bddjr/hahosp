package hahosp

import (
	"crypto/tls"
	"net"
	"unsafe"
)

type VisualListener struct {
	net.Listener
	TLSConf    *tls.Config
	nextChan   chan struct{}
	acceptChan chan any
}

func NewVisualListener(l net.Listener, config *tls.Config) *VisualListener {
	return &VisualListener{
		Listener:   l,
		TLSConf:    config,
		nextChan:   make(chan struct{}, 1),
		acceptChan: make(chan any, 1),
	}
}

func (vl *VisualListener) Accept() (net.Conn, error) {
	vl.nextChan <- struct{}{}
	c := <-vl.acceptChan
	if conn, ok := c.(net.Conn); ok {
		return conn, nil
	}
	err := c.(error)
	if err == net.ErrClosed {
		close(vl.acceptChan)
	}
	return nil, err
}

func (vl *VisualListener) Serve() {
	for {
		c, err := vl.Listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				<-vl.nextChan
				vl.acceptChan <- err
				continue
			}
			vl.Listener.Close()
			<-vl.nextChan
			vl.acceptChan <- net.ErrClosed
			close(vl.nextChan)
			return
		}
		go vl.serve(c)
	}
}

func (vl *VisualListener) serve(c net.Conn) {
	b := make([]byte, 1)
	for {
		n, err := c.Read(b)
		if err != nil {
			c.Close()
			<-vl.nextChan
			vl.acceptChan <- &ConnReadError{
				Conn: c,
				err:  err,
			}
			return
		}
		if n == 1 {
			break
		}
	}

	crb := &ConnReadBuffer{
		Conn: c,
		buf:  b[0],
	}

	switch b[0] {
	case 20, // recordTypeChangeCipherSpec
		21,   // recordTypeAlert
		22,   // recordTypeHandshake
		23,   // recordTypeApplicationData
		0x80: // error: unsupported SSLv2 handshake received
		// TLS
		tc := tls.Server(crb, vl.TLSConf)
		c = tc
		crb.higher = (*Conn)(unsafe.Pointer(tc))

	case 'G', // GET
		'H', // HEAD
		'P', // POST PUT PATCH
		'O', // OPTIONS
		'D', // DELETE
		'C', // CONNECT
		'T': // TRACE
		// HTTP
		crb.higher = &Conn{crb}
		c = crb.higher

	default:
		// unknown
		c.Close()
		return
	}

	<-vl.nextChan
	vl.acceptChan <- c
}
