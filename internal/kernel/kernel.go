package kernel

import (
	"database/sql"
	"fmt"
	"github.com/ivankoTut/ping-url/internal/config"
	"github.com/ivankoTut/ping-url/internal/logger"
	"github.com/ivankoTut/ping-url/internal/storage/redis"
	"log/slog"
)

type (

	// DBConnection дескриптор подключения к бд
	DBConnection interface {
		MustRunMigrations()
		DB() *sql.DB
		ConnectionName() string
	}

	// Kernel основная структура которая хранит в себе и дает доступ к:
	// конфигу, базе данных, логеру, кликхаусу ...
	Kernel struct {
		cfg   *config.Config
		log   *slog.Logger
		db    DBConnection
		redis *redis.ClientRedis
	}
)

func MustCreateKernel(cfg *config.Config, db DBConnection, redis *redis.ClientRedis) *Kernel {
	log := logger.NewLogger(cfg.Env, cfg.LogFile)
	log.Debug("logger initialize!")

	db.MustRunMigrations()
	log.Debug(fmt.Sprintf("migration is complete for: %s", db.ConnectionName()))
	return &Kernel{
		cfg:   cfg,
		log:   log,
		db:    db,
		redis: redis,
	}
}

// Config возвращает структуру конфига для чтения
func (k *Kernel) Config() config.Config {
	return *k.cfg
}

// Log возвращает логер
func (k *Kernel) Log() *slog.Logger {
	return k.log
}

// Db дескриптор подключения к бд
func (k *Kernel) Db() *sql.DB {
	return k.db.DB()
}
