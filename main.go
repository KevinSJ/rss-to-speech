/*
*rss-to-tts A progrm to read rss articles to tts mp3s
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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"github.com/kevinsj/rss-to-podcast/internal/helpers"
	"github.com/kevinsj/rss-to-podcast/internal/types"
	"github.com/mmcdole/gofeed"
	"golang.org/x/sync/errgroup"
)

var FEEDS = [...]string{
	//"https://chinadigitaltimes.net/chinese/category/404%e6%96%87%e5%ba%93/feed",
	//"https://agora0.gitlab.io/blog/feed.xml",
	//"https://chinadigitaltimes.net/chinese/category/404%e6%96%87%e5%ba%93/feed",
	//"https://chinadigitaltimes.net/chinese/category/%e2%96%a3%e7%89%88%e9%9d%a2%e4%b8%8e%e9%82%ae%e4%bb%b6/level-2-article/feed",
	"https://rsshub.app/theinitium/channel/latest/zh-hans",
}

const CONCURRENT_WORKER = 5

func main() {

	configPath, _ := filepath.Abs("./config.yaml")

	config, err := helpers.ParseConfig(configPath)
	if err != nil {
		log.Fatalf("Unable to parse config file, error: %v", err)
	}

	credentialAbsPath, _ := filepath.Abs(config.CredentialPath)
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialAbsPath)

	fp := gofeed.NewParser()

	for _, v := range config.Feeds {
		log.Printf("v: %v\n", v)
		feed, err := fp.ParseURL(v)

		if err != nil {
			log.Fatalf("Error GET: %v\n", err)
		}
		//create folder based on RSS update date, this will be used to store all
		//generated mp3s.
		directory, err := types.CreateDirectory(*feed)
		if err != nil {
			panic(err)
		}

		if err := os.Chdir(*directory); err != nil {
			panic(err)
		}

		g := new(errgroup.Group)

		ctx := context.Background()

		client, err := texttospeech.NewClient(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()

		// ignore error
		createSpeechFromItems(feed, g, client, ctx, config.ItemSince)

		if err := g.Wait(); err != nil {
			log.Fatal(err.Error())
		}
	}
}

func collectFeedItems(feed *gofeed.Feed, g *errgroup.Group, client *texttospeech.Client, ctx context.Context, itemSince float64) {
	for _, _item := range feed.Items {
		if time.Since(*_item.PublishedParsed).Hours() <= itemSince {
			log.Printf("e.Title: %v\n", _item.Title)
			log.Printf("e.Published: %v\n", _item.Published)

			item := _item

			g.Go(func() error {
				if err := synthesizeSpeech(item, client, ctx); err != nil {
					return nil
				} else {
					return err
				}
			})
		}
	}
}

func createSpeechFromItems(feed *gofeed.Feed, g *errgroup.Group, client *texttospeech.Client, ctx context.Context, itemSince float64) {
	for _, _item := range feed.Items {
		if time.Since(*_item.PublishedParsed).Hours() <= itemSince {
			log.Printf("e.Title: %v\n", _item.Title)
			log.Printf("e.Published: %v\n", _item.Published)

			item := _item

			g.Go(func() error {
				if err := synthesizeSpeech(item, client, ctx); err != nil {
					return nil
				} else {
					return err
				}
			})
		}
	}
}

// This code is taken from sample google TTS code with some modification
// Source: https://cloud.google.com/text-to-speech/docs/libraries
func synthesizeSpeech(e *gofeed.Item, client *texttospeech.Client, ctx context.Context) error {
	log.Printf("Processing... %s", e.Title)
	audioContent := make([]byte, 0)

	reqs := types.GetSynthesizeSpeechRequests(e)

	for _, ssr := range reqs {
		resp, err := client.SynthesizeSpeech(ctx, ssr)
		if err != nil {
			return err
		}

		audioContent = append(audioContent, resp.AudioContent...)
	}

	// The resp's AudioContent is binary.
	filename := e.Title + ".mp3"
	if err := ioutil.WriteFile(strings.ReplaceAll(filename, "/", "\\/"), audioContent, 0644); err != nil {
		log.Fatal(err)
		return err
	}

	log.Printf("Audio content written to file: %v\n", filename)
	return nil
}
