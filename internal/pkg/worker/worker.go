package worker

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"github.com/KevinSJ/rss-to-podcast/internal/config"
	"github.com/KevinSJ/rss-to-podcast/internal/pkg/rss"
	"github.com/mmcdole/gofeed"
)

type WorkerRequest struct {
	// Item for this request
	Item *gofeed.Item

	// Directory to which the file wil write to
	Directory string

	// Language of the item
	LanguageCode string

	// Whether to use natural Voice
	UseNaturalVoice bool
}

func CreateSpeechFromItems(feed *gofeed.Feed, config *config.Config, work *chan *WorkerRequest, direcory *string) {
	log.Printf("feed.Title: %v\n", feed.Title)

	itemSize := func(size int, limit int) int {
		if size > limit {
			return limit
		}

		return size
	}(len(feed.Items), config.MaxItemPerFeed)

	isInRange := func(itemPublishTime *time.Time) bool {
		return time.Since((*itemPublishTime).Local()).Hours() <= config.ItemSince
	}

	for _, item := range feed.Items[:itemSize] {
		if isInRange(item.PublishedParsed) {
			*work <- &WorkerRequest{
				Item:            item,
				LanguageCode:    feed.Language,
				Directory:       *direcory,
				UseNaturalVoice: config.UseNaturalVoice,
			}
		}
	}
}

// This code is taken from sample google TTS code with some modification
// Source: https://cloud.google.com/text-to-speech/docs/libraries
func speechSynthesizeWorker(wg *sync.WaitGroup, client *texttospeech.Client, workerItems *chan *WorkerRequest, ctx context.Context) error {
	defer wg.Done()

	for workerItem := range *workerItems {
		feedItem := workerItem.Item

		sanitizedTitle := strings.ReplaceAll(feedItem.Title, "/", "\\/")
		filename := sanitizedTitle + ".mp3"
		filepath, _ := filepath.Abs(workerItem.Directory + "/" + filename)

		if _, err := os.Stat(filepath); err == nil {
			log.Printf("File exists at path: %s\n, skip generating", filepath)
			return nil
		}

		log.Printf("Start procesing %v ", feedItem.Title)

		speechRequests := rss.GetSynthesizeSpeechRequests(feedItem, workerItem.LanguageCode, workerItem.UseNaturalVoice)
		audioContent := make([]byte, 0)

		for _, ssr := range speechRequests {
			resp, err := client.SynthesizeSpeech(ctx, ssr)
			if err != nil {
				log.Printf("err: %v\n", err)
				return err
			}

			audioContent = append(audioContent, resp.AudioContent...)
		}

		if err := os.WriteFile(filepath, audioContent, 0o755); err != nil {
			log.Printf("err: %v\n", err)
			return err
		}

		log.Printf("Finished Processing: %v, written to %v\n", feedItem.Title, filepath)
	}

	return nil
}

func NewWorkerGroup(workerCount int, wg *sync.WaitGroup, channelSize int, client *texttospeech.Client, ctx context.Context) *chan *WorkerRequest {
	work := make(chan *WorkerRequest, channelSize)
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go speechSynthesizeWorker(wg, client, &work, ctx)
	}

	return &work
}
