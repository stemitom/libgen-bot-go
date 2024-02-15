package libgen

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
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

// Pretty returns a formatted string representation of the book
func (b *Book) Pretty() string {
	return fmt.Sprintf("<b>%s</b>\n\n👤 %s\nFormat: %s\n", b.Title, b.Author, b.Extension)
}

// PrettyWithIndex returns a formatted string representation of the book with index
func (b *Book) PrettyWithIndex(index int) string {
	return fmt.Sprintf("%d. <b>%s</b>\n👤 %s\nYear: %s, Type: %s\n", index, b.Title, b.Author, b.Year, b.Extension)
}

// MD5URL returns the URL for downloading the book by MD5
func (b *Book) MD5URL() string {
	return fmt.Sprintf("https://library.lol/main/%s", b.MD5)
}

// String returns a string representation of the book
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

// Search searches for books based on the provided query text and returns their IDs.
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

// GetBooks retrieves books by their IDs concurrently.
func (l *LibGenClient) GetBooks(ids []string) ([]Book, error) {
	var (
		mu    sync.Mutex
		wg    sync.WaitGroup
		books []Book
		errs  []error
	)

	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			book, err := l.GetBookByID(id)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			books = append(books, book)
			mu.Unlock()
		}(id)
	}

	wg.Wait()

	if len(errs) > 0 {
		var errMsg strings.Builder
		for _, err := range errs {
			errMsg.WriteString(fmt.Sprintf("%v\n", err))
		}
		return nil, fmt.Errorf("failed to fetch some books:\n%s", errMsg.String())
	}

	return books, nil
}

// GetBookByID retrieves a book by its ID.
func (l *LibGenClient) GetBookByID(id string) (Book, error) {
	url := fmt.Sprintf("%s?%s", l.APIURL, buildQueryParams(map[string]string{
		"fields": JSONQuery,
		"ids":    id,
	}))

	res, err := l.Client.Get(url)
	if err != nil {
		return Book{}, fmt.Errorf("failed to fetch book %s: %w", id, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Book{}, fmt.Errorf("failed to fetch book %s: unexpected status code: %d", id, res.StatusCode)
	}

	var book Book
	if err := json.NewDecoder(res.Body).Decode(&book); err != nil {
		return Book{}, fmt.Errorf("failed to decode book %s response: %w", id, err)
	}

	return book, nil
}

// func (l *LibGenClient) GetBooks(ids []string) ([]Book, error) {
// 	url := fmt.Sprintf("%s?%s", l.APIURL, buildQueryParams(map[string]string{
// 		"fields": JSONQuery,
// 		"ids":    strings.Join(ids, ","),
// 	}))
// 	res, err := l.Client.Get(url)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch books: %w", err)
// 	}
// 	defer res.Body.Close()
//
// 	if res.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("failed to fetch books: unexpected status code: %d", res.StatusCode)
// 	}
//
// 	var books []Book
// 	if err := json.NewDecoder(res.Body).Decode(&books); err != nil {
// 		return nil, fmt.Errorf("failed to decode books response: %w", err)
// 	}
//
// 	return books, nil
// }
//
// func (l *LibGenClient) GetDownloadURL(book Book) (string, error) {
// 	url := book.MD5URL()
//
// 	res, err := l.Client.Get(url)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to fetch download URL: %w", err)
// 	}
// 	defer res.Body.Close()
//
// 	if res.StatusCode != http.StatusOK {
// 		return "", fmt.Errorf("failed to fetch download URL: unexpected status code: %d", res.StatusCode)
// 	}
//
// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
// 	if err != nil {
// 		return "", err
// 	}
//
// 	var downloadURL string
// 	doc.Find("#download a").Each(func(_ int, s *goquery.Selection) {
// 		downloadLink, exists := s.Attr("href")
// 		if exists {
// 			downloadURL = downloadLink
// 		}
// 	})
//
// 	if downloadURL == "" {
// 		return "", fmt.Errorf("download URL not found for book %s", book.Title)
// 	}
//
// 	return downloadURL, nil
// }
