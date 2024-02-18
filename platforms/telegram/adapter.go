package telegram

import (
	"log"
	"strings"

	"libgen-bot/services/libgen"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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
		tb.SendMessage(message.Chat.ID, "Please provide a search query. Example: <i>/search The Hobbits</i>")
		return
	}

	tb.SendMessage(message.Chat.ID, "🤖 Loading...")
	books, err := tb.LibGen.GetBooks(query)
	if err != nil {
		tb.SendMessage(message.Chat.ID, "Mmm, something went bad while searching for books. Try again later...")
		return
	}

	if len(books) == 0 {
		tb.SendMessage(message.Chat.ID, "Sorry, I don't have any result for that...")
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

// CallbackHandler handles callback queries from users.
// It retrieves book details based on the callback data and sends an edited message
// with the book details and a URL keyboard for book download link.
func (tb *TelegramBot) CallbackHandler(query tgbotapi.CallbackQuery) error {
	chatID := query.Message.Chat.ID

	ids := strings.Split(query.Data, ",")
	if len(ids) != 1 {
		return tb.sendEditMessage(chatID, query.Message.MessageID, "💥")
	}

	idArray := []string{ids[0]}
	books, err := tb.LibGen.GetBooksByIDs(idArray)
	if err != nil {
		return tb.sendEditMessage(chatID, query.Message.MessageID, "💥")
	}

	book := books[0]
	urlKeyboard := makeURLKeyboard(book.MD5URL())

	msg := tgbotapi.NewEditMessageText(chatID, query.Message.MessageID, book.Pretty())
	msg.ParseMode = "html"
	msg.ReplyMarkup = &urlKeyboard
	_, err = tb.Bot.Send(msg)
	if err != nil {
		return err
	}

	return nil
}

// sendEditMessage sends an edited message to the specified chat with the given text.
// It returns any error encountered during the sending process.
func (tb *TelegramBot) sendEditMessage(chatID int64, messageID int, text string) error {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	_, err := tb.Bot.Send(msg)
	if err != nil {
		log.Printf("Error sending edit message to %d: %v", chatID, err)
	}
	return err
}
