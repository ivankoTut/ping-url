package clickhouse

import (
	"database/sql"
	"fmt"
	"github.com/ivankoTut/ping-url/internal/config"
	"github.com/ivankoTut/ping-url/internal/model"
	"github.com/ivankoTut/ping-url/internal/storage"
	"github.com/mailru/go-clickhouse/v2"
	"log"
	"time"
)

type (
	Db struct {
		cfg  config.Config
		conn *sql.DB
	}
)

func MustCreateConnection(cfg config.Config) *Db {
	connStr := fmt.Sprintf("http://%s:%d/?user=%s&password=%s", cfg.Database.Clickhouse.Host, cfg.Database.Clickhouse.Port, cfg.Database.Clickhouse.User, cfg.Database.Clickhouse.Password)
	driver := "chhttp"
	clickhouse.NewConfig()
	connect, err := sql.Open(driver, connStr)

	if err != nil {
		log.Fatal(err)
	}

	if err := connect.Ping(); err != nil {
		log.Fatal(111, err)
	}

	db := &Db{
		cfg:  cfg,
		conn: connect,
	}

	db.MustRunMigrations()

	return db
}

func (db *Db) MustRunMigrations() {
	tx, err := db.conn.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(`
		CREATE TABLE IF NOT EXISTS url_status (
			userId Int64,
			url String,
			statusCode Int64,
			error String,
			pingTime Float64,
			createdAt DateTime
		)
		ENGINE = MergeTree
		ORDER BY (userId, url, statusCode);
		`)

	if _, err := stmt.Exec(); err != nil {
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (db *Db) ConnectionName() string {
	return storage.ClickhouseConnectionName
}

func (db *Db) DB() *sql.DB {
	return db.conn
}

func (db *Db) InsertRows(urls model.PingResultList) error {
	tx, err := db.conn.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO url_status (userId, url, statusCode, error, pingTime, createdAt)
		VALUES (
			?, ?, ?, ?, ?, ?
		)`)

	if err != nil {
		log.Fatal(err)
	}

	for _, v := range urls {

		errMessage := ""
		if v.Error != nil {
			errMessage = v.Error.Error()
		}

		if _, err := stmt.Exec(
			v.Ping.UserId,
			v.Ping.Url,
			v.StatusCode,
			errMessage,
			v.RealConnectionTime,
			time.Now(),
		); err != nil {
			fmt.Println(err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	return nil
}
