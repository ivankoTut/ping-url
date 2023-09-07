package command

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

const stateRemoveUrlNone = -1

const (
	stateRemoveUrlBegin  = iota //Начало удаление ссылки
	stateRemoveUrlSetUrl        //ссылка которая будет удалена
)

type (
	// UrlRemover этот интерфейс реализует возможность удалять ссылки
	UrlRemover interface {
		RemoveUrl(userId int64, url string) error
		UrlExist(userId int64, url string) (bool, error)
	}

	// RemoveUrl структура для обработки команды удаления ссылки
	RemoveUrl struct {
		urlRepo   UrlRemover
		dialog    DialogChain
		state     int
		questions []string
	}
)

func NewRemoveUrlCommand(dialog DialogChain, urlRepo UrlRemover) *RemoveUrl {
	return &RemoveUrl{
		urlRepo: urlRepo,
		dialog:  dialog,
		questions: []string{
			"Укажите url адрес который необходимо удалить",
		},
	}
}

func (r *RemoveUrl) CommandName() string {
	return RemoveUrlCommand
}

func (r *RemoveUrl) HelpText() string {
	return "help text"
}

func (r *RemoveUrl) key(message *tgbotapi.Message) string {
	return fmt.Sprintf("%d_%s", message.Chat.ID, r.CommandName())
}

func (r *RemoveUrl) keyAnswer(message *tgbotapi.Message) string {
	return fmt.Sprintf("%d_%s_answer", message.Chat.ID, r.CommandName())
}

func (r *RemoveUrl) IsSupport(ctx context.Context, message *tgbotapi.Message) (bool, error) {

	if message.IsCommand() == true {
		return message.Command() == r.CommandName(), nil
	}

	key := r.key(message)

	is, err := r.dialog.DialogExist(ctx, key)
	if err != nil {
		return false, err
	}

	return is, nil
}

func (r *RemoveUrl) IsComplete(ctx context.Context, message *tgbotapi.Message) (bool, error) {

	key := r.key(message)

	is, err := r.dialog.DialogExist(ctx, key)
	if err != nil {
		return false, err
	}

	return is == false, nil
}

func (r *RemoveUrl) Run(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("Run command: %s", r.CommandName()))
	defer span.End()
	key := r.key(message)
	msg := tgbotapi.NewMessage(message.Chat.ID, "")

	state, err := r.dialog.CurrentState(ctx, key)
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
	case stateRemoveUrlNone:
		nextState = stateRemoveUrlBegin
		msg.Text = r.questions[nextState]
	case stateRemoveUrlBegin:
		nextState = stateRemoveUrlSetUrl

		is, errExist := r.urlRepo.UrlExist(message.Chat.ID, message.Text)
		if errExist != nil {
			msg.Text = "Ошибка при проверке ссылки, повторите ввод"
			span.RecordError(err)
			return msg, errExist
		}

		if !is {
			msg.Text = "Данная ссылка не существует"
			return msg, nil
		}

		if err := r.urlRepo.RemoveUrl(message.Chat.ID, message.Text); err != nil {
			msg.Text = "Произошла ошибка при удалении, повторите позже"

			span.RecordError(err)

			return msg, err
		}

		if err := r.ClearData(ctx, message); err != nil {
			msg.Text = "Произошла ошибка, повторите позже"

			span.RecordError(err)

			return msg, err
		}

		msg.Text = "Ссылка удалена"

		return msg, err
	default:
		nextState = stateRemoveUrlNone
		msg.Text = r.questions[0]
	}

	_, errSave := r.dialog.SaveState(ctx, key, nextState)
	if errSave != nil {
		msg.Text = "ошибка при сохранении текущего шага"

		span.RecordError(errSave)

		return msg, errSave
	}

	return msg, err
}

func (r *RemoveUrl) ClearData(ctx context.Context, message *tgbotapi.Message) error {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("clear data for %s", r.CommandName()))
	defer span.End()

	if err := r.dialog.DeleteDialog(ctx, r.key(message)); err != nil {
		span.RecordError(err)
		return err
	}

	return r.dialog.DeleteDialog(ctx, r.keyAnswer(message))
}
