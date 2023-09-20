package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ivankoTut/ping-url/internal/kernel"
	"github.com/ivankoTut/ping-url/internal/server/handlers/ping"
	"github.com/ivankoTut/ping-url/internal/server/handlers/statistics"
	"github.com/ivankoTut/ping-url/internal/server/middleware/authorize"
	"github.com/ivankoTut/ping-url/internal/server/middleware/logger"
	"github.com/ivankoTut/ping-url/internal/storage/clickhouse"
	"github.com/ivankoTut/ping-url/internal/storage/postgres/repository"
	"net/http"
	"time"
)

func RunApiServer(userRepo *repository.User, k *kernel.Kernel, clickhouseStatsRepo *clickhouse.Db, pingRepository *repository.Ping) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(logger.New(k.Log()))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(authorize.ApiAuth(userRepo))
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/statistics", func(r chi.Router) {
		r.Get("/all", statistics.NewAll(k.Log(), clickhouseStatsRepo))
		r.Get("/url", statistics.NewUrl(k.Log(), pingRepository, clickhouseStatsRepo))
	})

	r.Route("/ping", func(r chi.Router) {
		r.Get("/", ping.NewList(k.Log(), pingRepository))
		r.Delete("/{id}", ping.NewDelete(k.Log(), pingRepository))
	})

	http.ListenAndServe(k.Config().BaseApiUrl, r)
}
