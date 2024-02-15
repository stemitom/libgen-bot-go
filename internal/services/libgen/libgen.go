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
	/*
		// Create a channel to receive the download URLs
		urlsCh := make(chan string, len(books))
		defer close(urlsCh)

		// Create a wait group to wait for all goroutines to finish
		var wg sync.WaitGroup

		// Fetch the download URLs concurrently for each book
		for i := range books {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				log.Printf("Async downloading book %s", books[i].MD5URL())
				downloadURL, err := l.GetDownloadURL(books[i])
				if err != nil {
					log.Printf("Error fetching download URL for book %s: %v", books[i].Title, err)
					return
				}
				urlsCh <- downloadURL
			}(i)
		}

		// Wait for all goroutines to finish
		wg.Wait()

		// Assign the download URLs to the books
		for i := range books {
			books[i].DownloadURL = <-urlsCh
		}
	*/

	return books, nil
}

func (l *LibGenClient) GetDownloadURL(book Book) (string, error) {
	url := book.MD5URL()
	client := http.Client{
		Timeout: time.Second * 50,
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

	if res.StatusCode != http.StatusOK {
		log.Printf("Failed to fetch page: %s, status code: %d", url, res.StatusCode)
		return "", err
	}

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
		downloadLink, exists := s.Attr("href")
		if exists {
			downloadURL = downloadLink
		}
	})

	if downloadURL == "" {
		return "", fmt.Errorf("download URL not found for book %s", book.Title)
	}

	return downloadURL, nil
}
