package hahosp

import (
	"crypto/tls"
	"net"
	"net/http"

	hahosp_utils "github.com/bddjr/hahosp/utils"
)

func ListenAndServeTLS(srv *http.Server, certFile string, keyFile string) error {
	if hahosp_utils.IsShuttingDown(srv) {
		return http.ErrServerClosed
	}
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()

	return Serve(l, srv, certFile, keyFile)
}

// Deprecated: Move to [ListenAndServeTLS]
func ListenAndServe(srv *http.Server, certFile string, keyFile string) error {
	return ListenAndServeTLS(srv, certFile, keyFile)
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

	return srv.Serve(NewVisualListener(l, config, srv))
}
