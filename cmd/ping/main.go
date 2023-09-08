package main

import (
	"github.com/ivankoTut/ping-url/internal/config"
	"github.com/ivankoTut/ping-url/internal/kernel"
	"github.com/ivankoTut/ping-url/internal/ping"
	"github.com/ivankoTut/ping-url/internal/secure"
	"github.com/ivankoTut/ping-url/internal/storage/clickhouse"
	"github.com/ivankoTut/ping-url/internal/storage/postgres"
	postgresRepository "github.com/ivankoTut/ping-url/internal/storage/postgres/repository"
	"github.com/ivankoTut/ping-url/internal/storage/redis"
	redisRepository "github.com/ivankoTut/ping-url/internal/storage/redis/repository"
	"github.com/ivankoTut/ping-url/internal/telegram"
	"github.com/ivankoTut/ping-url/internal/telegram/command"
)

func main() {
	// инициализация конфига
	cfg := config.MustLoadConfig()

	// создаем дескриптор подключения к postgres
	db := postgres.MustCreateConnection(cfg)

	// создаем дескриптор подключения к redis
	r := redis.MustCreateClientRedis(cfg)

	//подключаемся к clickhouse
	statisticRepo := clickhouse.MustCreateConnection(*cfg)

	// инициируем ядро, которое хранит и дает доступ к основным ресурсам приложения
	k := kernel.MustCreateKernel(cfg, db, r)
	k.Log().Debug("kernel is initialize")

	// инициируем бота и начинаем слушать сообщения и команды в нем
	bot := telegram.MustCreateBot(k, secure.NewUserProvider(cfg))
	go bot.StartListen()

	// подключаем команды, которые хотим обрабатывать и слушаем их
	dc := redisRepository.NewCommandRepository(r)
	pingRepository := postgresRepository.NewPing(db)
	userRepo := postgresRepository.NewUser(db)

	handlerBot := command.NewCommand(k, bot, []command.HandlerCommand{
		command.NewAddUrlCommand(dc, pingRepository),
		command.NewRemoveUrlCommand(dc, pingRepository),
		command.NewRegistrationCommand(userRepo),
		command.NewListUrlCommand(pingRepository),
		command.NewMuteCommand(userRepo),
		command.NewUnmuteAllCommand(userRepo),
		command.NewStatisticAllCommand(statisticRepo),
	})
	go handlerBot.ListenCommandAndMessage()

	// инициируем и запускаем "пингер"
	runer := ping.NewPing(pingRepository, k, statisticRepo, bot)
	go runer.Run()

	// слушаем события от бота по командам
	go runer.ListenCommandEvents(handlerBot.CommandEventChanelRead())

	// сохраняем данные по "пингам"
	runer.SaveCompleteUrl()
}
