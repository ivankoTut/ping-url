package repository

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/ivankoTut/ping-url/internal/kernel"
	"github.com/ivankoTut/ping-url/internal/model"
	"net/http"
	"time"
)

type User struct {
	connection kernel.DBConnection
}

func NewUser(db kernel.DBConnection) *User {
	return &User{connection: db}
}

func (u *User) UserExist(ctx context.Context, userId int64) (bool, error) {
	const op = "storage.postgres.repository.user.UserExist"

	stmt, err := u.connection.DB().Prepare("SELECT count(*) FROM users WHERE id = $1")
	if err != nil {
		return false, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var count int
	err = stmt.QueryRow(userId).Scan(&count)

	return count > 0, err
}

func (u *User) UserSave(ctx context.Context, userId int64, login string) error {
	const op = "storage.postgres.repository.user.UserSave"

	stmt, err := u.connection.DB().Prepare("INSERT INTO users(id, login) VALUES($1, $2)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(userId, login)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (u *User) Mute(ctx context.Context, userId int64) error {
	return u.muteUnmute(ctx, userId, true)
}

func (u *User) Unmute(ctx context.Context, userId int64) error {
	return u.muteUnmute(ctx, userId, false)
}

func (u *User) muteUnmute(ctx context.Context, userId int64, mute bool) error {
	const op = "storage.postgres.repository.user.muteUnmute"

	stmt, err := u.connection.DB().Prepare("update users set mute = $1 where id = $2")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(mute, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (u *User) UserFromRequest(r *http.Request) (*model.User, error) {
	const apiKeY = "api-key"
	var key string

	keyQuery := r.URL.Query()[apiKeY]
	if len(keyQuery) == 1 {
		key = keyQuery[0]
	}

	if key == "" {
		key = r.Header.Get(apiKeY)
	}

	if key == "" {
		return nil, errors.New("не передан заголовок содержащий api ключ доступа")
	}

	stmt, err := u.connection.DB().Prepare("SELECT id, login, mute FROM users WHERE api_key = $1")
	if err != nil {
		return nil, fmt.Errorf("prepare statement: %w", err)
	}

	var user model.User
	err = stmt.QueryRow(key).Scan(&user.Id, &user.Login, &user.Mute)

	if user.Id == 0 {
		return nil, errors.New("пользователь не найден")
	}

	return &user, err
}

func (u *User) ApiKey(ctx context.Context, userId int64) (string, error) {
	const op = "storage.postgres.repository.user.ApiKey"

	sum := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", time.Now().String(), userId)))
	keyApi := fmt.Sprintf("%x", sum)

	stmt, err := u.connection.DB().Prepare("update users set api_key = $1 where id = $2")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(keyApi, userId)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return keyApi, nil
}
