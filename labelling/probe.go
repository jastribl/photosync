package labelling

import "github.com/jastribl/photosync/photos"

// FolderProvenEmpty reports whether every picked photo of a folder is known
// to be missing from the album — each one failed as an anchor with the
// not-in-album error (see photos.IsInvalidMediaItemID). A folder with no
// picked photos is not proven empty: it simply has no label. Only a
// conclusive proof skips a label; an untested photo keeps it.
func FolderProvenEmpty(items []*photos.MediaItem, knownMissing map[string]bool) bool {
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		if !knownMissing[item.ID] {
			return false
		}
	}
	return true
}
