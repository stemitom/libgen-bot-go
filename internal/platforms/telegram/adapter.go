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

const (
	maxButtonsPerRow = 5 // Maximum number of buttons per row in a keyboard
)

type TelegramBot struct {
	Bot     *tgbotapi.BotAPI
	Updates tgbotapi.UpdatesChannel
	LibGen  *libgen.LibGenClient
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

	l := libgen.NewLibGenClient()
	return &TelegramBot{Bot: bot, LibGen: l}, nil
}

// OnMessage sets up a handler for incoming messages
func (tb *TelegramBot) OnMessage(handler Handler) {
	updates := tgbotapi.NewUpdate(0)
	updates.Timeout = 60
	var err error
	tb.Updates, err = tb.Bot.GetUpdatesChan(updates)
	if err != nil {
		log.Println("Error getting updates:", err)
		return
	}

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
	if _, err := tb.Bot.Send(msg); err != nil {
		log.Printf("Error sending message to %d: %v", chatID, err)
	}
}

// handleStartCommand handles the "/start" command.
func (tb *TelegramBot) handleStartCommand(message *Message) {
	tb.SendMessage(message.Chat.ID, "Welcome to the VivioMagus Bot! Use /help to see available commands.")
}

func (tb *TelegramBot) handleHelpCommand(message *Message) {
	tb.SendMessage(message.Chat.ID, "Available commands:\n"+
		"/start - Start the bot\n"+
		"/search - Search for books\n"+
		"/help - Show this help message")
}

// handleSearchCommand handles the "/search" command.
func (tb *TelegramBot) handleSearchCommand(message *Message) {
	query := strings.TrimSpace(strings.TrimPrefix(message.Text, "/search"))
	if query == "" {
		tb.SendMessage(message.Chat.ID, "Please provide a search query. Example: /search The Hobbits")
		return
	}

	ids, err := tb.LibGen.Search(query, 10)
	if err != nil {
		log.Println("Error searching for books:", err)
		tb.SendMessage(message.Chat.ID, "An error occurred while searching for books.")
		return
	}

	books, err := tb.LibGen.GetBooks(ids)
	if err != nil {
		log.Println("Error getting book information:", err)
		tb.SendMessage(message.Chat.ID, "An error occurred while getting book information.")
		return
	}

	if len(books) == 0 {
		tb.SendMessage(message.Chat.ID, "No books found for your query.")
		return
	}

	msgText := makeMessage(books)
	tb.SendMessage(message.Chat.ID, msgText)

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
	case "help":
		tb.handleHelpCommand(message)
	default:
		tb.SendMessage(message.Chat.ID, "Unknown command. Type /help for a list of available commands.")
	}
}

// HandleIncomingMessage is a general handler for all incoming messages.
func (tb *TelegramBot) HandleIncomingMessage(message *Message) {
	if message.IsCommand() {
		tb.HandleCommand(message, message.Command())
	} else {
		tb.SendMessage(message.Chat.ID, "I don't know how to handle this type of message.")
	}
}

// makeMessage creates a message string from a slice of Books.
func makeMessage(books []libgen.Book) string {
	var msg strings.Builder
	for i, b := range books {
		msg.WriteString(fmt.Sprintf("%d. %s\n", i+1, b.Title))
	}
	return msg.String()
}

// makeURLKeyboard creates a keyboard with a URL button.
func makeURLKeyboard(urlStr string) tgbotapi.InlineKeyboardMarkup {
	url, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return tgbotapi.InlineKeyboardMarkup{}
	}
	button := tgbotapi.NewInlineKeyboardButtonURL("Download", url.String())
	keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{button})
	return keyboard
}

// makeKeyboard creates a keyboard for selecting books.
func makeKeyboard(books []libgen.Book) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for i, book := range books {
		buttonText := strconv.Itoa(i+1) + ". " + book.Title
		if len(buttonText) > 40 { // Truncate long titles
			buttonText = buttonText[:37] + "..."
		}
		callbackData := "book:" + book.ID
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)

		if (i+1)%maxButtonsPerRow == 0 {
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
		currentRow = append(currentRow, button)
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
