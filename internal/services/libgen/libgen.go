package libgen

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	ID          string
	Title       string
	Author      string
	Year        string
	Extension   string
	Filesize    string
	Pages       string
	MD5         string
	Publisher   string
	Language    string
	PageURL     string
	CoverURL    string
	DownloadURL string
}

type LibGenClient struct {
	BaseURL string
	APIURL  string
}

func NewLibGenClient() *LibGenClient {
	return &LibGenClient{BaseURL: LibgenURL, APIURL: LibgenAPIURL}
}

func buildQueryParams(params map[string]string) string {
	var queryParams []string
	for key, value := range params {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(queryParams, "&")
}

func (l *LibGenClient) Search(queryText string, limit int) ([]string, error) {
	url := fmt.Sprintf("%s?req=%s", LibgenURL, queryText)
	client := http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	res, err := client.Get(url)
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

	books, err := l.GetBooks(ids)
	if err != nil {
		return nil, err
	}
	for _, book := range books {
		log.Printf("Book: %+v\n", book)
	}
	return ids, nil
}

func (l *LibGenClient) GetBooks(ids []string) ([]Book, error) {
	url := fmt.Sprintf("%s?%s", l.APIURL, buildQueryParams(map[string]string{
		"fields": JSONQuery,
		"ids":    strings.Join(ids, ","),
	}))

	client := http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	res, err := client.Get(url)
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

func (l *LibGenClient) GetDownloadURL(book Book) {
}

// func (l *LibGenClient) GetValidDownloadURLs(mirrors []url.URL) url.URL {
// }
