package labelling

import (
	"log"
	"os"
	"sort"

	"github.com/jastribl/photosync/files"
	"github.com/jastribl/photosync/photos"
)

type FolderInfo struct {
	Path          string
	NumMediaItems int
	LastMediaItem *photos.MediaItem
	// MediaItems is the folder's picked photos, oldest first, so
	// MediaItems[0] is LastMediaItem. BuildLabelPlan uses this list as
	// the folder's anchor candidates, oldest first.
	MediaItems []*photos.MediaItem
}

func GetTopLevelFolderInfo(
	rootDir string,
	client *photos.Client,
	allMediaItems []*photos.MediaItem,
) []*FolderInfo {
	sort.Slice(allMediaItems, func(i, j int) bool {
		return allMediaItems[i].MediaMetadata.CreationTime > allMediaItems[j].MediaMetadata.CreationTime
	})

	lowercaseFilenameToAlbumMediaMap := photos.MediaItemsToLowercaseFilenameMap(allMediaItems)

	// Index of each picked item in the CreationTime-descending list, so
	// a folder's items can be ordered oldest first (highest index first).
	indexByID := map[string]int{}
	for i, item := range allMediaItems {
		indexByID[item.ID] = i
	}

	// Find all top level files and assert they are all topLevelDirs
	topLevelDirs, err := os.ReadDir(rootDir)
	if err != nil {
		log.Fatal(err)
	}
	listOfFolderInfo := []*FolderInfo{}
	for _, topLevelDir := range topLevelDirs {
		if topLevelDir.Name() == ".DS_Store" {
			continue
		}
		if !topLevelDir.IsDir() {
			log.Fatalf("Found non-dir in top level of given root dir - must be  dir: %s", topLevelDir.Name())
		}

		fullPathWithRoot := rootDir + topLevelDir.Name() + "/"

		lowercaseFileNamesInDir := files.GetAllLowercaseFilenamesInDir(fullPathWithRoot)

		// Collect the folder's picked photos, oldest first. The picked
		// list is sorted newest first, so oldest means the highest
		// index in it. A picked photo can match more than one local
		// file with the same name, so dedupe by media item ID.
		seen := map[string]bool{}
		var folderItems []*photos.MediaItem
		for _, lowercaseFilename := range lowercaseFileNamesInDir {
			for _, item := range lowercaseFilenameToAlbumMediaMap[lowercaseFilename] {
				if seen[item.ID] {
					continue
				}
				seen[item.ID] = true
				folderItems = append(folderItems, item)
			}
		}
		sort.Slice(folderItems, func(a, b int) bool {
			return indexByID[folderItems[a].ID] > indexByID[folderItems[b].ID]
		})

		folderInfo := &FolderInfo{
			Path:          fullPathWithRoot,
			NumMediaItems: len(folderItems),
			MediaItems:    folderItems,
		}
		if len(folderItems) > 0 {
			folderInfo.LastMediaItem = folderItems[0]
		}
		listOfFolderInfo = append(listOfFolderInfo, folderInfo)
	}

	return listOfFolderInfo
}
