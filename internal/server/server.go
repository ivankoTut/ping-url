package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ivankoTut/ping-url/internal/server/middleware/authorize"
	"github.com/ivankoTut/ping-url/internal/storage/postgres/repository"
	"net/http"
	"time"
)

func RunApiServer(userRepo *repository.User, addr string) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(authorize.ApiAuth(userRepo))

	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	http.ListenAndServe(addr, r)
}
