package rss

import (
	"log"
	"net/url"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
)

func getUpdatedDate(f gofeed.Feed) *time.Time {
	if f.UpdatedParsed != nil {
		return f.UpdatedParsed
	}
	if f.PublishedParsed != nil {
		return f.PublishedParsed
	}

	currentDate := time.Now()
	return &currentDate
}

func CreateDirectory(f gofeed.Feed) (dir *string, err error) {
	validLink := f.FeedLink
	if len(validLink) == 0 {
		validLink = f.Link
	}

	link, err := url.Parse(validLink)
	directory := link.Hostname()

	if err := os.MkdirAll(directory, 0o755); err != nil {
		log.Printf("Failed to create directory for feed: %v", f.Title)
		return nil, err
	}

	return &directory, nil
}
