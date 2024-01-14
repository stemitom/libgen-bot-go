package libgen

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	LibgenURL    = "https://libgen.is/search.php"
	LibgenAPIURL = "https://libgen.is/json.php"
)

type Book struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Year      string `json:"year"`
	Extension string `json:"extension"`
	MD5       string `json:"md5"`
}

type Search struct {
	Title  string
	Author string
}

type Utils struct {
	Client *http.Client
}

func (s *Search) searchParams() map[string]string {
	params := make(map[string]string)
	if s.Title != "" {
		params["title"] = s.Title
	}
	if s.Author != "" {
		params["author"] = s.Author
	}

	return params
}

func NewUtils() *Utils {
	return &Utils{
		Client: &http.Client{},
	}
}

func buildQueryParams(params map[string]string) string {
	var queryParams []string
	for key, value := range params {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(queryParams, "&")
}

func (u *Utils) Search(query Search, limit int) ([]string, error) {
	url := fmt.Sprintf("%s?%s", LibgenURL)
}

func SearchBook(query string) (*Book, error) {
	// search for book from libgen
	book := &Book{
		Title:  "Sample Book",
		Author: "John Doe",
		Year:   2002,
		URL:    "http://libgen.is/book/123456",
	}
	return book, nil
}
