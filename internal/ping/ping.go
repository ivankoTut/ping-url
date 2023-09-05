package ping

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ivankoTut/ping-url/internal/kernel"
	"github.com/ivankoTut/ping-url/internal/model"
	"github.com/ivankoTut/ping-url/internal/telegram"
	"github.com/ivankoTut/ping-url/internal/telegram/command"
	"github.com/ivankoTut/ping-url/internal/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"
)

const defaultCompleteUrlItems = 1000

var refreshCommandList = []string{
	command.AddUrlCommand,
	command.MuteAllCommand,
	command.UnMuteAllCommand,
}

type (
	// UrlListProvider Интерфейс реалезует возможность получать список ссылок для "пингов"
	UrlListProvider interface {
		UrlList(limit, offset int) (model.TimerPingList, error)
		Count() (int, error)
	}

	// SaveUrlStatistic Интерфейс реалезует возможность сохранять данные по "пингам"
	SaveUrlStatistic interface {
		InsertRows(model.PingResultList) error
	}

	Ping struct {
		listProvider  UrlListProvider
		kernel        *kernel.Kernel
		completeUrl   model.PingResultList
		statisticRepo SaveUrlStatistic
		countPing     int
		bot           *telegram.Bot
		rwm           sync.RWMutex
		pingList      chan model.TimerPingList
		pingQuit      chan struct{}
		saveUrlQuit   chan struct{}
	}
)

var tracer trace.Tracer

func NewPing(listProvider UrlListProvider, k *kernel.Kernel, statisticRepo SaveUrlStatistic, bot *telegram.Bot) *Ping {
	return &Ping{
		listProvider:  listProvider,
		statisticRepo: statisticRepo,
		kernel:        k,
		bot:           bot,
		completeUrl:   newCompleteList(),
		pingList:      make(chan model.TimerPingList),
		pingQuit:      make(chan struct{}),
		saveUrlQuit:   make(chan struct{}),
	}
}

func (p *Ping) Run() {
	const op = "ping.ping.Run"

	cfg := p.kernel.Config().Jaeger
	tp, err := tracing.NewJaegerTraceProvider(cfg.Url, "ping", cfg.Env)
	if err != nil {
		p.kernel.Log().Error(fmt.Sprintf("%s: ошибка инициализации Jaeger: %s", op, err))
	}

	tracer = tp.Tracer("ping")

	go func() {
		timerList, err := p.listProvider.UrlList(100, 0)
		if err != nil {
			log.Fatal(err)
		}
		p.countPing, err = p.listProvider.Count()

		if err != nil {
			p.kernel.Log().Info(fmt.Sprintf("%s, ошибка в получении кол-ва записей: %s", op, err))
		}

		p.pingList <- timerList
	}()

	for {
		for pingTime, list := range <-p.pingList {

			p.kernel.Log().Info(fmt.Sprintf("-------- start %s(%d)-------", pingTime, len(list)))

			ctx, span := tracer.Start(context.Background(), "start timer")
			span.SetAttributes(attribute.String("Ping time", pingTime), attribute.Int("Count records", len(list)))

			go p.startTicker(ctx, pingTime, list)

			span.End()
		}
	}
}

func (p *Ping) startTicker(ctx context.Context, pingTime string, list model.PingList) {
	_, span := tracer.Start(ctx, fmt.Sprintf("start for timer: %s", pingTime))
	span.End()

	timer, err := time.ParseDuration(pingTime)
	if err != nil {
		timer = time.Duration(p.kernel.Config().DefaultTimePing) * time.Second
		p.kernel.Log().Info(fmt.Sprintf("ParseDuration ERROR: %s | SET default duration %d", err, p.kernel.Config().DefaultTimePing))
	}

	ticker := time.NewTicker(timer)

	for {
		select {
		case <-ticker.C:
			for _, ping := range list {
				go p.ping(ping)
			}
		case <-p.pingQuit:
			ticker.Stop()
			return
		}
	}
}

