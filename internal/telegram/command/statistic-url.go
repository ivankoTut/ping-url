package command

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ivankoTut/ping-url/internal/model"
	"github.com/redis/go-redis/v9"
	"strings"
)

const stateStatisticUrlNone = -1

const (
	stateStatisticUrlBegin  = iota //Начало получения статистики по ссылке
	stateStatisticUrlSetUrl        //ссылка по которой будет предоставлена статистика
)

type (
	// UrlStatistic этот интерфейс реализует возможность полученияданных по ссылке
	UrlStatistic interface {
		StatisticByUrl(userId int64, url string) (model.Statistic, error)
	}

	// UrlRepositoryExist этот интерфейс реализует возможность проверить наличие ссылки у пользователя
	UrlRepositoryExist interface {
		UrlExist(userId int64, url string) (bool, error)
	}

	// StatisticUrl структура для обработки команды вывода статистики для определенной ссылки
	StatisticUrl struct {
		statisticRepo UrlStatistic
		dialog        DialogChain
		urlRepo       UrlRepositoryExist
		questions     []string
	}
)

func NewStatisticUrlCommand(statisticRepo UrlStatistic, dialog DialogChain, urlRepo UrlRepositoryExist) *StatisticUrl {
	return &StatisticUrl{
		statisticRepo: statisticRepo,
		dialog:        dialog,
		urlRepo:       urlRepo,
		questions: []string{
			"Укажите url адрес по которому необходимо вывести статистику",
		},
	}
}

func (s *StatisticUrl) CommandName() string {
	return StatisticUrlCommand
}

func (s *StatisticUrl) HelpText() string {
	return "help text"
}

func (s *StatisticUrl) IsSupport(ctx context.Context, message *tgbotapi.Message) (bool, error) {

	if message.IsCommand() == true {
		return message.Command() == s.CommandName(), nil
	}

	key := s.key(message)

	is, err := s.dialog.DialogExist(ctx, key)
	if err != nil {
		return false, err
	}

	return is, nil
}

func (s *StatisticUrl) key(message *tgbotapi.Message) string {
	return fmt.Sprintf("%d_%s", message.Chat.ID, s.CommandName())
}

func (s *StatisticUrl) keyAnswer(message *tgbotapi.Message) string {
	return fmt.Sprintf("%d_%s_answer", message.Chat.ID, s.CommandName())
}

func (s *StatisticUrl) Run(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("Run command: %s", s.CommandName()))
	defer span.End()
	key := s.key(message)
	msg := tgbotapi.NewMessage(message.Chat.ID, "")

	state, err := s.dialog.CurrentState(ctx, key)
	if err != nil && err != redis.Nil {
		msg.Text = "ошибка при попытке создать запись"
		span.RecordError(err)
		return msg, err
	}

	if err == redis.Nil {
		err = nil
		state = stateRemoveUrlNone
	}

	var nextState int
	switch state {
	case stateStatisticUrlNone:
		nextState = stateStatisticUrlBegin
		msg.Text = s.questions[nextState]
	case stateStatisticUrlBegin:
		nextState = stateStatisticUrlSetUrl

		is, errExist := s.urlRepo.UrlExist(message.Chat.ID, message.Text)
		if errExist != nil {
			msg.Text = "Ошибка при проверке ссылки, повторите ввод"
			span.RecordError(err)
			return msg, errExist
		}

		if !is {
			msg.Text = "Данная ссылка не существует"
			return msg, nil
		}

		stats, err := s.statisticRepo.StatisticByUrl(message.Chat.ID, message.Text)
		if err != nil {
			msg.Text = "Произошла ошибка при получении статистики"

			span.RecordError(err)

			return msg, err
		}

		if err := s.ClearData(ctx, message); err != nil {
			msg.Text = "Произошла ошибка, повторите позже"

			span.RecordError(err)

			return msg, err
		}

		str := strings.Builder{}
		str.WriteString(fmt.Sprintf("🌐 <code>%s</code> \n"+
			"🔄 Коли-во соединений - <code>%d</code> \n"+
			"👌 Коли-во успешных соединений - <code>%d</code> \n"+
			"⛔️ Коли-во прерваных соединений - <code>%d</code> \n"+
			"⏳ Макс-ое время ожидания - <code>%.4f</code> \n"+
			"⏳ Мин-ое время ожидания - <code>%.4f</code> \n"+
			"🕤 Среднее время ожидания - <code>%.4f</code>\n\n",
			stats.Url, stats.CountPing, stats.CorrectCount, stats.CancelCount, stats.MaxConnectionTime, stats.MinConnectionTime, stats.AvgConnectionTime,
		))

		if len(stats.Errors) > 0 {
			str.WriteString("Спсиок ошибок\n\n")
			for _, errText := range stats.Errors {
				str.WriteString(fmt.Sprintf("Кол-во: %d\n<code>%s</code> \n\n", errText.Count, errText.Text))
			}
		}

		msg.ParseMode = tgbotapi.ModeHTML
		msg.Text = str.String()

		return msg, err
	default:
		nextState = stateRemoveUrlNone
		msg.Text = s.questions[0]
	}

	_, errSave := s.dialog.SaveState(ctx, key, nextState)
	if errSave != nil {
		msg.Text = "ошибка при сохранении текущего шага"

		span.RecordError(errSave)

		return msg, errSave
	}

	return msg, err
}

func (s *StatisticUrl) ClearData(ctx context.Context, message *tgbotapi.Message) error {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("clear data for %s", s.CommandName()))
	defer span.End()

	if err := s.dialog.DeleteDialog(ctx, s.key(message)); err != nil {
		span.RecordError(err)
		return err
	}

	return s.dialog.DeleteDialog(ctx, s.keyAnswer(message))
}

func (s *StatisticUrl) IsComplete(ctx context.Context, message *tgbotapi.Message) (bool, error) {
	key := s.key(message)

	is, err := s.dialog.DialogExist(ctx, key)
	if err != nil {
		return false, err
	}

	return is == false, nil
}
