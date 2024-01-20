package main

import (
	"libgen-bot/internal/platforms/telegram"
	"log"
	"os"
)

func main() {
	// Retrieve env variable for telegram token
	botToken := os.Getenv("TOKEN")
	if botToken == "" {
		log.Fatal("TBOT_TOKEN environment variable not set")
	}

	bot, err := telegram.NewTelegramBot(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Bot is running...")
	bot.OnMessage(func(msg *telegram.Message, tb *telegram.TelegramBot) {
		tb.HandleIncomingMessage(msg)
	})
	// select {}
}
