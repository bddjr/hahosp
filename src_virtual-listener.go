package hahosp

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
	"unsafe"
)

type VirtualListener struct {
	net.Listener
	TLSConf *tls.Config
	Server  *http.Server

	acceptChan chan net.Conn
	closeChan  chan struct{}
	initOnce   sync.Once
}

func NewVisualListener(l net.Listener, config *tls.Config, Server *http.Server) *VirtualListener {
	return &VirtualListener{
		Listener: l,
		TLSConf:  config,
		Server:   Server,
	}
}

func (vl *VirtualListener) Accept() (net.Conn, error) {
	vl.initOnce.Do(func() {
		vl.acceptChan = make(chan net.Conn)
		vl.closeChan = make(chan struct{})
		go vl.serve()
	})

	select {
	case <-vl.closeChan:
		return nil, net.ErrClosed
	case c := <-vl.acceptChan:
		return c, nil
	}
}

func (vl *VirtualListener) serve() {
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		c, err := vl.Listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				vl.logf("hahosp: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			close(vl.closeChan)
			return
		}
		tempDelay = 0
		go vl.conn(c)
	}
}

func (vl *VirtualListener) conn(c net.Conn) {
	if vl.Server.ReadHeaderTimeout != 0 {
		c.SetReadDeadline(time.Now().Add(vl.Server.ReadHeaderTimeout))
	} else if vl.Server.ReadTimeout != 0 {
		c.SetReadDeadline(time.Now().Add(vl.Server.ReadTimeout))
	}

	buf := make([]byte, 576)

	for {
		n, err := c.Read(buf)
		if err != nil {
			c.Close()
			return
		}
		if n != 0 {
			if n != len(buf) {
				buf = buf[:n]
			}
			break
		}
	}

	crb := &connReadBuffer{
		Conn: c,
		buf:  buf,
	}

	if buf[0] <= 23 && buf[0] >= 20 {
		// TLS
		tc := tls.Server(crb, vl.TLSConf)
		c = tc
		crb.higher = (*conn)(unsafe.Pointer(tc))
	} else if buf[0] >= 'A' && buf[0] <= 'Z' {
		// HTTP
		crb.higher = &conn{crb}
		c = crb.higher
	} else {
		// unknown
		c.Close()
		return
	}

	c.SetReadDeadline(time.Time{})

	select {
	case <-vl.closeChan:
		c.Close()
	case vl.acceptChan <- c:
		//
	}
}

func (vl *VirtualListener) logf(format string, args ...interface{}) {
	if vl.Server.ErrorLog != nil {
		vl.Server.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}
