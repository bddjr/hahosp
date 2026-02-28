package main_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/bddjr/hahosp"
	hahosp_utils "github.com/bddjr/hahosp/utils"
	"golang.org/x/net/http2"
)

func tlsVersionName(version uint16) string {
	switch version {
	case 0x0301:
		return "1.0"
	case 0x0302:
		return "1.1"
	case 0x0303:
		return "1.2"
	case 0x0304:
		return "1.3"
	default:
		return fmt.Sprintf("Unknown 0x%04X", version)
	}
}

func request(serverAddr string) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	err := http2.ConfigureTransport(transport)
	if err != nil {
		panic(err)
	}

	client := http.Client{
		Transport: transport,
		Timeout:   time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			println("Redirect")
			return nil
		},
	}
	defer client.CloseIdleConnections()

	const uri = "/test?a=b&c=d"

	for _, scheme := range []string{"http", "https"} {
		url := scheme + "://" + serverAddr + uri
		println(url)
		resp, err := client.Get(url)
		if err != nil {
			panic(err)
		}
		if scheme == "https" && resp.ProtoMajor != 2 {
			panic("Response does not using h2 protocol!")
		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		print(string(respBody))
	}
}

func Test(t *testing.T) {
	const serverAddr = "127.0.0.1:5688"
	srv := &http.Server{
		Addr: serverAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "HEAD" {
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			enc := json.NewEncoder(w)
			enc.SetEscapeHTML(false)
			var TLS map[string]any
			if r.TLS != nil {
				TLS = map[string]any{
					"TLS_Version":    tlsVersionName(r.TLS.Version),
					"TLS_ServerName": r.TLS.ServerName,
				}
			}
			m := map[string]any{
				"Method": r.Method,
				"Proto":  r.Proto,
				"TLS":    TLS,
				"Host":   r.Host,
				"URI":    r.RequestURI,
				// "RequestHeader":  r.Header,
			}
			err := enc.Encode(m)
			if err != nil {
				panic(err)
			}
		}),
	}

	println("Listen " + serverAddr)
	if hahosp_utils.IsShuttingDown(srv) {
		panic(true)
	}

	var err error
	go func() {
		err = hahosp.ListenAndServeTLS(srv, "localhost.crt", "localhost.key")
	}()
	time.Sleep(100 * time.Millisecond)
	if err != nil {
		panic(err)
	}
	println()

	request(serverAddr)

	println("Shutdown")
	err = srv.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
	if !hahosp_utils.IsShuttingDown(srv) {
		panic(false)
	}
	println()
}
