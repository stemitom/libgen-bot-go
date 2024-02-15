package main

import (
	"log"
	"os"

	"libgen-bot/internal/platforms/telegram"
)

func main() {
	// Retrieve env variable for telegram token
	botToken := os.Getenv("TOKEN")
	if botToken == "" {
		log.Fatal("TOKEN environment variable not set")
	}

	bot, err := telegram.NewTelegramBot(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting Bot..")
	bot.OnMessage(func(msg *telegram.Message, tb *telegram.TelegramBot) {
		tb.HandleIncomingMessage(msg)
	})
}
