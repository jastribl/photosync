package labelling

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jastribl/photosync/photos"
)

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

func TestGetTopLevelFolderInfoIncludesAllFolders(t *testing.T) {
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

	// Folders come back in directory order: Empty, Pictures from John,
	// then Trip A. No folders are ignored anymore (the old deny/allow
	// rules are gone), so "Pictures from John" is included like any
	// other folder.
	folders := GetTopLevelFolderInfo(root+"/", nil, mediaItems)
	if len(folders) != 3 {
		t.Fatalf("GetTopLevelFolderInfo() returned %d folders, want 3", len(folders))
	}
	if folders[0].NumMediaItems != 0 || folders[0].LastMediaItem != nil {
		t.Errorf("Empty folder = %+v, want 0 items and no last media item", folders[0])
	}
	if folders[1].NumMediaItems != 1 || folders[1].LastMediaItem == nil || folders[1].LastMediaItem.Filename != "IMG_002.jpg" {
		t.Errorf("Pictures from John folder = %+v, want 1 item with last IMG_002.jpg", folders[1])
	}
	if folders[2].NumMediaItems != 1 {
		t.Errorf("Trip A NumMediaItems = %d, want 1", folders[2].NumMediaItems)
	}
}
