package command

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ivankoTut/ping-url/internal/model"
	"strings"
)

type (
	// StatisticUrlList этот интерфейс реализует возможность получения ссылок статистики для определенного пользователя
	StatisticUrlList interface {
		StatisticByUser(userId int64) (model.StatisticResultList, error)
	}

	// StatisticAll структура для обработки команды вывода общей статистики
	StatisticAll struct {
		statisticRepo StatisticUrlList
	}
)

func NewStatisticAllCommand(statisticRepo StatisticUrlList) *StatisticAll {
	return &StatisticAll{
		statisticRepo: statisticRepo,
	}
}

func (s *StatisticAll) CommandName() string {
	return StatisticAllCommand
}

func (s *StatisticAll) HelpText() string {
	return "help text"
}

func (s *StatisticAll) IsSupport(ctx context.Context, message *tgbotapi.Message) (bool, error) {

	if message.IsCommand() != true {
		return false, nil
	}

	return message.Command() == s.CommandName(), nil
}

func (s *StatisticAll) Run(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	userId := message.Chat.ID
	msg := tgbotapi.NewMessage(userId, "")

	_, span := tracer.Start(ctx, fmt.Sprintf("run_statistic_%d", userId))
	defer span.End()

	list, err := s.statisticRepo.StatisticByUser(userId)

	if err != nil {
		span.RecordError(err)
		msg.Text = "Произошла ошибка при получении списка, повторите позже"
		return msg, err
	}

	if len(list) == 0 {
		msg.Text = "У вас еще нет записей"
		return msg, nil
	}

	str := strings.Builder{}
	for _, url := range list {
		str.WriteString(fmt.Sprintf("🌐 <code>%s</code> \n"+
			"🔄 Коли-во соединений = <code>%d</code> \n"+
			"⏳ Макс-ое время ожидания - <code>%.4f</code> \n"+
			"⏳ Мин-ое время ожидания - <code>%.4f</code> \n"+
			"🕤 Среднее время ожидания - <code>%.4f</code>\n\n",
			url.Url, url.CountPing, url.MaxConnectionTime, url.MinConnectionTime, url.AvgConnectionTime),
		)
	}
	msg.ParseMode = tgbotapi.ModeHTML
	msg.Text = str.String()

	return msg, nil
}

func (s *StatisticAll) ClearData(ctx context.Context, message *tgbotapi.Message) error {
	return nil
}

func (s *StatisticAll) IsComplete(ctx context.Context, message *tgbotapi.Message) (bool, error) {
	return true, nil
}
