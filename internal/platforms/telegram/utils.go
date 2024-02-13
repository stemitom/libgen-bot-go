package telegram

import (
	"fmt"
	"libgen-bot/internal/services/libgen"
	"net/url"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// makeMessage creates a message string from a slice of books
func makeMessage(books []libgen.Book) string {
	var msg string
	for i, b := range books {
		msg += fmt.Sprintf("%s\n", b.PrettyWithIndex(i+1))
	}
	return msg
}

// makeURLKeyboard creates an inline keyboard with a single button linking to a given URL
func makeURLKeyboard(urlString string) tgbotapi.InlineKeyboardMarkup {
	url, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}

	button := tgbotapi.NewInlineKeyboardButtonURL("Download", url.String())
	row := []tgbotapi.InlineKeyboardButton{button}
	keyboard := [][]tgbotapi.InlineKeyboardButton{row}

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// makeKeyboard creates an inline keyboard with callback buttons for each book
func makeKeyboard(books []libgen.Book) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for i, book := range books {
		button := tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(i+1), book.ID)
		row := []tgbotapi.InlineKeyboardButton{button}
		keyboard = append(keyboard, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}
