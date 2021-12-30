package labelling

import (
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/jastribl/photosync/files"
	"github.com/jastribl/photosync/photos"
)

// 2021
var (
	FOLDER_DENY_REGEXS = [...]*regexp.Regexp{
		regexp.MustCompile(".*[pP]ictures [fF]rom .*$"),
		regexp.MustCompile(".*[pP]hotos [fF]rom .*$"),
		regexp.MustCompile("^Wendy$"),
	}

	FOLDER_ALLOW_REGEXS = [...]*regexp.Regexp{
		regexp.MustCompile("^Photos from Michael$"),
	}
)

type FolderInfo struct {
	Path          string
	NumMediaItems int
	LastMediaItem *photos.MediaItem
}

func ShouldIgnoreFolder(folderName string) bool {
	for _, denyRegex := range FOLDER_DENY_REGEXS {
		if denyRegex.MatchString(folderName) {
			denyRuling := true
			for _, allowRegex := range FOLDER_ALLOW_REGEXS {
				if allowRegex.MatchString(folderName) {
					denyRuling = false
					break
				}
			}
			if denyRuling {
				return true
			}
		}
	}
	return false
}

func GetTopLevelFolderInfo(
	rootDir string,
	client *photos.Client,
	album *photos.Album,
) []*FolderInfo {
	albumMediaItems, err := client.GetAllMediaItemsForAlbum(album)
	if err != nil {
		log.Fatal(err)
	}

	lowercaseFilenameToAlbumMediaMap := photos.MediaItemsToLowercaseFilenameMap(albumMediaItems)

	lowercaseFilenameToIndexInAlbum := map[string]int{}
	for i, item := range albumMediaItems {
		lowercaseFilenameToIndexInAlbum[strings.ToLower(item.Filename)] = i
	}

	// Find all top level files and assert they are all topLevelDirs
	topLevelDirs, err := ioutil.ReadDir(rootDir)
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
		if ShouldIgnoreFolder(fullPathWithRoot) {
			log.Printf("Ignoring path 2: %s\n", fullPathWithRoot)
			continue
		}

		lowercaseFileNamesInDir := files.GetAllLowercaseFilenamesInDir(
			fullPathWithRoot,
			FOLDER_DENY_REGEXS[:],
			FOLDER_ALLOW_REGEXS[:],
		)

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
			folderInfo.LastMediaItem = albumMediaItems[highestIndexInAlbum]
		}
		listOfFolderInfo = append(listOfFolderInfo, folderInfo)
	}

	return listOfFolderInfo
}
