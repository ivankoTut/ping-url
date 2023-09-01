package redis

import (
	"context"
	"github.com/ivankoTut/ping-url/internal/config"
	"github.com/ivankoTut/ping-url/internal/storage"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

const TtlTime = time.Hour * 12

type ClientRedis struct {
	rdb *redis.Client
	cfg *config.Config
}

func MustCreateClientRedis(cfg *config.Config) *ClientRedis {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Database.Redis.Addr,
		Password: cfg.Database.Redis.Password,
		DB:       cfg.Database.Redis.Db,
	})
	//rdb.AddHook(new(redisHook))

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal(err)
	}

	if err := redisotel.InstrumentTracing(rdb); err != nil {
		log.Fatal(err)
	}

	// Enable metrics instrumentation.
	if err := redisotel.InstrumentMetrics(rdb); err != nil {
		log.Fatal(err)
	}

	return &ClientRedis{
		rdb: rdb,
		cfg: cfg,
	}
}

func (c *ClientRedis) Client() redis.Client {
	return *c.rdb
}

func (c *ClientRedis) ConnectionName() string {
	return storage.RedisConnectionName
}
