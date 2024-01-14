package libgen

type Book struct {
	Title  string
	Author string
	Year   int
	URL    string
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
