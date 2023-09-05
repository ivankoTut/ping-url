package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ivankoTut/ping-url/internal/kernel"
	"log"
)

type (
	AccessUserProvider interface {
		IsAccess(userId int64) bool
	}
	Bot struct {
		bot                *tgbotapi.BotAPI
		cfg                tgbotapi.UpdateConfig
		kernel             *kernel.Kernel
		accessUserProvider AccessUserProvider
		Command            chan *tgbotapi.Message
		Message            chan *tgbotapi.Message
	}
)

func MustCreateBot(kernel *kernel.Kernel, accessUserProvider AccessUserProvider) *Bot {
	bot, err := tgbotapi.NewBotAPI(kernel.Config().BotToken)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return &Bot{
		bot:                bot,
		cfg:                u,
		kernel:             kernel,
		accessUserProvider: accessUserProvider,
		Command:            make(chan *tgbotapi.Message),
		Message:            make(chan *tgbotapi.Message),
	}
}

func (b *Bot) StartListen() {
	updates := b.bot.GetUpdatesChan(b.cfg)

	b.kernel.Log().Debug("start listen bot command and message")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !b.accessUserProvider.IsAccess(update.Message.Chat.ID) {
			b.kernel.Log().Info(fmt.Sprintf("not access for user: %d", update.Message.Chat.ID))
			continue
		}

		if update.Message.IsCommand() {
			b.Command <- update.Message
			continue
		}

		b.Message <- update.Message
	}
}

func (b *Bot) SendMessage(msg tgbotapi.MessageConfig) error {
	_, err := b.bot.Send(msg)

	return err
}
