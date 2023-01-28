package helper

import (
	"reflect"
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestGetSanitizedContent(t *testing.T) {
	type args struct {
		item *gofeed.Item
	}
	itemWithDescription := &gofeed.Item{
		Title:       "title",
		Description: "description",
	}
	itemWithContent := &gofeed.Item{
		Title:   "title",
		Content: "content",
	}
	itemWithHTMLContent := &gofeed.Item{
		Title:   "title",
		Content: "<p>content</p>",
	}
	tests := []struct {
		name           string
		args           args
		wantTextchunks []string
	}{
		{
			name:           "return sanitize content with description appended",
			args:           args{itemWithDescription},
			wantTextchunks: []string{itemWithDescription.Title + "\n\n" + itemWithDescription.Description},
		},
		{
			name:           "return sanitze content with content appended",
			args:           args{itemWithContent},
			wantTextchunks: []string{itemWithDescription.Title + "\n\n" + itemWithContent.Content},
		},
		{
			name:           "return sanitized content with html removed",
			args:           args{itemWithHTMLContent},
			wantTextchunks: []string{itemWithDescription.Title + "\n\n" + itemWithContent.Content},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTextchunks := GetSanitizedContentChunks(tt.args.item); !reflect.DeepEqual(gotTextchunks, tt.wantTextchunks) {
				t.Errorf("GetSanitizedContent() = %v, want %v", gotTextchunks, tt.wantTextchunks)
			}
		})
	}
}
