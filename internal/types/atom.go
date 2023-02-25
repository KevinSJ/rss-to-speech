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
package types

import (
	"log"
	"os"
	"time"

	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/KevinSJ/rss-to-podcast/internal/helper"
	"github.com/mmcdole/gofeed"
)

var VOICE_NAME_MAP_WAVENET = map[string]string{
	"zh-CN": "cmn-CN-Wavenet-A",
	"en-US": "en-US-Neural2-C",
}

var VOICE_NAME_MAP_STANDARD = map[string]string{
	"zh-CN": "cmn-CN-Standard-D",
	"en-US": "en-US-Standard-C",
}

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

// Create directory based on feed updated date
func CreateDirectory(f gofeed.Feed) (dir *string, err error) {
	updatedDate := getUpdatedDate(f)

	directory := f.Title + "_" + updatedDate.Local().Format("2006-01-02")

	if err := os.MkdirAll(directory, 0o755); err != nil {
		log.Printf("Failed to create directory")
		return nil, err
	}

	return &directory, nil
}

func GetSynthesizeSpeechRequests(item *gofeed.Item, lang string, UseNaturalVoice bool) []*texttospeechpb.SynthesizeSpeechRequest {
	if len(lang) == 0 {
		lang = "zh-CN"
	}

	lang = helper.GetSanitizedLangCode(lang)

	itemContent := helper.GetSanitizedContentChunks(item)

	languageName := VOICE_NAME_MAP_STANDARD[lang]
	if UseNaturalVoice {
		languageName = VOICE_NAME_MAP_WAVENET[lang]
	}

	synthesizeRequest := make([]*texttospeechpb.SynthesizeSpeechRequest, 0)

	for _, v := range itemContent {

		req := texttospeechpb.SynthesizeSpeechRequest{
			// Set the text input to be synthesized.
			Input: &texttospeechpb.SynthesisInput{
				InputSource: &texttospeechpb.SynthesisInput_Text{Text: v},
			},
			// Build the voice request, select the language code ("en-US") and the SSML
			// voice gender ("neutral").
			Voice: &texttospeechpb.VoiceSelectionParams{
				LanguageCode: lang,
				Name:         languageName,
				SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
			},
			// Select the type of audio file you want returned.
			AudioConfig: &texttospeechpb.AudioConfig{
				AudioEncoding: texttospeechpb.AudioEncoding_MP3,
				SpeakingRate:  1.25,
			},
		}

		synthesizeRequest = append(synthesizeRequest, &req)
	}

	return synthesizeRequest
}
