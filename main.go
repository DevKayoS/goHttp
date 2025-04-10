package main

import (
	"encoding/json"
	"errors"
	"fmt"
	middleware_project "goHttp/internal/middleware"
	"io"
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

type Response struct {
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

func sendJson(w http.ResponseWriter, resp Response, status int) {
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("erro ao fazer marshak de json: ", err)
		sendJson(w, Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		fmt.Println("error ao enviar a resposta: ", err)
		return
	}

	fmt.Println("oq eu to enviando", data)
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

		r.Get("/users/{id:[0-9]+}", handleGetUsers(db))
		r.Post("/users", handlePostUsers(db))
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
		if !ok {
			sendJson(w, Response{Error: "usuario nao encontrado"}, http.StatusNotFound)
			return
		}

		sendJson(w, Response{Data: user}, http.StatusOK)
	}
}

func handlePostUsers(db map[int64]User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1000)
		data, err := io.ReadAll(r.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				sendJson(w, Response{Error: "body too large"}, http.StatusRequestEntityTooLarge)
				return
			}
			fmt.Println(err)
			sendJson(w, Response{Error: "Something went wrong"}, http.StatusInternalServerError)
		}

		var user User
		if err := json.Unmarshal(data, &user); err != nil {
			sendJson(w, Response{Error: "Invalid body"}, http.StatusUnprocessableEntity)
			return
		}

		db[user.Id] = user

		sendJson(w, Response{Data: user}, http.StatusCreated)
	}
}
