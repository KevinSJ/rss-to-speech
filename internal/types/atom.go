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
	"os"

	"github.com/kevinsj/rss-to-podcast/internal/helpers"
	"github.com/mmcdole/gofeed"

	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

// Create directory based on feed updated date
func CreateDirectory(f gofeed.Feed) (*string, error) {
	directory := f.Title + " " + f.UpdatedParsed.Format("2006-01-02")

	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, err
	}

	return &directory, nil
}

func GetSynthesizeSpeechRequests(e *gofeed.Item) []*texttospeechpb.SynthesizeSpeechRequest {
	text := e.Title + "\n\n"
	if len(e.Content) > 0 {
		text += helpers.StripHtmlTags(e.Content)
	} else if len(e.Description) > 0 {
		text += helpers.StripHtmlTags(e.Description)
	}
	splitedText := helpers.ChunksByte(text, 5000)

	synthesizeRequest := make([]*texttospeechpb.SynthesizeSpeechRequest, 0)

	for _, v := range splitedText {

		req := texttospeechpb.SynthesizeSpeechRequest{
			// Set the text input to be synthesized.
			Input: &texttospeechpb.SynthesisInput{
				InputSource: &texttospeechpb.SynthesisInput_Text{Text: v},
			},
			// Build the voice request, select the language code ("en-US") and the SSML
			// voice gender ("neutral").
			Voice: &texttospeechpb.VoiceSelectionParams{
				LanguageCode: "zh-CN",
				SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
			},
			// Select the type of audio file you want returned.
			AudioConfig: &texttospeechpb.AudioConfig{
				AudioEncoding: texttospeechpb.AudioEncoding_MP3,
			},
		}

		synthesizeRequest = append(synthesizeRequest, &req)
	}
	return synthesizeRequest
}
