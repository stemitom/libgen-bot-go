package main

import (
	"log"

	"libgen-bot/internal/bot"
)

func main() {
	b, err := bot.NewBot()
	if err != nil {
		log.Panic(err)
	}
	b.Start()
}