func (p *Ping) ping(ping model.Ping) {
	start := time.Now()
	connectionTimeout, err := time.ParseDuration(ping.ConnectionTime)
	if err != nil {

	}
	client := &http.Client{Timeout: connectionTimeout}

	res, err := client.Get(ping.Url)
	if err != nil {
		p.sendErrorMessageInBot(ping, err)
		p.addCompleteUrl(ping, err, 504, time.Since(start).Seconds())
		return
	}
	defer res.Body.Close()

	p.addCompleteUrl(ping, nil, res.StatusCode, time.Since(start).Seconds())
}

func (p *Ping) addCompleteUrl(ping model.Ping, requestError error, statusCode int, realTime float64) {
	r := model.PingResult{
		Ping:               ping,
		Error:              requestError,
		RealConnectionTime: realTime,
		StatusCode:         statusCode,
	}

	p.rwm.Lock()
	p.completeUrl = append(p.completeUrl, r)
	p.rwm.Unlock()
}

func (p *Ping) SaveCompleteUrl() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		p.saveUrlQuit <- struct{}{}
	}()

	for {
		select {
		case <-time.After(time.Second * 30):
			p.startInserting()
			p.refreshPingList(false)
		case <-p.saveUrlQuit:
			p.startInserting()
			return
		}
	}
}

func (p *Ping) ListenCommandEvents(commandEvent <-chan model.CommandEvent) {
	for {
		select {
		case event := <-commandEvent:
			if p.isRefreshEvent(event) {
				p.kernel.Log().Debug(fmt.Sprintf("refresh %s command %s", event.Process, event.Command))
				p.refreshPingList(true)
			}
		}
	}
}

func (p *Ping) refreshPingList(force bool) {
	const op = "ping.ping.refreshPingList"

	count, err := p.listProvider.Count()
	if err != nil {
		p.kernel.Log().Error(fmt.Sprintf("%s, error: %s", op, err))
		return
	}

	p.kernel.Log().Debug(fmt.Sprintf("old count: %d, new count: %d", p.countPing, count))

	if count == p.countPing && force == false {
		return
	}

	timerList, err := p.listProvider.UrlList(100, 0)
	if err != nil {
		log.Fatal(fmt.Sprintf("%s, error: %s", op, err))
	}

	p.stopPing()

	p.countPing = count
	p.pingList <- timerList
}

func (p *Ping) startInserting() {
	rows := p.withdraw()
	countRows := len(rows)

	if len(rows) == 0 {
		p.kernel.Log().Info("нет строк для записи")
		return
	}

	err := p.statisticRepo.InsertRows(rows)
	if err != nil {
		p.kernel.Log().Error(fmt.Sprintf("ошибка при вставке в кликхаус: %s", err))
	} else {
		p.kernel.Log().Info(fmt.Sprintf("данные успешно вставлены, строк: %d", countRows))
	}
}

func (p *Ping) withdraw() model.PingResultList {
	p.rwm.Lock()
	defer p.rwm.Unlock()
	urls := p.completeUrl

	p.completeUrl = newCompleteList()

	return urls
}

func (p *Ping) sendErrorMessageInBot(ping model.Ping, err error) {
	if ping.User.Mute {
		return
	}

	msg := tgbotapi.NewMessage(ping.UserId, fmt.Sprintf("<code>⚠️%s</code> \n \n <u>%s</u>", ping.Url, err))
	msg.ParseMode = tgbotapi.ModeHTML
	p.bot.SendMessage(msg)
}

func (p *Ping) stopPing() {
	p.pingQuit <- struct{}{}
}

func (p *Ping) isRefreshEvent(event model.CommandEvent) bool {
	return slices.Contains(refreshCommandList, event.Command)
}

func newCompleteList() model.PingResultList {
	return make(model.PingResultList, 0, defaultCompleteUrlItems)
}
