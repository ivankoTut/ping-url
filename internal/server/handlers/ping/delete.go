package ping

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/ivankoTut/ping-url/internal/model"
	"github.com/ivankoTut/ping-url/internal/server/middleware/authorize"
	"log/slog"
	"net/http"
)

type (
	// UrlRemover этот интерфейс реализует возможность удалять ссылки
	UrlRemover interface {
		RemoveUrlById(userId int64, id string) error
		UrlExistById(userId int64, id string) (bool, error)
	}
)

func NewDelete(log *slog.Logger, urlListRepo UrlRemover) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const (
			op              = "server.handlers.statistics.delete"
			errorMessage    = "Ошибка удаления ссылки"
			notFoundMessage = "Ссылка не найдена"
		)

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		user := r.Context().Value(authorize.UserContextKey).(*model.User)
		urlId := chi.URLParam(r, "id")

		is, err := urlListRepo.UrlExistById(user.Id, urlId)

		if err != nil {
			log.Error(fmt.Sprintf("%s: %s", errorMessage, err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, errorMessage)
			return
		}

		if !is {
			log.Error(fmt.Sprintf("%s: %s", notFoundMessage, err))
			sendErrorMessage(w, r, errorMessage)
			return
		}

		err = urlListRepo.RemoveUrlById(user.Id, urlId)
		if err != nil {
			log.Error(fmt.Sprintf("%s: %s", errorMessage, err))
			sendErrorMessage(w, r, errorMessage)
			return
		}

		log.Info(fmt.Sprintf("delete url - id: %s list user_id: %d", urlId, user.Id))

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, "")
	}
}

func sendErrorMessage(w http.ResponseWriter, r *http.Request, errorMessage string) {
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, errorMessage)
}
