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

	// Speed of Synthesized Speech
	SpeechSpeed float64
}

type WorkerGroup struct {
	config  *config.Config
	channel *chan *WorkerRequest
	client  *texttospeech.Client
}

func (w *WorkerGroup) Close() {
	close(*w.channel)
}

func (w *WorkerGroup) CreateSpeechFromItems(feed *gofeed.Feed, direcory *string) {
	log.Printf("feed.Title: %v\n", feed.Title)

	itemSize := func(size int, limit int) int {
		if size > limit {
			return limit
		}

		return size
	}(len(feed.Items), w.config.MaxItemPerFeed)

	isInRange := func(itemPublishTime *time.Time) bool {
		return time.Since((*itemPublishTime).Local()).Hours() <= w.config.ItemSince
	}

	for _, item := range feed.Items[:itemSize] {
		if isInRange(item.PublishedParsed) {
			*w.channel <- &WorkerRequest{
				Item:            item,
				LanguageCode:    feed.Language,
				Directory:       *direcory,
				UseNaturalVoice: w.config.UseNaturalVoice,
				SpeechSpeed:     w.config.SpeechSpeed,
			}
		}
	}
}

// This code is taken from sample google TTS code with some modification
// Source: https://cloud.google.com/text-to-speech/docs/libraries
func processSpeechGeneration(wg *sync.WaitGroup, client *texttospeech.Client, workerItems *chan *WorkerRequest, ctx context.Context) error {
	defer wg.Done()

	for workerItem := range *workerItems {
		feedItem := workerItem.Item

		fileName := strings.ReplaceAll(feedItem.Title, "/", "\\/") + ".mp3"
		filepath, _ := filepath.Abs(workerItem.Directory + "/" + fileName)

		if _, err := os.Stat(filepath); err == nil {
			log.Printf("File exists at path: %s\n, skip generating", filepath)
			return nil
		}

		log.Printf("Start procesing %v ", feedItem.Title)

		speechRequests := rss.GetSynthesizeSpeechRequests(feedItem, workerItem.LanguageCode, workerItem.UseNaturalVoice, workerItem.SpeechSpeed)
		audioContent := make([]byte, 0)

		for _, ssr := range speechRequests {
			resp, err := client.SynthesizeSpeech(ctx, ssr)
			if err != nil {
				log.Printf("Encountered error when calling google text to speech service: %v\n", err)
				return err
			}

			audioContent = append(audioContent, resp.AudioContent...)
		}

		if err := os.WriteFile(filepath, audioContent, 0o755); err != nil {
			log.Printf("err writing synthesized file: %v\n", err)
			return err
		}

		fileTime := func(item *gofeed.Item) time.Time {
			if item.UpdatedParsed != nil {
				return item.UpdatedParsed.Local()
			}
			if item.PublishedParsed != nil {
				return item.PublishedParsed.Local()
			}
			return time.Now().Local()
		}(feedItem)

		if err := os.Chtimes(filepath, fileTime, fileTime); err != nil {
			log.Printf("err: %v\n", err)
			return err
		}

		log.Printf("Finished Processing: %v, written to %v\n", feedItem.Title, filepath)
	}

	return nil
}

func NewWorkerGroup(config *config.Config, wg *sync.WaitGroup, client *texttospeech.Client, ctx context.Context) *WorkerGroup {
	channelSize := config.MaxItemPerFeed * len(config.Feeds)
	work := make(chan *WorkerRequest, channelSize)
	wg.Add(channelSize)

	for i := 0; i < config.ConcurrentWorkers; i++ {
		go processSpeechGeneration(wg, client, &work, ctx)
	}

	return &WorkerGroup{
		config:  config,
		channel: &work,
		client:  client,
	}
}
