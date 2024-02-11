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
	url := fmt.Sprintf("%s?req=%s", LibgenURL, strings.ReplaceAll(queryText, " ", "+"))
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

	downloadURLChan := make(chan string)
	errorsChan := make(chan error)

	for _, book := range books {
		go func(b Book) {
			downloadURL, err := l.GetDownloadURL(b)
			if err != nil {
				errorsChan <- err
				return
			}
			downloadURLChan <- downloadURL
		}(book)
	}

	for i := range books {
		select {
		case downloadURL := <-downloadURLChan:
			books[i].DownloadURL = downloadURL
		case err := <-errorsChan:
			return nil, err
		}
	}

	return books, nil
}

func (l *LibGenClient) GetDownloadURL(book Book) (string, error) {
	url := fmt.Sprintf("%s/%s", "https://library.lol/main", book.MD5)
	client := http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	res, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}

	var downloadURL string
	doc.Find("#download a").Each(func(_ int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists && downloadURL == "" {
			downloadURL = link
		}
	})

	if downloadURL == "" {
		return "", fmt.Errorf("download URL not found for book %s", book.Title)
	}

	return downloadURL, nil
}
