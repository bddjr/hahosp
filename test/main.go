package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bddjr/hahosp"
)

func main() {
	srv := &http.Server{
		Handler:           http.HandlerFunc(handler),
		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Println()
	fmt.Println("  curl -v -k http://localhost:5678")
	fmt.Println("  curl -v -k https://localhost:5678")
	fmt.Println()

	go func() {
		err := hahosp.ListenAndServe(srv, "localhost.crt", "localhost.key")
		fmt.Println(err)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	<-c
	fmt.Println("Close")
	err := srv.Close()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.WriteHeader(200)
		if r.TLS != nil {
			io.WriteString(w, "You'r using HTTPS\n")
		} else {
			io.WriteString(w, "You'r using HTTP\n")
		}
	} else {
		http.NotFound(w, r)
	}
}
