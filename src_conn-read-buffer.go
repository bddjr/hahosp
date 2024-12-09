package hahosp

import (
	"net"
)

type ConnReadBuffer struct {
	net.Conn
	higher *Conn
	buf    byte
}

func (c *ConnReadBuffer) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	b[0] = c.buf
	c.higher.Conn = c.Conn
	return 1, nil
}
