package labelling

import (
	"log"
	"os"
	"sort"
	"strings"

	"github.com/jastribl/photosync/files"
	"github.com/jastribl/photosync/photos"
)

type FolderInfo struct {
	Path          string
	NumMediaItems int
	LastMediaItem *photos.MediaItem
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

	lowercaseFilenameToIndexInAlbum := map[string]int{}
	for i, item := range allMediaItems {
		lowercaseFilenameToIndexInAlbum[strings.ToLower(item.Filename)] = i
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

		highestIndexInAlbum := -1
		numMediaItemsInDir := 0
		for _, lowercaseFilename := range lowercaseFileNamesInDir {
			// find highest index item in folder
			indexInAlbum, foundFile := lowercaseFilenameToIndexInAlbum[lowercaseFilename]
			if foundFile && (highestIndexInAlbum == -1 || indexInAlbum > highestIndexInAlbum) {
				highestIndexInAlbum = indexInAlbum
			}

			// also find the number of items in each folder
			numMediaItemsInDir += len(lowercaseFilenameToAlbumMediaMap[lowercaseFilename])
		}

		folderInfo := &FolderInfo{
			Path:          fullPathWithRoot,
			NumMediaItems: numMediaItemsInDir,
		}
		if highestIndexInAlbum != -1 {
			folderInfo.LastMediaItem = allMediaItems[highestIndexInAlbum]
		}
		listOfFolderInfo = append(listOfFolderInfo, folderInfo)
	}

	return listOfFolderInfo
}
