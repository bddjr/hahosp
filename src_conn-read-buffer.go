package hahosp

import (
	"net"
)

type connReadBuffer struct {
	net.Conn
	higher *conn
	buf    []byte
}

func (c *connReadBuffer) Read(b []byte) (int, error) {
	if len(c.buf) == 0 {
		return c.Conn.Read(b)
	}

	n := copy(b, c.buf)
	if n == len(c.buf) {
		c.buf = nil
	} else {
		c.buf = c.buf[n:]
	}

	if c.higher != nil {
		c.higher.Conn = c.Conn
		c.higher = nil
	}

	return n, nil
}
