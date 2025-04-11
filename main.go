package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	middleware_project "goHttp/internal/middleware"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

type User struct {
	Username string
	Id       int64 `json:",string"`
	Role     string
	Password Password `json:"-"`
}

type Response struct {
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

func sendJson(w http.ResponseWriter, resp Response, status int) {
	data, err := json.Marshal(resp)
	if err != nil {
		slog.Error("erro ao fazer marshal de json", "error", err)
		sendJson(w, Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		slog.Error("erro ao enviar a resposta", "error", err)
		return
	}
}

type Password string

func (p Password) String() string {
	return "[REDACTED]"
}

func (p Password) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}

const LevelFoo = slog.Level(-50)

func main() {
	z, _ := zap.NewProduction()
	zs := slog.New(zapslog.NewHandler(z.Core(), nil))
	zs.Info("uma mensagem de teste")
	p := Password("123456")
	u := User{Password: p}
	slog.Info("Password", "p", p)
	slog.Info("User", "u", u)
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     LevelFoo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "level" {
				level := a.Value.String()
				if level == "DEBUG-46" {
					a.Value = slog.StringValue("FOO")
				}
			}
			return a
		},
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	slog.SetDefault(log)

	slog.Debug("foo")

	slog.Info("Servico sendo iniciado", "version", "1.0.0")

	log = log.With(slog.Group("app_info", slog.String("version", "1.0.0")))
	log.Info("this is a test", "user", u)
	log.LogAttrs(context.Background(), LevelFoo, "qualquer mensagem")
	log.LogAttrs(context.Background(), slog.LevelInfo, "tivemos um http request",
		slog.Group(
			"http_data",
			slog.String("method", http.MethodDelete),
			slog.Int("status", http.StatusOK),
		),
		slog.Duration("time_taken", time.Second),
		slog.String("user_agent", "agent"),
	)

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

			slog.Error("falha ao ler json do usuario", "error", err)
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
