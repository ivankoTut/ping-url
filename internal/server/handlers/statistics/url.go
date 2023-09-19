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

func NewUrl(log *slog.Logger, pingRepo command.UrlRepositoryExist, statsRepo command.UrlStatistic) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const (
			op           = "server.handlers.statistics.url"
			errorMessage = "Ошибка получения статистики"
			urlNotFound  = "Ссылка не найдена"
		)

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var url string
		keyQuery := r.URL.Query()["url"]
		if len(keyQuery) == 1 {
			url = keyQuery[0]
		}

		user := r.Context().Value(authorize.UserContextKey).(*model.User)
		is, err := pingRepo.UrlExist(user.Id, url)

		if err != nil {
			log.Error(fmt.Sprintf("%s: %s", errorMessage, err))
			render.JSON(w, r, errorMessage)
			return
		}

		if is == false {
			log.Error(fmt.Sprintf("%s: %s", urlNotFound, url))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, urlNotFound)
			return
		}

		stats, err := statsRepo.StatisticByUrl(user.Id, url)
		if err != nil {
			log.Error(fmt.Sprintf("%s: %s", errorMessage, err))
			render.JSON(w, r, errorMessage)
			return
		}

		log.Info(fmt.Sprintf("show statistics by url: %s, user_id: %d", url, user.Id))

		render.JSON(w, r, stats)
	}
}
