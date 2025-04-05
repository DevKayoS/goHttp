package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		next.ServeHTTP(w, r)
		fmt.Println(r.URL.String(), r.Method, time.Since(begin))
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"POST /healthcheck",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "hello, world")
		},
	)

	mux.HandleFunc(
		"POST /api/users/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")
			fmt.Fprintln(w, id)
		},
	)

	server := &http.Server{
		Addr:                         ":8080",
		Handler:                      log(mux),
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    &tls.Config{},
		ReadTimeout:                  10 * time.Second,
		WriteTimeout:                 10 * time.Second,
		IdleTimeout:                  1 * time.Minute,
	}

	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}
}
