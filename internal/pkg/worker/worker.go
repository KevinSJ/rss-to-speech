package worker

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/KevinSJ/rss-to-podcast/internal/config"
	"github.com/KevinSJ/rss-to-podcast/internal/pkg/rss"
	"github.com/mmcdole/gofeed"
)

const SPEECH_SYNTHESIZE_RETRY_CNT = 5

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
	config    *config.Config
	channel   chan *WorkerRequest
	client    *texttospeech.Client
	waitGroup *sync.WaitGroup
}

func (w *WorkerGroup) Close() {
	defer log.Printf("Closing channel")
	close(w.channel)
}

func isInRange(itemPublishTime *time.Time, itemSince float64) bool {
	return time.Since((*itemPublishTime).Local()).Hours() <= itemSince
}

func (w *WorkerGroup) CreateSpeechFromItems(feed *gofeed.Feed, direcory *string) {
	log.Printf("feed.Title: %v\n", feed.Title)

	itemSize := func(size int, limit int) int {
		if size > limit {
			return limit
		}

		return size
	}(len(feed.Items), w.config.MaxItemPerFeed)

	feedLanguage := func(lang string) string {
		if strings.Contains(strings.ToLower(lang), "zh") {
			return "cmn-CN"
		}

		return lang
	}(feed.Language)

	itemCnt := 0

	for _, item := range feed.Items[:itemSize] {
		if isInRange(item.PublishedParsed, w.config.ItemSince) && itemCnt < itemSize {
			log.Printf("Adding item... title: %s", item.Title)
			w.channel <- &WorkerRequest{
				Item:            item,
				LanguageCode:    feedLanguage,
				Directory:       *direcory,
				UseNaturalVoice: w.config.UseNaturalVoice,
				SpeechSpeed:     w.config.SpeechSpeed,
			}
			itemCnt++
		}
	}
}

func fileExistsAndLog(path string) bool {
	if _, err := os.Stat(path); err == nil {
		log.Printf("File exists at path: %s\n, skip generating", path)
		return true
	}
	return false
}

// This code is taken from sample google TTS code with some modification
// Source: https://cloud.google.com/text-to-speech/docs/libraries
func processSpeechGeneration(wg *sync.WaitGroup, client *texttospeech.Client, workerItems chan *WorkerRequest, ctx context.Context) error {
	defer wg.Done()

	for workerItem := range workerItems {
		feedItem := workerItem.Item

		log.Printf("Start procesing %v ", feedItem.Title)
		hash := md5.New().Sum([]byte(feedItem.Title))
		hashString := hex.EncodeToString(hash[:])
		if hashSize := len(hashString); hashSize > 50 {
			hashString = hashString[:50]
		}
		fileName := workerItem.Directory + "/" + hashString
		filePath, _ := filepath.Abs(fileName + ".mp3")
		metaFilePath, _ := filepath.Abs(fileName + ".meta")
		legacyFilePath, _ := filepath.Abs(strings.ReplaceAll(feedItem.Title, "/", "\\/") + ".mp3")


		if fileExistsAndLog(legacyFilePath) || fileExistsAndLog(filePath) {
			continue
		}

		speechRequests := rss.GetSynthesizeSpeechRequests(feedItem, workerItem.LanguageCode, workerItem.UseNaturalVoice, workerItem.SpeechSpeed)
		audioContent := make([]byte, 0)

		for _, ssr := range speechRequests {
			var err error = nil
			var resp *texttospeechpb.SynthesizeSpeechResponse = nil

			for i := range SPEECH_SYNTHESIZE_RETRY_CNT {
				if i > 0 {
					log.Printf("While writing %v \n", feedItem.Title)
					log.Printf("Retry speech synthesize in 1 second due to error %v, count: %v\n", err, i)
					time.Sleep(time.Second)
				}

				resp, err = client.SynthesizeSpeech(ctx, ssr)
				if err != nil {
					log.Printf("Error Encountered, Response: %v\n", err.Error())
					continue
				}

				if len(resp.AudioContent) > 0 {
					audioContent = append(audioContent, resp.AudioContent...)
					break
				}
			}
			if err != nil {
				return err
			}
		}

		if err := os.WriteFile(metaFilePath, []byte("["+workerItem.Directory+"] "+feedItem.Title), 0o644); err != nil {
			log.Printf("err writing synthesized file: %v\n", err)
			return err
		}

		if err := os.WriteFile(filePath, audioContent, 0o644); err != nil {
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

		if err := os.Chtimes(filePath, fileTime, fileTime); err != nil {
			log.Printf("err: %v\n", err)
			return err
		}

		log.Printf("Finished Processing: %v, written to %v\n", feedItem.Title, filePath)
	}

	return nil
}

func NewWorkerGroup(config *config.Config, wg *sync.WaitGroup, client *texttospeech.Client, ctx context.Context) *WorkerGroup {
	channelSize := config.MaxItemPerFeed * len(config.Feeds)
	channel := make(chan *WorkerRequest, channelSize)

	workerSize := int(math.Min(float64(config.ConcurrentWorkers), float64(channelSize)))
	wg.Add(workerSize)

	for range workerSize {
		go processSpeechGeneration(wg, client, channel, ctx)
	}

	return &WorkerGroup{
		config:    config,
		channel:   channel,
		client:    client,
		waitGroup: wg,
	}
}
