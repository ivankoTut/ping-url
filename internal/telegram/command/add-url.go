package command

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
	"net/url"
	"time"
)

const stateNone = -1

const (
	stateBegin             = iota //Начало добавление ссылки
	stateAddConnectionTime        //максимальное время ожидания ответа по ссылке
	statePingTime                 //Время через которое необходимо делать опрос по ссылке
)

const (
	answerUrl            = "url"             // see stateBegin
	answerConnectionTime = "connection_time" // see stateAddConnectionTime
	answerPingTime       = "ping_time"       // see statePingTime
)

type (
	// UrlSaver этот интерфейс реализует возможность сохранения новой ссылки
	UrlSaver interface {
		SaveUrl(userId int64, url, connectionTime, pingTime string) error
		UrlExist(userId int64, url string) (bool, error)
	}

	// AddUrl структура для обработки команды добавления новой ссылки
	AddUrl struct {
		urlRepo   UrlSaver
		dialog    DialogChain
		state     int
		questions []string
	}
)

func NewAddUrlCommand(dialog DialogChain, urlRepo UrlSaver) *AddUrl {
	return &AddUrl{
		urlRepo: urlRepo,
		dialog:  dialog,
		questions: []string{
			"Укажите url адрес",
			"Укажите максимально время ожидания ответа, примеры: 100ms|10s|1h|1s500ms",
			"Укажите время с какой периодичностью необходимо опрашивать ссылку в секундах (минимально 30), примеры: 30m20s|1h",
		},
	}
}

func (a *AddUrl) CommandName() string {
	return addUrlCommand
}

func (a *AddUrl) HelpText() string {
	return "help text"
}

func (a *AddUrl) key(message *tgbotapi.Message) string {
	return fmt.Sprintf("%d_%s", message.Chat.ID, a.CommandName())
}

func (a *AddUrl) keyAnswer(message *tgbotapi.Message) string {
	return fmt.Sprintf("%d_%s_answer", message.Chat.ID, a.CommandName())
}

func (a *AddUrl) IsSupport(ctx context.Context, message *tgbotapi.Message) (bool, error) {

	if message.IsCommand() == true {
		return message.Command() == a.CommandName(), nil
	}

	key := a.key(message)

	is, err := a.dialog.DialogExist(ctx, key)
	if err != nil {
		return false, err
	}

	return is, nil
}

func (a *AddUrl) Run(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("Run command: %s", a.CommandName()))
	defer span.End()
	key := a.key(message)
	msg := tgbotapi.NewMessage(message.Chat.ID, "")

	state, err := a.dialog.CurrentState(ctx, key)
	if err != nil && err != redis.Nil {
		msg.Text = "ошибка при попытке создать запись"
		span.RecordError(err)
		return msg, err
	}

	if err == redis.Nil {
		err = nil
		state = stateNone
	}

	var nextState int
	switch state {
	case stateNone:
		nextState = stateBegin
		msg.Text = a.questions[nextState]
	case stateBegin:
		nextState = stateAddConnectionTime

		// валидация ссылки
		u, errUrl := url.Parse(message.Text)
		if errUrl != nil || (u.Scheme == "" || u.Host == "") {
			msg.Text = "Не валидная ссылка, повторите ввод"
			span.RecordError(err)
			return msg, errUrl
		}

		is, errExist := a.urlRepo.UrlExist(message.Chat.ID, message.Text)
		if errExist != nil {
			msg.Text = "Ошибка при проверке ссылки, повторите ввод"
			span.RecordError(err)
			return msg, errUrl
		}

		if is {
			msg.Text = "Данная ссылка уже существует"
			return msg, nil
		}

		err = a.dialog.SaveAnswer(ctx, a.keyAnswer(message), answerUrl, message.Text)
		if err != nil {
			msg.Text = "ошибка при сохранении ссылки, повторите попытку"
			nextState = stateBegin
		} else {
			msg.Text = a.questions[nextState]
		}
	case stateAddConnectionTime:
		// валидация времени ожидания ответа от сервера
		_, err = time.ParseDuration(message.Text)
		if err != nil {
			fmt.Println("sdf")
			msg.Text = "указано неверное время, примеры: 100ms|10s|1h|1s500ms"

			span.RecordError(err)

			return msg, err
		}

		nextState = statePingTime
		err = a.dialog.SaveAnswer(ctx, a.keyAnswer(message), answerConnectionTime, message.Text)
		if err != nil {
			msg.Text = "ошибка при сохранении время отклика, повторите попытку"
			nextState = stateAddConnectionTime
		} else {
			msg.Text = a.questions[nextState]
		}
	case statePingTime:
		// валидация времени ожидания ответа от сервера
		timer, errTime := time.ParseDuration(message.Text)
		if errTime != nil {
			msg.Text = "указано неверное время, примеры: 100ms|10s|1h|1s500ms"

			span.RecordError(errTime)

			return msg, errTime
		}

		// проверяем на минимальное время повторений
		if timer < time.Second*30 {
			message.Text = "30s"
		}

		err = a.dialog.SaveAnswer(ctx, a.keyAnswer(message), answerPingTime, message.Text)
		if err != nil {
			msg.Text = "ошибка при сохранении время повторения, повторите попытку"
			nextState = statePingTime
		} else {
			err := a.saveUrl(ctx, message)
			if err != nil {
				msg.Text = "Произошла ошибка при сохранении, повторите позже"
			} else {
				msg.Text = "запись успешно добавлена"
			}

			if err := a.ClearData(ctx, message); err != nil {
				msg.Text = "Произошла ошибка, повторите позже"

				span.RecordError(err)

				return msg, err
			}

			return msg, err
		}
	default:
		nextState = stateNone
		msg.Text = a.questions[0]
	}

	_, errSave := a.dialog.SaveState(ctx, key, nextState)
	if errSave != nil {
		msg.Text = "ошибка при сохранении текущего шага"

		span.RecordError(errSave)

		return msg, errSave
	}

	return msg, err
}

func (a *AddUrl) saveUrl(ctx context.Context, message *tgbotapi.Message) error {

	answers, err := a.dialog.GetAnswer(ctx, a.keyAnswer(message))
	if err != nil {
		return err
	}

	return a.urlRepo.SaveUrl(message.Chat.ID, answers[answerUrl], answers[answerConnectionTime], answers[answerPingTime])
}

func (a *AddUrl) ClearData(ctx context.Context, message *tgbotapi.Message) error {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("clear data for %s", a.CommandName()))
	defer span.End()

	if err := a.dialog.DeleteDialog(ctx, a.key(message)); err != nil {
		span.RecordError(err)
		return err
	}

	return a.dialog.DeleteDialog(ctx, a.keyAnswer(message))
}
