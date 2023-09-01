package command

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ivankoTut/ping-url/internal/model"
	"strings"
)

type (
	// UserUrlList этот интерфейс реализует возможность получения ссылок текущего пользователя
	UserUrlList interface {
		UrlListByUser(userId int64) (model.PingList, error)
	}

	// ListUrl структура для обработки команды добавления новой ссылки
	ListUrl struct {
		urlRepo UserUrlList
	}
)

func NewListUrlCommand(urlRepo UserUrlList) *ListUrl {
	return &ListUrl{
		urlRepo: urlRepo,
	}
}

func (l *ListUrl) CommandName() string {
	return listUrlCommand
}

func (l *ListUrl) HelpText() string {
	return "help text"
}

func (l *ListUrl) IsSupport(ctx context.Context, message *tgbotapi.Message) (bool, error) {

	if message.IsCommand() != true {
		return false, nil
	}

	return message.Command() == l.CommandName(), nil
}

func (l *ListUrl) Run(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	userId := message.Chat.ID
	msg := tgbotapi.NewMessage(userId, "")

	list, err := l.urlRepo.UrlListByUser(userId)

	if err != nil {
		msg.Text = "Произошла ошибка при получении списка, повторите позже"
		return msg, err
	}

	str := strings.Builder{}
	for _, url := range list {
		str.WriteString(fmt.Sprintf("🌐 <code>%s</code> \n⏳ Время ожидания - <code>%s</code> \n🕤 Время периодичности - <code>%s</code>\n\n", url.Url, url.ConnectionTime, url.PingTime))
	}
	msg.ParseMode = tgbotapi.ModeHTML
	msg.Text = str.String()

	return msg, nil
}

func (l *ListUrl) ClearData(ctx context.Context, message *tgbotapi.Message) error {

	return nil
}
