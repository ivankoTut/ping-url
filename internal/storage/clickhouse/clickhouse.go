package clickhouse

import (
	"database/sql"
	"fmt"
	"github.com/ivankoTut/ping-url/internal/config"
	"github.com/ivankoTut/ping-url/internal/model"
	"github.com/ivankoTut/ping-url/internal/storage"
	"github.com/mailru/go-clickhouse/v2"
	"log"
	"strings"
	"time"
)

const baseStatisticSelect = `url as Url,
			count(url) as CountPing,
			SUM(CASE
				WHEN isCancel = false THEN 1
			    ELSE 0
			END) as CorrectCount,
		    SUM(CASE
				WHEN isCancel = true THEN 1
			    ELSE 0
			END) as CancelCount,
			max(CASE
				WHEN isCancel = false THEN pingTime
			END) as MaxConnectionTime,
			min(CASE
				WHEN isCancel = false THEN pingTime
			END) as MinConnectionTime,
			avg(CASE
				WHEN isCancel = false THEN pingTime
			END) as AvgConnectionTime
		from url_status `

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

	stmt, err = tx.Prepare(`
		ALTER TABLE url_status ADD COLUMN IF NOT EXISTS isCancel Bool DEFAULT true
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
		INSERT INTO url_status (userId, url, statusCode, error, pingTime, createdAt, isCancel)
		VALUES (
			?, ?, ?, ?, ?, ?, ?
		)`)

	if err != nil {
		return err
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
			v.IsCancel,
		); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (db *Db) StatisticByUser(userId int64) (model.StatisticResultList, error) {
	rows, err := db.conn.Query(`
		select `+baseStatisticSelect+` where userId = ?
		group by url
		order by AvgConnectionTime desc;
	`, userId)

	if err != nil {
		return nil, err
	}

	var list model.StatisticResultList
	for rows.Next() {
		var item model.Statistic
		if err := rows.Scan(
			&item.Url,
			&item.CountPing,
			&item.CorrectCount,
			&item.CancelCount,
			&item.MaxConnectionTime,
			&item.MinConnectionTime,
			&item.AvgConnectionTime,
		); err != nil {
			return nil, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (db *Db) CurrentStatisticByUser(userId int64, urlList []string) (model.StatisticResultList, error) {
	params := []interface{}{userId}
	for _, v := range urlList {
		params = append(params, v)
	}

	rows, err := db.conn.Query(`
		select `+baseStatisticSelect+` where 
		    userId = ? and url in (?, `+strings.Repeat("?, ", len(urlList)-1)+`)
		group by url
		order by AvgConnectionTime desc;
	`, params...)

	if err != nil {
		return nil, err
	}

	var list model.StatisticResultList
	for rows.Next() {
		var item model.Statistic
		if err := rows.Scan(
			&item.Url,
			&item.CountPing,
			&item.CorrectCount,
			&item.CancelCount,
			&item.MaxConnectionTime,
			&item.MinConnectionTime,
			&item.AvgConnectionTime,
		); err != nil {
			return nil, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (db *Db) StatisticByUrl(userId int64, url string) (model.Statistic, error) {

	rows, err := db.conn.Query(`select error as errorText, count(error) as count from url_status where userId = ? and url = ? and error <> '' group by error order by count desc`, userId, url)
	if err != nil {
		return model.Statistic{}, err
	}
	var errorList []model.ErrorMessage
	for rows.Next() {
		var errorText model.ErrorMessage
		if err := rows.Scan(&errorText.Text, &errorText.Count); err != nil {
			return model.Statistic{}, err
		}

		errorList = append(errorList, errorText)
	}

	statsList, err := db.CurrentStatisticByUser(userId, []string{url})
	if err != nil {
		return model.Statistic{}, err
	}

	statsList[0].Errors = errorList

	return statsList[0], nil
}
