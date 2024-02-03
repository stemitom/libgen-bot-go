package telegram

import (
	"fmt"
	"libgen-bot/internal/services/libgen"
	"log"
	"net/url"
	"strconv"
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

// Handler is a function signature for handling incoming messages.
type Handler func(msg *Message, tb *TelegramBot)

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

// handleStartCommand handles the "/start" command.
func (tb *TelegramBot) handleStartCommand(message *Message) {
	response := "Welcome to the VivioMagus Bot! Use /help to see available commands."
	tb.SendMessage(message.Chat.ID, response)
}

// handleSearchCommand handles the "/search" command.
func (tb *TelegramBot) handleSearchCommand(message *Message) {
	query := strings.TrimSpace(strings.TrimPrefix(message.Text, "/search"))
	if query == "" {
		tb.SendMessage(message.Chat.ID, "Please provide a search query. Example: /search The Hobbits")
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

	if len(books) == 0 {
		tb.SendMessage(message.Chat.ID, "No books found for your query.")
		return
	}

	// Use makeMessage to create a message with book titles
	msgText := makeMessage(books)
	tb.SendMessage(message.Chat.ID, msgText)

	// Use makeKeyboard to create a keyboard with book options
	keyboard := makeKeyboard(books)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Select a book:")
	msg.ReplyMarkup = keyboard
	tb.Bot.Send(msg)
}

// HandleCommand handles commands received from users.
func (tb *TelegramBot) HandleCommand(message *Message, command string) {
	switch command {
	case "start":
		tb.handleStartCommand(message)
	case "search":
		tb.handleSearchCommand(message)
	default:
		tb.SendMessage(message.Chat.ID, "Unknown command. Type /help for a list of available commands.")
	}
}

// handleTextMessage handles general text messages.
func (tb *TelegramBot) handleTextMessage(message *Message) {
	response := "I received a text message. Use /help to see available commands."
	tb.SendMessage(message.Chat.ID, response)
}

// HandleIncomingMessage is a general handler for all incoming messages.
func (tb *TelegramBot) HandleIncomingMessage(message *Message) {
	switch {
	case message.IsCommand():
		tb.HandleCommand(message, message.Command())
	default:
		tb.SendMessage(message.Chat.ID, "I don't know how to handle this type of message.")
	}
}

// makeMessage creates a message string from a slice of Books.
func makeMessage(books []libgen.Book) string {
	msg := ""
	for i, b := range books {
		msg += fmt.Sprintf("%d. %s\n", i+1, b.Title)
	}
	return msg
}

// makeURLKeyboard creates a keyboard with a URL button.
func makeURLKeyboard(urlStr string) tgbotapi.InlineKeyboardMarkup {
	url, _ := url.Parse(urlStr)
	button := tgbotapi.NewInlineKeyboardButtonURL("Download", url.String())

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(button),
	)
	return keyboard
}

// makeKeyboard creates a keyboard from a slice of Books.
func makeKeyboard(books []libgen.Book) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for i, book := range books {
		if i%5 == 0 {
			keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{})
		}
		button := tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(i+1), book.ID)
		rowIndex := i / 5
		keyboard[rowIndex] = append(keyboard[rowIndex], button)
	}

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}
