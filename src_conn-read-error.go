package hahosp

import "net"

type ConnReadError struct {
	net.Conn
	err error
}

func (ce ConnReadError) Read(b []byte) (int, error) {
	err := ce.err
	ce.err = net.ErrClosed
	return 0, err
}
