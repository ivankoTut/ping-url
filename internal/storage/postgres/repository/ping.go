package repository

import (
	"fmt"
	"github.com/ivankoTut/ping-url/internal/kernel"
	"github.com/ivankoTut/ping-url/internal/model"
)

type Ping struct {
	connection kernel.DBConnection
}

func NewPing(db kernel.DBConnection) *Ping {
	return &Ping{connection: db}
}

func (p *Ping) SaveUrl(userId int64, url, connectionTime, pingTime string) error {
	const op = "storage.postgres.repository.ping.SaveUrl"
	stmt, err := p.connection.DB().Prepare(`INSERT INTO ping(user_id, url, connection_time, ping_time) VALUES($1, $2, $3, $4)`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(userId, url, connectionTime, pingTime)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Ping) RemoveUrl(userId int64, url string) error {
	const op = "storage.postgres.repository.ping.RemoveUrl"
	stmt, err := p.connection.DB().Prepare(`delete from ping where user_id = $1 and url = $2`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(userId, url)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Ping) UrlExist(userId int64, url string) (bool, error) {
	const op = "storage.postgres.repository.ping.SaveUrl"
	stmt, err := p.connection.DB().Prepare(`SELECT count(*) FROM ping WHERE user_id = $1 and url = $2`)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	var count int
	err = stmt.QueryRow(userId, url).Scan(&count)

	return count > 0, err
}

func (p *Ping) UrlListByUser(userId int64) (model.PingList, error) {
	const op = "storage.postgres.repository.ping.UrlListByUser"

	rows, err := p.connection.DB().Query(`
		select p.url, p.user_id, p.connection_time, p.ping_time, u.id, u.login, u.mute from ping as p 
		left join users as u on p.user_id = u.id
		where p.user_id = $1`, userId,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var links model.PingList
	var link model.Ping
	for rows.Next() {
		err := rows.Scan(&link.Url, &link.UserId, &link.ConnectionTime, &link.PingTime, &link.User.Id, &link.User.Login, &link.User.Mute)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		links = append(links, link)
	}

	return links, nil
}

func (p *Ping) UrlList(limit, offset int) (model.TimerPingList, error) {
	const op = "storage.postgres.repository.ping.UrlList"

	rows, err := p.connection.DB().Query(`
		select p.url, p.user_id, p.connection_time, p.ping_time, u.id, u.login, u.mute from ping as p 
		left join users as u on p.user_id = u.id limit $1 offset $2`,
		limit,
		offset,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	links := make(model.TimerPingList)
	var link model.Ping
	for rows.Next() {
		err := rows.Scan(&link.Url, &link.UserId, &link.ConnectionTime, &link.PingTime, &link.User.Id, &link.User.Login, &link.User.Mute)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		links[link.PingTime] = append(links[link.PingTime], link)
	}

	return links, nil
}

func (p *Ping) Count() (int, error) {
	const op = "storage.postgres.repository.ping.UrlList"

	rows, err := p.connection.DB().Query(`select count(*) from ping`)
	count := 0
	if err != nil {
		return count, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&count)

	return count, err
}
