package labelling

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jastribl/photosync/photos"
)

func TestShouldIgnoreFolder(t *testing.T) {
	tests := []struct {
		name       string
		folderName string
		want       bool
	}{
		{name: "ordinary folder", folderName: "Trip 2021", want: false},
		{name: "pictures from someone", folderName: "Pictures from John", want: true},
		{name: "photos from someone", folderName: "photos from John", want: true},
		{name: "Wendy", folderName: "Wendy", want: true},
		// The allow-list entry is written as a bare name, so it only
		// applies when the name is passed without a path prefix.
		{name: "photos from Michael allowed", folderName: "Photos from Michael", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldIgnoreFolder(tt.folderName); got != tt.want {
				t.Errorf("ShouldIgnoreFolder(%q) = %v, want %v", tt.folderName, got, tt.want)
			}
		})
	}
}

func TestGetTopLevelFolderInfo(t *testing.T) {
	root := t.TempDir()
	for _, relPath := range []string{
		"Trip A/IMG_001.jpg",
		"Trip A/IMG_002.jpg",
		"Trip B/IMG_003.jpg",
	} {
		fullPath := filepath.Join(root, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	mediaItems := []*photos.MediaItem{
		{ID: "1", Filename: "IMG_001.jpg", MediaMetadata: photos.MediaMetadata{CreationTime: "2021-01-02T10:00:00Z"}},
		{ID: "2", Filename: "IMG_002.jpg", MediaMetadata: photos.MediaMetadata{CreationTime: "2021-01-03T10:00:00Z"}},
		{ID: "3", Filename: "IMG_003.jpg", MediaMetadata: photos.MediaMetadata{CreationTime: "2021-01-01T10:00:00Z"}},
		{ID: "9", Filename: "IMG_999.jpg", MediaMetadata: photos.MediaMetadata{CreationTime: "2020-12-31T10:00:00Z"}},
	}

	// The client argument is unused by GetTopLevelFolderInfo.
	folders := GetTopLevelFolderInfo(root+"/", nil, mediaItems)
	if len(folders) != 2 {
		t.Fatalf("GetTopLevelFolderInfo() returned %d folders, want 2", len(folders))
	}

	// Items are sorted newest-first, so the "last" media item of a
	// folder is its oldest item in the album order.
	if folders[0].NumMediaItems != 2 || folders[0].LastMediaItem == nil || folders[0].LastMediaItem.Filename != "IMG_001.jpg" {
		t.Errorf("Trip A folder = %+v, want 2 items with last IMG_001.jpg", folders[0])
	}
	if folders[1].NumMediaItems != 1 || folders[1].LastMediaItem == nil || folders[1].LastMediaItem.Filename != "IMG_003.jpg" {
		t.Errorf("Trip B folder = %+v, want 1 item with last IMG_003.jpg", folders[1])
	}
}

func TestGetTopLevelFolderInfoSkipsIgnoredAndEmptyFolders(t *testing.T) {
	root := t.TempDir()
	for _, relPath := range []string{
		"Trip A/IMG_001.jpg",
		"Pictures from John/IMG_002.jpg",
	} {
		fullPath := filepath.Join(root, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "Empty"), 0755); err != nil {
		t.Fatal(err)
	}

	mediaItems := []*photos.MediaItem{
		{ID: "1", Filename: "IMG_001.jpg", MediaMetadata: photos.MediaMetadata{CreationTime: "2021-01-02T10:00:00Z"}},
		{ID: "2", Filename: "IMG_002.jpg", MediaMetadata: photos.MediaMetadata{CreationTime: "2021-01-03T10:00:00Z"}},
	}

	// Folders come back in directory order: Empty, then Trip A.
	// "Pictures from John" is skipped as an ignored folder.
	folders := GetTopLevelFolderInfo(root+"/", nil, mediaItems)
	if len(folders) != 2 {
		t.Fatalf("GetTopLevelFolderInfo() returned %d folders, want 2 (Empty and Trip A)", len(folders))
	}
	if folders[0].NumMediaItems != 0 || folders[0].LastMediaItem != nil {
		t.Errorf("Empty folder = %+v, want 0 items and no last media item", folders[0])
	}
	if folders[1].NumMediaItems != 1 {
		t.Errorf("Trip A NumMediaItems = %d, want 1", folders[1].NumMediaItems)
	}
}
