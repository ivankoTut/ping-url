package statistics

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/ivankoTut/ping-url/internal/model"
	"github.com/ivankoTut/ping-url/internal/server/middleware/authorize"
	"github.com/ivankoTut/ping-url/internal/telegram/command"
	"log/slog"
	"net/http"
)

func NewAll(log *slog.Logger, statsRepo command.StatisticUrlList) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const (
			op           = "server.handlers.statistics.all"
			errorMessage = "Ошибка получения статистики"
		)

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		user := r.Context().Value(authorize.UserContextKey).(*model.User)
		stats, err := statsRepo.StatisticByUser(user.Id)

		if err != nil {
			log.Error(fmt.Sprintf("%s: %s", errorMessage, err))
			render.JSON(w, r, errorMessage)
			return
		}

		log.Info(fmt.Sprintf("show all statistics, user_id: %d", user.Id))

		render.JSON(w, r, stats)
	}
}
