package command

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ivankoTut/ping-url/internal/kernel"
	"github.com/ivankoTut/ping-url/internal/telegram"
	"github.com/ivankoTut/ping-url/internal/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	registrationCommand = "start"
	addUrlCommand       = "add_url"
	listUrlCommand      = "list_url"
)

var tracer trace.Tracer

type (
	// DialogChain это интерфес реализует цепочку диалога для консольной команжы
	DialogChain interface {
		SaveState(ctx context.Context, key string, state int) (bool, error)          // сохранить текущий шаг цепочки
		DialogExist(ctx context.Context, key string) (bool, error)                   // проверить есть ли цепочка диалога
		CurrentState(ctx context.Context, key string) (int, error)                   // получить текущий шаг цепочки
		SaveAnswer(ctx context.Context, key, state string, answer interface{}) error // сохранить ответ для текущего шага
		GetAnswer(ctx context.Context, key string) (map[string]string, error)        // получить данные по ответам
		DeleteDialog(ctx context.Context, key string) error                          // очистить данные по команде и пользователю
	}

	// HandlerCommand интерфейс которому должны удовлетворять все команды
	HandlerCommand interface {
		CommandName() string
		HelpText() string
		IsSupport(ctx context.Context, message *tgbotapi.Message) (bool, error)
		Run(ctx context.Context, message *tgbotapi.Message) (tgbotapi.MessageConfig, error)
		ClearData(ctx context.Context, message *tgbotapi.Message) error
	}

	// Command структура обертка для работы с всеми командами
	Command struct {
		kernel   *kernel.Kernel
		bot      *telegram.Bot
		commands []HandlerCommand
	}
)

func NewCommand(kernel *kernel.Kernel, bot *telegram.Bot, commands []HandlerCommand) *Command {
	return &Command{
		bot:      bot,
		commands: commands,
		kernel:   kernel,
	}
}

func (c *Command) ListenCommandAndMessage() {
	for {
		select {
		case message := <-c.bot.Message:
			c.runCommand(message)
		case command := <-c.bot.Command:
			c.runCommand(command)
		}
	}
}

func (c *Command) runCommand(message *tgbotapi.Message) {
	const op = "telegram.command.runCommand."

	cfg := c.kernel.Config().Jaeger
	tp, err := tracing.NewJaegerTraceProvider(cfg.Url, cfg.Name, cfg.Env)
	if err != nil {
		c.kernel.Log().Error(fmt.Sprintf("%s: ошибка инициализации Jaeger: %s", op, err))
	}

	tracer = tp.Tracer(cfg.Name)

	for _, handle := range c.commands {
		ctx := context.Background()
		is, err := handle.IsSupport(ctx, message)
		if err != nil {
			c.kernel.Log().Error(fmt.Sprintf("%s%s: error: %s", op, handle.CommandName(), err))
			continue
		}

		if is != true {
			continue
		}

		ctx, span := tracer.Start(ctx, fmt.Sprintf("message_from_%d", message.Chat.ID))
		if message.IsCommand() {
			span.SetAttributes(attribute.String("command start", handle.CommandName()))
			if err := handle.ClearData(ctx, message); err != nil {
				c.kernel.Log().Error(fmt.Sprintf("%s%s: %s", op, handle.CommandName(), err))
				span.RecordError(err)
			}
		}

		msg, err := handle.Run(ctx, message)
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("Error RUN", handle.CommandName()))
			c.kernel.Log().Error(fmt.Sprintf("%s%s: error: %s", op, handle.CommandName(), err))
		}

		err = c.bot.SendMessage(msg)
		if err != nil {
			span.SetAttributes(attribute.String("error send message", handle.CommandName()))
			span.RecordError(err)
			c.kernel.Log().Error(fmt.Sprintf("%s%s: %s", op, handle.CommandName(), err))
		}

		span.End()
	}
}