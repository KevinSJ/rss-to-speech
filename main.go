/*
*rss-to-tts A program to read rss articles to tts mp3s
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
	"log"
	"os"
	"path/filepath"
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

func main() {
	configPath, _ := filepath.Abs("./config.yaml")
	config, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatalf("Unable to parse config file, error: %v", err)
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

	work := *worker.NewWorkerGroup(config.ConcurrentWorkers, &wg, config.MaxItemPerFeed*len(config.Feeds), client, ctx)

	for _, _v := range config.Feeds {
		v := _v
		g.Go(func() error {
			log.Printf("feed: %v\n", v)
			feed, err := fp.ParseURL(v)
			if err != nil {
				log.Fatalf("Error GET: %v\n", err)
			}

			hasValidItems := slices.IndexFunc(feed.Items, func(item *gofeed.Item) bool {
				return time.Since(item.PublishedParsed.Local()).Hours() <= config.ItemSince
			})

			if hasValidItems == -1 {
				return nil
			}

			// create folder based on RSS update date, this will be used to store all
			// generated mp3s.
			dir, err := rss.CreateDirectory(*feed)
			if err != nil {
				log.Panicf("error: %v", err)
			}

			worker.CreateSpeechFromItems(feed, config, &work, dir)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		log.Fatal(err.Error())
	}

	close(work)
	wg.Wait()

	log.Printf("Done processing all feeds")
}
