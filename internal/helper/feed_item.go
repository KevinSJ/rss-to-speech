package helper

import "github.com/mmcdole/gofeed"

func GetSanitizedContentChunks(item *gofeed.Item) (textchunks []string) {
	content := item.Title + "\n\n"

	if len(item.Content) > 0 {
		content += stripHtmlTags(item.Content)
	} else if len(item.Description) > 0 {
		content += stripHtmlTags(item.Description)
	}

	return chunksByte(content, 5000)
}
