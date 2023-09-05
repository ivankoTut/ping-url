package command

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type (
	// UserMuteNotification этот интерфейс реализует возможность отключить уведомления для пользователя
	UserMuteNotification interface {
		Mute(ctx context.Context, userId int64) error
	}

	// MuteAll структура для обработки команды отключения нотификаций
	MuteAll struct {
		userMute UserMuteNotification
	}
)

func NewMuteCommand(userMute UserMuteNotification) *MuteAll {
	return &MuteAll{
		userMute: userMute,
	}
}

func (m *MuteAll) CommandName() string {
	return MuteAllCommand
}

func (m *MuteAll) HelpText() string {
	return "help text"
}

func (m *MuteAll) IsSupport(ctx context.Context, message *tgbotapi.Message) (bool, error) {

	if message.IsCommand() != true {
		return false, nil
	}

	return message.Command() == m.CommandName(), nil
}

func (m *MuteAll) Run(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	userId := message.Chat.ID
	msg := tgbotapi.NewMessage(userId, "")

	err := m.userMute.Mute(ctx, userId)

	if err != nil {
		msg.Text = "Произошла ошибка при отключении уведомлений, повторите позже"
		return msg, err
	}

	msg.Text = "Уведомления отключены"

	return msg, nil
}

func (m *MuteAll) ClearData(ctx context.Context, message *tgbotapi.Message) error {
	return nil
}

func (m *MuteAll) IsComplete(ctx context.Context, message *tgbotapi.Message) (bool, error) {
	return true, nil
}
