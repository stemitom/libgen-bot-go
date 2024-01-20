package bot

import (
	"fmt"
	"libgen-bot/internal/platforms/telegram"
	"libgen-bot/internal/services/libgen"
	"log"
	"os"
)

type Bot struct {
	Telegram *telegram.TelegramBot
	Client   *libgen.LibGenClient
}

func NewBot() (*Bot, error) {
	token := os.Getenv("TELE_TOKEN")
	telegramBot, err := telegram.NewTelegramBot(token)
	if err != nil {
		return nil, err
	}

	libgenClient := libgen.NewLibGenClient()

	return &Bot{
		Telegram: telegramBot,
		Client:   libgenClient,
	}, nil
}

func (b *Bot) handleIncomingMessage(msg *telegram.Message) {
	// Handle incoming messages

	// For now, let's just echo the received message
	response := fmt.Sprintf("You said: %s", msg.Text)
	b.Telegram.SendMessage(msg.Chat.ID, response)

	// Simulate a Libgen search
	ids, err := b.Client.Search()
	if err != nil {
		log.Println("Error searching for books:", err)
		return
	}

	// Get book information
	books, err := b.Client.GetBooks(ids)
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
