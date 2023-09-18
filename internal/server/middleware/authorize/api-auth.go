package authorize

import (
	"context"
	"fmt"
	"github.com/ivankoTut/ping-url/internal/model"
	"net/http"
)

type (
	UserProvider interface {
		UserFromRequest(r *http.Request) (*model.User, error)
	}

	contextKey string
)

const UserContextKey = contextKey("user")

func ApiAuth(provider UserProvider) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := provider.UserFromRequest(r)

			if err != nil {
				http.Error(w, fmt.Sprintf("Forbidden: %s", err), http.StatusForbidden)
			} else {
				ctx := context.WithValue(r.Context(), UserContextKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
}
