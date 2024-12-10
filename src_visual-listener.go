package hahosp

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"runtime"
	"unsafe"
)

type VisualListener struct {
	net.Listener
	TLSConf  *tls.Config
	ErrorLog *log.Logger

	nextChan   chan struct{}
	acceptChan chan net.Conn
	context    context.Context
	cancel     context.CancelFunc
	closed     bool
	started    bool
}

func NewVisualListener(l net.Listener, config *tls.Config, ErrorLog *log.Logger) *VisualListener {
	return &VisualListener{
		Listener: l,
		TLSConf:  config,
		ErrorLog: ErrorLog,
	}
}

func (vl *VisualListener) Close() error {
	vl.closed = true
	return vl.Listener.Close()
}

func (vl *VisualListener) Accept() (net.Conn, error) {
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

func (vl *VisualListener) serve() {
	for {
		c, err := vl.Listener.Accept()
		if err != nil {
			vl.cancel()
			return
		}
		go vl.conn(c)
	}
}

func (vl *VisualListener) logf(format string, v ...any) {
	if vl.ErrorLog != nil {
		vl.ErrorLog.Printf(format, v...)
	} else {
		log.Printf(format, v...)
	}
}

func (vl *VisualListener) conn(c net.Conn) {
	defer func() {
		// catch panic
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			vl.logf("hahosp: panic serving %s: %v\n%s", c.RemoteAddr(), err, buf)
		}
	}()

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
	case 20, // recordTypeChangeCipherSpec
		21,   // recordTypeAlert
		22,   // recordTypeHandshake
		23,   // recordTypeApplicationData
		0x80: // error: unsupported SSLv2 handshake received
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
