package hahosp

import (
	"crypto/tls"
	"net"
	"net/http"
	"reflect"
	"sync/atomic"
	"unsafe"
)

func ListenAndServe(srv *http.Server, certFile string, keyFile string) error {
	if IsShuttingDown(srv) {
		return http.ErrServerClosed
	}

	l, err := net.Listen("tcp", ":5678")
	if err != nil {
		return err
	}

	return Serve(l, srv, certFile, keyFile)
}

func Serve(l net.Listener, srv *http.Server, certFile string, keyFile string) error {
	// Setup HTTP/2
	if srv.TLSConfig == nil {
		srv.TLSConfig = &tls.Config{}
	}
	if len(srv.TLSConfig.NextProtos) == 0 {
		srv.TLSConfig.NextProtos = []string{"h2", "http/1.1"}
	}

	// clone tls config
	config := srv.TLSConfig.Clone()

	configHasCert := len(config.Certificates) > 0 || config.GetCertificate != nil || config.GetConfigForClient != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
	}

	vl := NewVisualListener(l, config)
	go vl.Serve()
	return srv.Serve(vl)
}

func IsShuttingDown(srv *http.Server) bool {
	inShutdown := (*atomic.Bool)(unsafe.Pointer(reflect.ValueOf(srv).Elem().FieldByName("inShutdown").UnsafeAddr()))
	return inShutdown.Load()
}
