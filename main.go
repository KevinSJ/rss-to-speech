/*
*rss-to-speech A program to read rss articles to tts mp3s
*Copyright © 2022 Kevin Jiang
*
*Permission is hereby granted, free of charge, to any person obtaining
*a copy of this software and associated documentation files (the "Software"),
*to deal in the Software without restriction, including without limitation
*the rights to use, copy, modify, merge, publish, distribute, sublicense,
*and/or sell copies of the Software, and to permit persons to whom the
*Software is furnished to do so, subject to the following conditions:
*
*The above copyright notice and this permission notice shall be included
*in all copies or substantial portions of the Software.
*
*THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
*EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
*OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
*IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
*DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
*TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
*OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"sync"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"github.com/KevinSJ/rss-to-podcast/internal/config"
	"github.com/KevinSJ/rss-to-podcast/internal/pkg/rss"
	"github.com/KevinSJ/rss-to-podcast/internal/pkg/worker"
	"github.com/mmcdole/gofeed"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
)

const FEED_RETRY_CNT = 5

func main() {
	logger := log.New(os.Stdout, "[info] ", log.Ldate|log.Ltime)

	defer logger.Printf("Done processing all feeds")
	configFile := flag.String("c", "./config.yaml", "config file of rss-to-speech")
	flag.Parse()
	config, err := config.NewConfig(*configFile)
	if err != nil {
		logger.Fatalf("Unable to parse config file, error: %v", err)
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", config.CredentialPath)

	fp := gofeed.NewParser()
	g := new(errgroup.Group)
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	var wg sync.WaitGroup

	workerGroup := worker.NewWorkerGroup(config, &wg, client, ctx)

	for _, _v := range config.Feeds {
		v := _v
		g.Go(func() error {
			logger.Printf("feed: %v\n", v)
			feed := getFeedWithRetry(fp, v)

			if feed == nil {
				logger.Printf("Fail to fetch feed: %v \n", v)
				return nil
			}

			hasValidItems := slices.IndexFunc(feed.Items, func(item *gofeed.Item) bool {
				return time.Since(item.PublishedParsed.Local()).Hours() <= config.ItemSince
			})

			if hasValidItems == -1 {
				logger.Printf("feed: %v has no valid item.\n", v)
				return nil
			}

			// create folder based on RSS update date, this will be used to store all
			// generated mp3s.
			dir, err := rss.CreateDirectory(*feed)
			if err != nil {
				logger.Printf("error: %v", err)
				return err
			}

			workerGroup.CreateSpeechFromItems(feed, dir)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		logger.Fatal(err.Error())
	}

	workerGroup.Close()
	wg.Wait()
}

func getFeedWithRetry(fp *gofeed.Parser, v string) *gofeed.Feed {
	var feed *gofeed.Feed = nil
	var err error = nil

	for i := 0; i < FEED_RETRY_CNT; i++ {
		if i > 0 {
			log.Printf("Retry due to Error GET: %v. \n", err)
			time.Sleep(2000)
		}

		feed, err = fp.ParseURL(v)
		if err == nil && feed != nil {
			return feed
		}
	}

	return feed
}
