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
		Addr: ":5688",
		Handler: &hahosp.HandlerSelector{
			HTTPS: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, "You'r using HTTPS\n")
			}),
			HTTP: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, "You'r using HTTP\n")
			}),
		},
		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Println()
	fmt.Println("  http://localhost" + srv.Addr)
	fmt.Println("  https://localhost" + srv.Addr)
	fmt.Println()

	go func() {
		err := hahosp.ListenAndServeTLS(srv, "localhost.crt", "localhost.key")
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
