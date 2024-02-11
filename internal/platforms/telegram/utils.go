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

// makeMessage creates a message string from a slice of Books including details.
func makeMessage(books []libgen.Book) string {
	var msg strings.Builder
	for _, b := range books {
		msg.WriteString(fmt.Sprintf("Title: %s\nFilesize: %s\nExtension: %s\nDownloadUrl: %s\n", b.Title, b.Filesize, b.Extension, b.DownloadURL))
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

// makeKeyboard creates a keyboard for selecting books to download.
func makeKeyboard(books []libgen.Book) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for i := range books {
		buttonText := "Download Book " + strconv.Itoa(i+1)
		callbackData := "select:" + strconv.Itoa(i)
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
