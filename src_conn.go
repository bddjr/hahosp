package hahosp

import (
	"crypto/tls"
	"net"
	"reflect"
)

type conn struct {
	net.Conn
}

// Automatic type checking
var _ = func() (_ struct{}) {
	const errmsg = "github.com/bddjr/hahosp: failed to check type conn"
	a := reflect.TypeOf(conn{})
	if a.NumField() != 1 {
		panic(errmsg)
	}
	if a.Field(0).Type != reflect.TypeOf(tls.Conn{}).Field(0).Type {
		panic(errmsg)
	}
	return
}()
