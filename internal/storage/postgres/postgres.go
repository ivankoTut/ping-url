package postgres

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ivankoTut/ping-url/internal/config"
	"github.com/ivankoTut/ping-url/internal/storage"
	"log"
)

type Db struct {
	db  *sql.DB
	cfg *config.Config
}

func MustCreateConnection(cfg *config.Config) *Db {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.Postgres.User,
		cfg.Database.Postgres.Password,
		cfg.Database.Postgres.Host,
		cfg.Database.Postgres.Port,
		cfg.Database.Postgres.Database,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return &Db{
		db:  db,
		cfg: cfg,
	}
}

func (db *Db) MustRunMigrations() {
	driver, err := postgres.WithInstance(db.db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		db.cfg.Database.Postgres.MigrationPath,
		"sqlite3",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
}

func (db *Db) ConnectionName() string {
	return storage.PostgresConnectionName
}

func (db *Db) DB() *sql.DB {
	return db.db
}
