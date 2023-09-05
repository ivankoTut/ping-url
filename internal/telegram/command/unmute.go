package command

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type (
	// UserUnmuteNotification этот интерфейс реализует возможность включить уведомления для пользователя
	UserUnmuteNotification interface {
		Unmute(ctx context.Context, userId int64) error
	}

	// UnmuteAll структура для обработки команды включения нотификаций
	UnmuteAll struct {
		userUnmute UserUnmuteNotification
	}
)

func NewUnmuteAllCommand(userUnmute UserUnmuteNotification) *UnmuteAll {
	return &UnmuteAll{
		userUnmute: userUnmute,
	}
}

func (u *UnmuteAll) CommandName() string {
	return UnMuteAllCommand
}

func (u *UnmuteAll) HelpText() string {
	return "help text"
}

func (u *UnmuteAll) IsSupport(ctx context.Context, message *tgbotapi.Message) (bool, error) {

	if message.IsCommand() != true {
		return false, nil
	}

	return message.Command() == u.CommandName(), nil
}

func (u *UnmuteAll) Run(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	userId := message.Chat.ID
	msg := tgbotapi.NewMessage(userId, "")

	err := u.userUnmute.Unmute(ctx, userId)

	if err != nil {
		msg.Text = "Произошла ошибка при включении уведомлений, повторите позже"
		return msg, err
	}

	msg.Text = "Уведомления включены"

	return msg, nil
}

func (u *UnmuteAll) ClearData(ctx context.Context, message *tgbotapi.Message) error {
	return nil
}

func (u *UnmuteAll) IsComplete(ctx context.Context, message *tgbotapi.Message) (bool, error) {
	return true, nil
}
