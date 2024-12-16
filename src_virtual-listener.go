package hahosp

import (
	"context"
	"crypto/tls"
	"net"
	"unsafe"
)

type VirtualListener struct {
	net.Listener
	TLSConf *tls.Config

	nextChan   chan struct{}
	acceptChan chan net.Conn
	context    context.Context
	cancel     context.CancelFunc
	closed     bool
	started    bool
}

func NewVisualListener(l net.Listener, config *tls.Config) *VirtualListener {
	return &VirtualListener{
		Listener: l,
		TLSConf:  config,
	}
}

func (vl *VirtualListener) Close() error {
	vl.closed = true
	return vl.Listener.Close()
}

func (vl *VirtualListener) Accept() (net.Conn, error) {
	if vl.closed {
		return nil, net.ErrClosed
	}
	if !vl.started {
		vl.started = true
		vl.context, vl.cancel = context.WithCancel(context.TODO())
		vl.nextChan = make(chan struct{}, 1)
		vl.acceptChan = make(chan net.Conn, 1)
		go vl.serve()
	}

	vl.nextChan <- struct{}{}
	select {
	case <-vl.context.Done():
		return nil, net.ErrClosed
	case c := <-vl.acceptChan:
		return c, nil
	}
}

func (vl *VirtualListener) serve() {
	for {
		c, err := vl.Listener.Accept()
		if err != nil {
			vl.closed = true
			vl.cancel()
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

	select {
	case <-vl.context.Done():
		//
	case <-vl.nextChan:
		vl.acceptChan <- c
	}
}
