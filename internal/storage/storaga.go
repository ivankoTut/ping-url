package storage

import "errors"

const (
	PostgresConnectionName   = "postgres"
	ClickhouseConnectionName = "clickhouse"
	RedisConnectionName      = "redis"
)

var (
	ErrUserExists = errors.New("пользователь уже зарегестрирован")
)
