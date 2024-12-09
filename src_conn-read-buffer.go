package hahosp

import (
	"net"
)

type ConnReadBuffer struct {
	net.Conn
	higher *Conn // read raw conn if nil
	buf    byte
}

func (c *ConnReadBuffer) Read(b []byte) (int, error) {
	if c.higher == nil {
		return c.Conn.Read(b)
	}

	b[0] = c.buf
	c.higher.Conn = c.Conn
	c.higher = nil
	return 1, nil
}
