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
	const errmsg = "hahosp: failed to check type conn"
	a := reflect.TypeOf(conn{})
	b := reflect.TypeOf(tls.Conn{})
	if a.Kind() != b.Kind() {
		panic(errmsg)
	}
	anf := a.NumField()
	for i := 0; i < anf; i++ {
		af := a.Field(i)
		bf := b.Field(i)
		if af.Offset != bf.Offset {
			panic(errmsg)
		}
		aft := af.Type
		aftk := aft.Kind()
		bft := bf.Type
		if aftk != bft.Kind() {
			panic(errmsg)
		}
		if aftk == reflect.Pointer {
			aft = af.Type.Elem()
			aftk = aft.Kind()
			bft = bf.Type.Elem()
			if aftk != bft.Kind() {
				panic(errmsg)
			}
		}
		if aft.Size() != bft.Size() {
			panic(errmsg)
		}
		if aft.PkgPath() != bft.PkgPath() {
			panic(errmsg)
		}
		if aft.Name() != bft.Name() {
			panic(errmsg)
		}
	}
	return
}()
