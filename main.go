package main

import (
	"encoding/json"
	"fmt"
	middleware_project "goHttp/internal/middleware"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type User struct {
	Username string
	Id       int64 `json:",string"`
	Role     string
	Password string `json:"-"`
}

func main() {
	r := chi.NewMux()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	db := map[int64]User{
		1: {
			Id:       1,
			Role:     "admin",
			Username: "admin",
			Password: "admin",
		},
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware_project.JsonMiddleware)

		r.Get("/users/{id:[0-9]}", handleGetUsers(db))
		r.Post("/users", handlePostUsers)
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/users/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				fmt.Println(id)
			})
		})

		r.Route("/v2", func(r chi.Router) {

		})

		r.With(middleware.RealIP).Get("/users", func(w http.ResponseWriter, r *http.Request) {})

		r.Group(func(r chi.Router) {
			r.Use(middleware.BasicAuth("", map[string]string{
				"admin": "admin",
			}))

			r.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "ping")
			})
		})
	})

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}

func handleGetUsers(db map[int64]User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			panic(err)
		}

		user, ok := db[id]
		if ok {
			data, err := json.Marshal(user)

			if err != nil {
				panic(err)
			}

			w.Write(data)
		}
	}
}

func handlePostUsers(w http.ResponseWriter, r *http.Request) {

}
