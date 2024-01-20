package telegram

import (
	"fmt"
	"libgen-bot/internal/services/libgen"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramBot struct {
	Bot     *tgbotapi.BotAPI
	Updates tgbotapi.UpdatesChannel
}

type Message struct {
	*tgbotapi.Message
}

func NewTelegramBot(token string) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TelegramBot{Bot: bot}, nil
}

// OnMessage sets up a handler for incoming messages
func (tb *TelegramBot) OnMessage(handler Handler) {
	updates := tgbotapi.NewUpdate(0)
	updates.Timeout = 60
	tb.Updates, _ = tb.Bot.GetUpdatesChan(updates)
	for update := range tb.Updates {
		if update.Message == nil {
			continue
		}

		message := &Message{Message: update.Message}
		handler(message, tb)
	}
}

// SendMessage sends a message to the specified chatID
func (tb *TelegramBot) SendMessage(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := tb.Bot.Send(msg)
	if err != nil {
		log.Println("Error sending message:", err)
	}
}

// HandleCommand handles commands received from users.
func (tb *TelegramBot) HandleCommand(message *Message, command string) {
	switch command {
	case "/start":
		tb.handleStartCommand(message)
	case "/search":
		tb.handleSearchCommand(message)
	default:
		tb.SendMessage(message.Chat.ID, "Unknown command. Type /help for a list of available commands.")
	}
}

// handleStartCommand handles the "/start" command.
func (tb *TelegramBot) handleStartCommand(message *Message) {
	response := "Welcome to the Libgen Bot! Use /help to see available commands."
	tb.SendMessage(message.Chat.ID, response)
}

// handleSearchCommand handles the "/search" command.
func (tb *TelegramBot) handleSearchCommand(message *Message) {
	query := strings.TrimSpace(strings.TrimPrefix(message.Text, "/search"))
	if query == "" {
		tb.SendMessage(message.Chat.ID, "Please provide a search query. Example: /search Harry Potter")
		return
	}

	l := libgen.NewLibGenClient()
	ids, err := l.Search(query, 10)
	if err != nil {
		log.Println("Error searching for books:", err)
		tb.SendMessage(message.Chat.ID, "An error occurred while searching for books.")
		return
	}

	// Get book information
	books, err := l.GetBooks(ids)
	if err != nil {
		log.Println("Error getting book information:", err)
		tb.SendMessage(message.Chat.ID, "An error occurred while getting book information.")
		return
	}

	for _, book := range books {
		response := fmt.Sprintf("Book Information:\nTitle: %s\nAuthor: %s\nYear: %s\n",
			book.Title, book.Author, book.Year)
		tb.SendMessage(message.Chat.ID, response)
	}
}

// HandleIncomingMessage is a general handler for all incoming messages.
func (tb *TelegramBot) HandleIncomingMessage(message *Message) {
	// Handle different types of messages
	switch {
	case message.IsCommand():
		// Handle commands
		tb.HandleCommand(message, message.Command())
	default:
		// Handle other types of messages
		tb.SendMessage(message.Chat.ID, "I don't know how to handle this type of message.")
	}
}

// handleTextMessage handles general text messages.
func (tb *TelegramBot) handleTextMessage(message *Message) {
	response := "I received a text message. Use /help to see available commands."
	tb.SendMessage(message.Chat.ID, response)
}

// Handler is a function signature for handling incoming messages.
type Handler func(msg *Message, tb *TelegramBot)
