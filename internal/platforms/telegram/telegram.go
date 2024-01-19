package telegram

import (
	"log"

	"libgen-bot/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramBot struct {
	Bot *tgbotapi.BotAPI
}

func NewTelegramBot() (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		return nil, err
	}

	return &TelegramBot{Bot: bot}, nil
}

func (tb *TelegramBot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tb.Bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Handle incoming messages
		// For now, let's just echo the recieved message
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		tb.Bot.Send(msg)
	}
}
