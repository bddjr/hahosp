package hahosp

import (
	"crypto/tls"
	"net"
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
		Conn:   c,
		higher: nil,
		buf:    b[0],
	}

	crb.higher = &Conn{crb}
	c = crb.higher

	if crb.buf < 'A' || crb.buf > 'Z' {
		// HTTPS
		c = tls.Server(c, vl.TLSConf)
	}

	<-vl.nextChan
	vl.acceptChan <- c
}
