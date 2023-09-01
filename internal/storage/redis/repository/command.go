package repository

import (
	"context"
	r "github.com/ivankoTut/ping-url/internal/storage/redis"
	"github.com/redis/go-redis/v9"
)

type CommandRepository struct {
	cr *r.ClientRedis
}

func NewCommandRepository(cr *r.ClientRedis) *CommandRepository {
	return &CommandRepository{
		cr: cr,
	}
}

func (c *CommandRepository) SaveState(ctx context.Context, key string, state int) (bool, error) {
	return true, c.cr.Client().Set(ctx, key, state, r.TtlTime).Err()
}

func (c *CommandRepository) DialogExist(ctx context.Context, key string) (bool, error) {
	_, err := c.cr.Client().Get(ctx, key).Int()
	if err == redis.Nil {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *CommandRepository) CurrentState(ctx context.Context, key string) (int, error) {
	return c.cr.Client().Get(ctx, key).Int()
}

func (c *CommandRepository) SaveAnswer(ctx context.Context, key, state string, answer interface{}) error {
	cli := c.cr.Client()
	if err := cli.HSet(ctx, key, state, answer).Err(); err != nil {
		return err
	}

	return cli.Expire(ctx, key, r.TtlTime).Err()
}

func (c *CommandRepository) GetAnswer(ctx context.Context, key string) (map[string]string, error) {
	return c.cr.Client().HGetAll(ctx, key).Result()
}

func (c *CommandRepository) DeleteDialog(ctx context.Context, key string) error {
	return c.cr.Client().Del(ctx, key).Err()
}
