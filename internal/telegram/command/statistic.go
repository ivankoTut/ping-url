package command

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ivankoTut/ping-url/internal/model"
	"strings"
)

type (
	// CurrentStatisticUrlList этот интерфейс реализует возможность получения ссылок статистики для определенного пользователя только по существующим ссылкам
	CurrentStatisticUrlList interface {
		CurrentStatisticByUser(userId int64, urlList []string) (model.StatisticResultList, error)
	}

	// Statistic структура для обработки команды вывода текущей статистики
	Statistic struct {
		statisticRepo CurrentStatisticUrlList
		urlRepo       UserUrlList
	}
)

func NewStatisticCommand(statisticRepo CurrentStatisticUrlList, urlRepo UserUrlList) *Statistic {
	return &Statistic{
		statisticRepo: statisticRepo,
		urlRepo:       urlRepo,
	}
}

func (s *Statistic) CommandName() string {
	return StatisticCommand
}

func (s *Statistic) HelpText() string {
	return "help text"
}

func (s *Statistic) IsSupport(ctx context.Context, message *tgbotapi.Message) (bool, error) {

	if message.IsCommand() != true {
		return false, nil
	}

	return message.Command() == s.CommandName(), nil
}

func (s *Statistic) Run(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	const errorMessage = "Произошла ошибка при получении списка, повторите позже"

	userId := message.Chat.ID
	msg := tgbotapi.NewMessage(userId, "")

	_, span := tracer.Start(ctx, fmt.Sprintf("run_statistic_%d", userId))
	defer span.End()

	urlList, err := s.urlUserList(message.Chat.ID)
	if err != nil {
		span.RecordError(err)
		msg.Text = errorMessage
		return msg, err
	}

	list, err := s.statisticRepo.CurrentStatisticByUser(userId, urlList)

	if err != nil {
		span.RecordError(err)
		msg.Text = errorMessage
		return msg, err
	}

	if len(list) == 0 {
		msg.Text = "У вас еще нет записей"
		return msg, nil
	}

	str := strings.Builder{}
	for _, url := range list {
		str.WriteString(fmt.Sprintf("🌐 <code>%s</code> \n"+
			"🔄 Коли-во соединений - <code>%d</code> \n"+
			"👌 Коли-во успешных соединений - <code>%d</code> \n"+
			"⛔️ Коли-во прерваных соединений - <code>%d</code> \n"+
			"⏳ Макс-ое время ожидания - <code>%.4f</code> \n"+
			"⏳ Мин-ое время ожидания - <code>%.4f</code> \n"+
			"🕤 Среднее время ожидания - <code>%.4f</code>\n\n",
			url.Url, url.CountPing, url.CorrectCount, url.CancelCount, url.MaxConnectionTime, url.MinConnectionTime, url.AvgConnectionTime),
		)
	}
	msg.ParseMode = tgbotapi.ModeHTML
	msg.Text = str.String()

	return msg, nil
}

func (s *Statistic) ClearData(ctx context.Context, message *tgbotapi.Message) error {
	return nil
}

func (s *Statistic) IsComplete(ctx context.Context, message *tgbotapi.Message) (bool, error) {
	return true, nil
}

func (s *Statistic) urlUserList(id int64) ([]string, error) {
	list, err := s.urlRepo.UrlListByUser(id)
	if err != nil {
		return nil, err
	}

	urlList := make([]string, len(list), len(list))
	for _, item := range list {
		urlList = append(urlList, item.Url)
	}

	return urlList, nil
}
