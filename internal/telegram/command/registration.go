package command

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ivankoTut/ping-url/internal/storage"
)

type (
	RegistrationUser interface {
		UserSave(ctx context.Context, userId int64, login string) error
		UserExist(ctx context.Context, userId int64) (bool, error)
	}
	Registration struct {
		userRepo RegistrationUser
	}
)

func NewRegistrationCommand(userRepo RegistrationUser) *Registration {
	return &Registration{
		userRepo: userRepo,
	}
}

func (r *Registration) CommandName() string {
	return registrationCommand
}

func (r *Registration) HelpText() string {
	return "help text"
}

func (r *Registration) IsSupport(ctx context.Context, message *tgbotapi.Message) (bool, error) {
	if message.IsCommand() != true {
		return false, nil
	}

	return message.Command() == r.CommandName(), nil
}

func (r *Registration) ClearData(ctx context.Context, message *tgbotapi.Message) error {
	return nil
}

func (r *Registration) Run(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")

	ok, err := r.userRepo.UserExist(ctx, message.Chat.ID)
	if ok {
		msg.Text = "Вы уже зарегестрированы"
		return msg, storage.ErrUserExists
	}

	if err != nil {
		msg.Text = "Произошла ошибка, повторите позже"
		return msg, err
	}

	login := message.Chat.UserName
	if login == "" {
		login = message.Chat.FirstName
	}

	msg.Text = "Вы успешно зарегестрировались"
	if err = r.userRepo.UserSave(ctx, message.Chat.ID, login); err != nil {
		msg.Text = "Произошла ошибка при сохранении, повторите позже"
	}

	return msg, err
}
