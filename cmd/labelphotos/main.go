package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jastribl/photosync/config"
	"github.com/jastribl/photosync/photos"
	"github.com/jastribl/photosync/utils"
)

func main() {
	// Setup logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Setup configs
	cfg := config.NewConfig()

	// Get a new Photos Client
	client, err := photos.NewClientForUser(cfg)
	utils.FatalError(err)

	args := os.Args[1:]
	rootPicturesDir := args[0]
	albumName := args[1]
	fmt.Println("Running for the following input")
	fmt.Println("Root picture dir: '" + rootPicturesDir + "'")
	fmt.Println("Album Name: '" + albumName + "'")

	album, err := client.GetAlbumWithTitleContains(albumName)
	utils.FatalError(err)

	if album == nil {
		log.Fatalln("Album not found with name '" + albumName + "'")
	}

	mediaItems, err := client.GetAllMediaItemsForAlbum(album)
	utils.FatalError(err)

	lowerCaseFileNameToIndexInAlbum := map[string]int{}
	for i, item := range mediaItems {
		lowerCaseFileNameToIndexInAlbum[strings.ToLower(item.Filename)] = i
	}

	folderToIngoreContainsLowercase := []string{
		"Photos from Others",
		"Pictures from Others",
		"Photos from Michael",
		"/Wendy",
	}

	ignoreFolderForLabeling := func(folderName string) bool {
		for _, ignoreMatch := range folderToIngoreContainsLowercase {
			if strings.Contains(strings.ToLower(folderName), strings.ToLower(ignoreMatch)) {
				return true
			}
		}
		return false
	}

	queue := []string{rootPicturesDir}
	listOfFolderInfo := []*FolderInfo{}
	for len(queue) > 0 {
		currentItem := queue[0]
		queue = queue[1:]

		files, err := ioutil.ReadDir(currentItem)
		utils.FatalError(err)

		lowercasePhotoFilenames := []string{}
		newfrontOfQueue := []string{}
		for _, file := range files {
			if file.IsDir() {
				newfrontOfQueue = append(newfrontOfQueue, currentItem+file.Name()+"/")
			} else {
				lowercasePhotoFilenames = append(lowercasePhotoFilenames, strings.ToLower(file.Name()))
			}
		}
		queue = append(newfrontOfQueue, queue...)
		folderInfo := &FolderInfo{
			path: currentItem,
		}
		highestIndexInAlbum := -1
		for _, lowerCaseFilename := range lowercasePhotoFilenames {
			indexInAlbum, foundFile := lowerCaseFileNameToIndexInAlbum[lowerCaseFilename]
			if foundFile && (highestIndexInAlbum == -1 || indexInAlbum > highestIndexInAlbum) {
				highestIndexInAlbum = indexInAlbum
			}
		}
		if highestIndexInAlbum != -1 {
			folderInfo.lastMediaItem = mediaItems[highestIndexInAlbum]
		}
		if ignoreFolderForLabeling(folderInfo.path) {
			log.Printf("Ignoring path: %s\n", folderInfo.path)
			continue
		}
		listOfFolderInfo = append(listOfFolderInfo, folderInfo)
	}

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
				response, err := client.AddTextEnrichmentToAlbum(album.ID, "", labelText)
				utils.FatalError(err)
				if response.Error == nil || response.Error.Status != "RESOURCE_EXHAUSTED" {
					break
				}
				// this means we need to retry after some time
				log.Println("Hit API Quota Limit, retrying after a short sleep...")
				time.Sleep(time.Duration(sleepSeconds) * time.Second)
			}
		} else {
			// firstPicOfNextFolder := ""
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
					response, err := client.AddTextEnrichmentToAlbum(album.ID, afterMediaID, labelText)
					utils.FatalError(err)
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

type FolderInfo struct {
	path          string
	lastMediaItem *photos.MediaItem
}
