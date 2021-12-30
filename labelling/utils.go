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
	mediaItems, err := client.GetAllMediaItemsForAlbum(album)
	if err != nil {
		log.Fatal(err)
	}

	lowercaseFilenameToIndexInAlbum := map[string]int{}
	for i, item := range mediaItems {
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
		for _, lowercaseFilename := range lowercaseFileNamesInDir {
			indexInAlbum, foundFile := lowercaseFilenameToIndexInAlbum[lowercaseFilename]
			if foundFile && (highestIndexInAlbum == -1 || indexInAlbum > highestIndexInAlbum) {
				highestIndexInAlbum = indexInAlbum
			}
		}

		folderInfo := &FolderInfo{
			Path: fullPathWithRoot,
		}
		if highestIndexInAlbum != -1 {
			folderInfo.LastMediaItem = mediaItems[highestIndexInAlbum]
		}
		listOfFolderInfo = append(listOfFolderInfo, folderInfo)
	}

	return listOfFolderInfo
}