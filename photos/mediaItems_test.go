package photos

import (
	"reflect"
	"testing"
)

func TestMediaItemsToLowercaseFilenameMap(t *testing.T) {
	tests := []struct {
		name       string
		mediaItems []*MediaItem
		want       map[string][]*MediaItem
	}{
		{
			name:       "nil input gives empty map",
			mediaItems: nil,
			want:       map[string][]*MediaItem{},
		},
		{
			name:       "empty input gives empty map",
			mediaItems: []*MediaItem{},
			want:       map[string][]*MediaItem{},
		},
		{
			name: "single item keyed by lowercase filename",
			mediaItems: []*MediaItem{
				{ID: "1", Filename: "IMG_001.JPG"},
			},
			want: map[string][]*MediaItem{
				"img_001.jpg": {{ID: "1", Filename: "IMG_001.JPG"}},
			},
		},
		{
			name: "filenames differing only by case group together in order",
			mediaItems: []*MediaItem{
				{ID: "1", Filename: "IMG_001.JPG"},
				{ID: "2", Filename: "img_001.jpg"},
				{ID: "3", Filename: "Img_001.Jpg"},
			},
			want: map[string][]*MediaItem{
				"img_001.jpg": {
					{ID: "1", Filename: "IMG_001.JPG"},
					{ID: "2", Filename: "img_001.jpg"},
					{ID: "3", Filename: "Img_001.Jpg"},
				},
			},
		},
		{
			name: "distinct filenames stay separate",
			mediaItems: []*MediaItem{
				{ID: "1", Filename: "a.jpg"},
				{ID: "2", Filename: "B.jpg"},
				{ID: "3", Filename: "a.jpg"},
			},
			want: map[string][]*MediaItem{
				"a.jpg": {
					{ID: "1", Filename: "a.jpg"},
					{ID: "3", Filename: "a.jpg"},
				},
				"b.jpg": {{ID: "2", Filename: "B.jpg"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MediaItemsToLowercaseFilenameMap(tt.mediaItems); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MediaItemsToLowercaseFilenameMap() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
