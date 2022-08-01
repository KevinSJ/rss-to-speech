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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"github.com/kevinsj/rss-to-podcast/internal/types"
	"github.com/mmcdole/gofeed"
	"golang.org/x/sync/errgroup"
)

const FEED_URL = "https://agora0.gitlab.io/blog/feed.xml"
//const FEED_URL = "https://chinadigitaltimes.net/chinese/category/404%e6%96%87%e5%ba%93/feed"

func main() {

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(FEED_URL)
	if err != nil {
		log.Printf("Error GET: %v\n", err)
		log.Panic(err)
	}

	g := new(errgroup.Group)

	//create folder based on RSS update date, this will be used to store all
	//generated mp3s.
	directory, err := types.CreateDirectory(*feed)
	if err != nil {
		panic(err)
	}

	if err := os.Chdir(*directory); err != nil {
		panic(err)
	}

	allEntries := feed.Items

	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	for _, v := range allEntries {
		if time.Since(*v.PublishedParsed).Hours() <= 72.0 {
			fmt.Printf("e.Title: %v\n", v.Title)
			fmt.Printf("e.Published: %v\n", v.Published)

			v := v

			g.Go(func() error {
				if err := synthesizeSpeech(v, client, ctx); err != nil {
					return err
				}
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		log.Fatal(err.Error())
	}
}

//This code is taken from sample google TTS code with some modification
//Source: https://cloud.google.com/text-to-speech/docs/libraries
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
