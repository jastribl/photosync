package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jastribl/photosync/config"
	"github.com/jastribl/photosync/files"
	"github.com/jastribl/photosync/photos"
)

// 2021
var (
	folderDenyRegexs = [...]*regexp.Regexp{
		regexp.MustCompile(".*[pP]ictures [fF]rom .*$"),
		regexp.MustCompile(".*[pP]hotos [fF]rom .*$"),
		regexp.MustCompile("^Wendy$"),
	}

	folderAllowRegexs = [...]*regexp.Regexp{
		regexp.MustCompile("^Photos from Michael$"),
	}
)

func ignoreFolderForLabeling(folderName string) bool {
	for _, denyRegex := range folderDenyRegexs {
		if denyRegex.MatchString(folderName) {
			denyRuling := true
			for _, allowRegex := range folderAllowRegexs {
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
			for sleepSeconds := 1; ; sleepSeconds *= 2 {
				if sleepSeconds > 10 {
					sleepSeconds = 10
				}
				fmt.Printf("Adding '%s' at the beginning of the album\n", labelText)
				if !createLabels {
					break
				}
				response, err := client.AddTextEnrichmentToAlbum(album.ID, "", labelText)
				if err != nil {
					log.Fatal(err)
				}
				if response.Error == nil || response.Error.Status != "RESOURCE_EXHAUSTED" {
					break
				}
				// this means we need to retry after some time
				log.Println("Hit API Quota Limit, retrying after a short sleep...")
				time.Sleep(time.Duration(sleepSeconds) * time.Second)
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
			} else {
				afterMediaID := lastMediaItemOfNextFolder.ID
				if afterMediaID == "" {
					log.Fatalln("Got empty media id for: " + lastMediaItemOfNextFolder.Filename)
				}
				// retry loop
				// TODO: See if this can be made generic on all requests
				for sleepSeconds := 1; ; sleepSeconds *= 2 {
					if sleepSeconds > 10 {
						sleepSeconds = 10
					}
					fmt.Printf(
						"Adding '%s' after '%s' (last pic of folder '%s') \n",
						labelText,
						lastMediaItemOfNextFolder.Filename,
						nextFolderPath[len(rootPicturesDir):len(nextFolderPath)-1],
					)
					if !createLabels {
						break
					}
					response, err := client.AddTextEnrichmentToAlbum(album.ID, afterMediaID, labelText)
					if err != nil {
						log.Fatal(err)
					}
					if response.Error == nil {
						break
					}
					if response.Error.Status != "RESOURCE_EXHAUSTED" {
						log.Printf("got error that's we're okay with?: %s - %s\n", response.Error.Message, response.Error.Status)
						break
					}
					// this means we need to retry after some time
					log.Println("Hit API Quota Limit, retrying after a short sleep...")
					time.Sleep(time.Duration(sleepSeconds) * time.Second)
				}
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

	lowerCaseFileNameToIndexInAlbum := map[string]int{}
	for i, item := range mediaItems {
		lowerCaseFileNameToIndexInAlbum[strings.ToLower(item.Filename)] = i
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

		filesNamesInDir := files.GetAllFileNamesInDir(
			fullPathWithRoot,
			folderDenyRegexs[:],
			folderAllowRegexs[:],
		)

		highestIndexInAlbum := -1
		for _, filename := range filesNamesInDir {
			lowerCaseFilename := strings.ToLower(filename)
			indexInAlbum, foundFile := lowerCaseFileNameToIndexInAlbum[lowerCaseFilename]
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
