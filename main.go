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
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"github.com/kevinsj/rss-to-podcast/internal/types"
	"golang.org/x/net/html/charset"
)

const FEED_URL = ""

func main() {
	resp, err := http.Get(FEED_URL)
	if err != nil {
		fmt.Printf("Error GET: %v\n", err)
		log.Panic(err)
	}
	defer resp.Body.Close()

	decoder := xml.NewDecoder(resp.Body)

	// This is needed as some of the feed is not in UTF-8
	decoder.CharsetReader = charset.NewReaderLabel
	rss := types.Feed{}

	if err := decoder.Decode(&rss); err != nil {
		panic(err)
	}

	//create folder based on RSS update date, this will be used to store all
	//generated mp3s.
	directory, err := rss.CreateDirectory()
	if err != nil {
		panic(err)
	}

	if err := os.Chdir(*directory); err != nil {
		panic(err)
	}

	fmt.Println(os.Getwd())
	allEntries := rss.Entries
	var wg sync.WaitGroup

	for _, v := range allEntries {
		publishedTime, _ := time.Parse(time.RFC3339, v.Published)
		if time.Since(publishedTime).Hours() <= 24.0 {
			fmt.Printf("e.Title: %v\n", v.Title)
			fmt.Printf("e.Published: %v\n", v.Published)

			/*
			 *filename := v.Title + ".mp3"
			 *fmt.Printf("path.Clean(filename): %v\n", path.Clean(filename))
			 */
			wg.Add(1)
			go func(e types.Entry, wg *sync.WaitGroup) {
				synthesizeSpeech(&e)
				defer wg.Done()
			}(v, &wg)
		}
	}
	wg.Wait()
}

//This code is copied from sample google TTS code with some modification
//Source: https://cloud.google.com/text-to-speech/docs/libraries
func synthesizeSpeech(e *types.Entry) {
	// Instantiates a client.
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	req := e.ToSynthesizeSpeechRequest()

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		log.Fatal(err)
	}

	// The resp's AudioContent is binary.
	filename := e.Title + ".mp3"
	err = ioutil.WriteFile(strings.ReplaceAll(filename, "/", "\\/"), resp.AudioContent, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Audio content written to file: %v\n", filename)

}
