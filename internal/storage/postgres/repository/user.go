package repository

import (
	"context"
	"fmt"
	"github.com/ivankoTut/ping-url/internal/kernel"
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
