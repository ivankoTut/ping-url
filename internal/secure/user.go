package secure

import (
	"github.com/ivankoTut/ping-url/internal/config"
	"slices"
)

type UserProvider struct {
	cfg *config.Config
}

func NewUserProvider(cfg *config.Config) *UserProvider {
	return &UserProvider{cfg: cfg}
}

func (u *UserProvider) IsAccess(userId int64) bool {
	if len(u.cfg.AccessUserList) == 0 {
		return true
	}

	return slices.Contains(u.cfg.AccessUserList, userId)
}
