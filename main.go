package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"/healthcheck",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "hello, world")
		},
	)

	server := &http.Server{
		Addr:                         ":8080",
		Handler:                      mux,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    &tls.Config{},
		ReadTimeout:                  10 * time.Second,
		WriteTimeout:                 10 * time.Second,
		IdleTimeout:                  1 * time.Minute,
		MaxHeaderBytes:               0,
		TLSNextProto:                 map[string]func(*http.Server, *tls.Conn, http.Handler){},
		ConnState: func(net.Conn, http.ConnState) {
			panic("TODO")
		},
		ErrorLog: &log.Logger{},
		BaseContext: func(net.Listener) context.Context {
			panic("TODO")
		},
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			panic("TODO")
		},
	}
}
