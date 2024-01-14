package libgen

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
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
	url := fmt.Sprintf("%s?%s", LibgenURL, buildQueryParams(query.searchParams()))
	res, err := u.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	var ids []string
	doc.Find()
}
