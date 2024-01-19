package bot

import (
	"fmt"
	"log"

	"libgen-bot/internal/libgen"
	"libgen-bot/internal/platforms/telegram"
)

type Bot struct {
	Telegram *telegram.TelegramBot
	Utils    *libgen.Utils
}

func NewBot() (*Bot, error) {
	telegramBot, err := telegram.NewTelegramBot()
	if err != nil {
		return nil, err
	}

	libgenUtils := libgen.NewUtils()

	return &Bot{
		Telegram: telegramBot,
		Utils:    libgenUtils,
	}, nil
}

func (b *Bot) handleIncomingMessage(msg *telegram.Message) {
	// Handle incoming messages

	// For now, let's just echo the received message
	response := fmt.Sprintf("You said: %s", msg.Text)
	b.Telegram.SendMessage(msg.Chat.ID, response)

	// Simulate a Libgen search
	ids, err := b.Utils.Search(libgen.Search{Title: msg.Text}, 5)
	if err != nil {
		log.Println("Error searching for books:", err)
		return
	}

	// Get book information
	books, err := b.Utils.GetBooks(ids)
	if err != nil {
		log.Println("Error getting book information:", err)
		return
	}

	// Send the book information
	for _, book := range books {
		response = fmt.Sprintf("Book Information:\nTitle: %s\nAuthor: %s\nYear: %s\n",
			book.Title, book.Author, book.Year)
		b.Telegram.SendMessage(msg.Chat.ID, response)
	}
}

func (b *Bot) Start() {
	// Start listening for incoming messages
	b.Telegram.OnMessage(b.handleIncomingMessage)

	// Keep the main goroutine running
	select {}
}
