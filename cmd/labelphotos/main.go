package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/jastribl/photosync/config"
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

func ignoreFolderForLabeling(folderName string) bool {
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

func main() {
	// Setup logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Setup configs
	cfg := config.NewConfig()

	// Get a new Photos Client
	client, err := photos.NewClientForUser(cfg)
	if err != nil {
		log.Fatal(err)
	}

	args := os.Args[1:]
	rootPicturesDir := args[0]
	albumName := args[1]
	createLabels := false
	if len(args) > 2 && args[2] == "--create" {
		createLabels = true
	}
	fmt.Println("Running for the following input")
	fmt.Println("Root picture dir: '" + rootPicturesDir + "'")
	fmt.Println("Album Name: '" + albumName + "'")

	album, err := client.GetAlbumWithTitleContains(albumName)
	if err != nil {
		log.Fatal(err)
	}

	if album == nil {
		log.Fatalln("Album not found with name '" + albumName + "'")
	}

	listOfFolderInfo := getTopLevelFolderInfoForLabelling(rootPicturesDir, client, album)

	for i, folderInfo := range listOfFolderInfo {
		if folderInfo.path == rootPicturesDir {
			log.Println("Not adding label for root")
			continue
		}
		if ignoreFolderForLabeling(folderInfo.path) {
			log.Printf("Not adding label for path: %s\n", folderInfo.path)
			continue
		}
		labelText := folderInfo.path[len(rootPicturesDir) : len(folderInfo.path)-1]
		if i == len(listOfFolderInfo)-1 {
			fmt.Printf("Adding '%s' at the beginning of the album\n", labelText)
			if !createLabels {
				break
			}
			_, err := client.AddTextEnrichmentToAlbum(album.ID, nil, labelText)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			var lastMediaItemOfNextFolder *photos.MediaItem = nil
			nextFolderPath := ""
			for j := i + 1; lastMediaItemOfNextFolder == nil && j < len(listOfFolderInfo); j += 1 {
				nextFolderInfo := listOfFolderInfo[j]
				nextFolderPath = nextFolderInfo.path
				lastMediaItemOfNextFolder = nextFolderInfo.lastMediaItem
			}
			if lastMediaItemOfNextFolder == nil {
				log.Fatalln("Got no first pic of next folder for folder")
			}

			afterMediaID := lastMediaItemOfNextFolder.ID
			if afterMediaID == "" {
				log.Fatalln("Got empty media id for: " + lastMediaItemOfNextFolder.Filename)
			}
			fmt.Printf(
				"Adding '%s' after '%s' (last pic of folder '%s') \n",
				labelText,
				lastMediaItemOfNextFolder.Filename,
				nextFolderPath,
			)
			if !createLabels {
				break
			}
			_, err := client.AddTextEnrichmentToAlbum(album.ID, lastMediaItemOfNextFolder, labelText)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func getTopLevelFolderInfoForLabelling(
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
		if ignoreFolderForLabeling(fullPathWithRoot) {
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
			path: fullPathWithRoot,
		}
		if highestIndexInAlbum != -1 {
			folderInfo.lastMediaItem = mediaItems[highestIndexInAlbum]
		}
		listOfFolderInfo = append(listOfFolderInfo, folderInfo)
	}

	return listOfFolderInfo
}

type FolderInfo struct {
	path          string
	lastMediaItem *photos.MediaItem
}
