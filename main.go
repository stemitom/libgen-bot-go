package main

import (
	"log"
	"os"

	"libgen-bot/platforms/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	botToken := os.Getenv("TOKEN")
	if botToken == "" {
		log.Fatal("TOKEN environment variable not set")
	}

	bot, err := telegram.NewTelegramBot(botToken)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Authorized on account %s", bot.Bot.Self.UserName)

	updatesConfig := tgbotapi.NewUpdate(0)
	updatesConfig.Timeout = 60
	updates, err := bot.Bot.GetUpdatesChan(updatesConfig)
	if err != nil {
		log.Println("Error getting updates:", err)
	}

	for update := range updates {
		if update.CallbackQuery != nil {
			bot.CallbackHandler(*update.CallbackQuery)
		} else if update.Message != nil {
			// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			message := &telegram.Message{Message: update.Message}
			bot.HandleIncomingMessage(message)
		}
	}
}
