package labelling

import (
	"testing"

	"github.com/jastribl/photosync/photos"
)

func TestFolderProvenEmpty(t *testing.T) {
	items := []*photos.MediaItem{
		{ID: "a", Filename: "a.jpg"},
		{ID: "b", Filename: "b.jpg"},
	}
	allMissing := map[string]bool{"a": true, "b": true}
	if !FolderProvenEmpty(items, allMissing) {
		t.Error("folder with all photos known missing should be proven empty")
	}
	onePresent := map[string]bool{"a": true}
	if FolderProvenEmpty(items, onePresent) {
		t.Error("folder with an untested photo should not be proven empty")
	}
	if FolderProvenEmpty(items, map[string]bool{}) {
		t.Error("folder with nothing known missing should not be proven empty")
	}
	if FolderProvenEmpty(items, nil) {
		t.Error("folder with a nil missing set should not be proven empty")
	}
	if FolderProvenEmpty(nil, allMissing) {
		t.Error("folder with no picked photos should not be proven empty")
	}
	if FolderProvenEmpty([]*photos.MediaItem{}, allMissing) {
		t.Error("folder with an empty item list should not be proven empty")
	}
}
