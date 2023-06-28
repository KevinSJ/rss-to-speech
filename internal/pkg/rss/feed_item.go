package rss

import (
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/KevinSJ/rss-to-podcast/internal/pkg/tool"
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

func getSanitizedContentChunks(item *gofeed.Item) (textchunks []string) {
	content := item.Title + "\n\n"

	if len(item.Content) > 0 {
		content += tool.StripHtmlTags(item.Content)
	} else if len(item.Description) > 0 {
		content += tool.StripHtmlTags(item.Description)
	}

	return tool.ChunksByte(content, 5000)
}

func GetSynthesizeSpeechRequests(item *gofeed.Item, lang string, UseNaturalVoice bool) []*texttospeechpb.SynthesizeSpeechRequest {
	itemContent := getSanitizedContentChunks(item)

	if len(lang) == 0 {
		lang = tool.GuessLanguageByUnicode(item.Title)
	}

	lang = tool.GetSanitizedLanguageCode(lang)

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
