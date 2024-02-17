package telegram

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"libgen-bot/internal/services/libgen"
)

type TelegramBot struct {
	Bot    *tgbotapi.BotAPI
	LibGen *libgen.LibGenClient
}

type Message struct {
	*tgbotapi.Message
}

func NewTelegramBot(token string) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	l := libgen.NewLibGenClient()
	return &TelegramBot{Bot: bot, LibGen: l}, nil
}

func (tb *TelegramBot) SendMessage(chatID int64, message string, parseMode ...string) {
	mode := "html"
	if len(parseMode) > 0 {
		mode = parseMode[0]
	}

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = mode
	if _, err := tb.Bot.Send(msg); err != nil {
		log.Printf("Error sending message to %d: %v", chatID, err)
	}
}

func (tb *TelegramBot) handleStartCommand(message *Message) {
	tb.SendMessage(message.Chat.ID, "Welcome to the VivioMagus Bot! Use /help to see available commands.")
}

func (tb *TelegramBot) handleHelpCommand(message *Message) {
	tb.SendMessage(message.Chat.ID, "Available commands:\n"+
		"/start - Start the bot\n"+
		"/search - Search for books\n"+
		"/help - Show this help message")
}

func (tb *TelegramBot) handleSearchCommand(message *Message) {
	query := strings.TrimSpace(strings.TrimPrefix(message.Text, "/search"))
	if query == "" {
		tb.SendMessage(message.Chat.ID, "Please provide a search query. Example: /search The Hobbits")
		return
	}

	books, err := tb.LibGen.GetBooks(query)
	if err != nil {
		tb.SendMessage(message.Chat.ID, "No books found for your query")
		return
	}

	text := makeMessage(books)
	keyboard := makeKeyboard(books)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "html"
	msg.ReplyMarkup = keyboard
	tb.Bot.Send(msg)
}

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

func (tb *TelegramBot) HandleIncomingMessage(message *Message) {
	if message.IsCommand() {
		tb.HandleCommand(message, message.Command())
	} else {
		tb.SendMessage(message.Chat.ID, "I don't know how to handle this type of message.")
	}
}
