package bot

import "libgen-bot/pkg/telegram"

type Bot struct {
	Telegram *telegram.TelegramBot
}

func NewBot() (*Bot, error) {
	telegramBot, err := telegram.NewTelegramBot()
	if err != nil {
		return nil, err
	}

	return &Bot{Telegram: telegramBot}, nil
}

func (b *Bot) Start() {
	go b.Telegram.Start()
	// We will add the rest here

	select {}
}
