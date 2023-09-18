package command

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type (
	// UpdateApiKey этот интерфейс реализует возможность генерации нового токена
	UpdateApiKey interface {
		ApiKey(ctx context.Context, userId int64) (string, error)
	}

	// ApiKeyRefresh структура для обработки команды генерации или обновления токена доступа к апи
	ApiKeyRefresh struct {
		keyGenerator UpdateApiKey
		baseUrl      string
	}
)

func NewApiKeyRefreshCommand(keyGenerator UpdateApiKey, baseUrl string) *ApiKeyRefresh {
	return &ApiKeyRefresh{
		keyGenerator: keyGenerator,
		baseUrl:      baseUrl,
	}
}

func (a *ApiKeyRefresh) CommandName() string {
	return ApiKeyRefreshCommand
}

func (a *ApiKeyRefresh) HelpText() string {
	return "help text"
}

func (a *ApiKeyRefresh) IsSupport(ctx context.Context, message *tgbotapi.Message) (bool, error) {

	if message.IsCommand() != true {
		return false, nil
	}

	return message.Command() == a.CommandName(), nil
}

func (a *ApiKeyRefresh) Run(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	userId := message.Chat.ID
	msg := tgbotapi.NewMessage(userId, "")

	key, err := a.keyGenerator.ApiKey(ctx, userId)
	if err != nil {
		msg.Text = "Произошла ошибка при генерации токена, повторите позже"
		return msg, err
	}

	msg.Text = fmt.Sprintf("api-key: <code>%s</code> \n\n ссылка: <code>%s?api-key=%s</code>", key, a.baseUrl, key)
	msg.ParseMode = tgbotapi.ModeHTML

	return msg, nil
}

func (a *ApiKeyRefresh) ClearData(ctx context.Context, message *tgbotapi.Message) error {
	return nil
}

func (a *ApiKeyRefresh) IsComplete(ctx context.Context, message *tgbotapi.Message) (bool, error) {
	return true, nil
}
