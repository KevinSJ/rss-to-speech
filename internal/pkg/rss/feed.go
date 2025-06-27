package rss

import (
	"log"
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
	directory := f.Title

	if err := os.MkdirAll(directory, 0o755); err != nil {
		log.Printf("Failed to create directory for feed: %v", f.Title)
		return nil, err
	}

	return &directory, nil
}
