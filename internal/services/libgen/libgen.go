package libgen

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	LibgenURL    = "https://libgen.is/search.php"
	LibgenAPIURL = "https://libgen.is/json.php"
)

type Book struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Year        string `json:"year"`
	Extension   string `json:"extension"`
	Filesize    string `json:"filesize"`
	Pages       string `json:"pages"`
	MD5         string `json:"md5"`
	Publisher   string `json:"publisher"`
	Language    string `json:"language"`
	PageURL     string `json:"pageUrl"`
	CoverURL    string `json:"coverUrl"`
	DownloadURL string `json:"downloadUrl"`
}

// Pretty returns a formatted string representation of the book.
func (b *Book) Pretty() string {
	return fmt.Sprintf("<b>%s</b>\n\n👤 %s\nFormat: %s\n", b.Title, b.Author, b.Extension)
}

// PrettyWithIndex returns a formatted string representation of the book with an index.
func (b *Book) PrettyWithIndex(index int) string {
	return fmt.Sprintf("%d. <b>%s</b>\n👤 %s\nYear: %s, Type: %s\n", index, b.Title, b.Author, b.Year, b.Extension)
}

// MD5URL returns the URL for downloading the book by MD5.
func (b *Book) MD5URL() string {
	return fmt.Sprintf("https://library.lol/main/%s", b.MD5)
}

// String returns a string representation of the book.
func (b *Book) String() string {
	return fmt.Sprintf("Book(%s, %s, %s)", b.Title, b.Author, b.MD5)
}

type LibGenClient struct {
	Client  *http.Client
	BaseURL string
	APIURL  string
}

func NewLibGenClient() *LibGenClient {
	return &LibGenClient{
		BaseURL: LibgenURL, APIURL: LibgenAPIURL,
		Client: &http.Client{
			Timeout: time.Second * 60,
			Transport: &http.Transport{
				Proxy:           http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				MaxIdleConns:    10,
				IdleConnTimeout: 90 * time.Second,
			},
		},
	}
}

func buildQueryParams(params map[string]string) string {
	var queryParams []string
	for key, value := range params {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(queryParams, "&")
}

func (l *LibGenClient) Search(queryText string, limit int) ([]string, error) {
	url := fmt.Sprintf("%s?req=%s", LibgenURL, strings.ReplaceAll(queryText, " ", "+"))
	res, err := l.Client.Get(url)
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
	counter := 0
	doc.Find("[valign='top']").Each(func(_ int, s *goquery.Selection) {
		if counter > 0 && len(ids) < limit {
			id := s.Children().First().Text()
			ids = append(ids, id)
		}
		counter++
	})
	return ids, nil
}

func (l *LibGenClient) GetBooksByIDs(ids []string) ([]Book, error) {
	url := fmt.Sprintf("%s?%s", l.APIURL, buildQueryParams(map[string]string{
		"fields": "id,title,author,year,extension,md5",
		"ids":    strings.Join(ids, ","),
	}))

	res, err := l.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var books []Book
	err = json.NewDecoder(res.Body).Decode(&books)
	if err != nil {
		return nil, err
	}

	return books, nil
}

func (l *LibGenClient) GetBooks(query string) ([]Book, error) {
	ids, err := l.Search(query, 5)
	if err != nil {
		return nil, err
	}

	if len(ids) > 0 {
		return l.GetBooksByIDs(ids)
	}

	return nil, fmt.Errorf("no book IDs found")
}
