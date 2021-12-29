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
	"github.com/jastribl/photosync/photos"
	"github.com/jastribl/photosync/utils"
)

// folderToIngoreContainsLowercase := []string{
// 	"Photos from Others",
// 	"Pictures from Others",
// 	"2018-03-09 - Club",
// 	"Pictures from Dad",
// 	"Dads Pictures",
// }

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
	utils.FatalError(err)

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
	utils.FatalError(err)

	if album == nil {
		log.Fatalln("Album not found with name '" + albumName + "'")
	}

	listOfFolderInfo, err := getFolderInfo(rootPicturesDir, client, album)
	utils.FatalError(err)

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
					if !createLabels {
						break
					}
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

func getFolderInfo(
	rootPicturesDir string,
	client *photos.Client,
	album *photos.Album,
) ([]*FolderInfo, error) {
	mediaItems, err := client.GetAllMediaItemsForAlbum(album)
	if err != nil {
		return nil, err
	}

	lowerCaseFileNameToIndexInAlbum := map[string]int{}
	for i, item := range mediaItems {
		lowerCaseFileNameToIndexInAlbum[strings.ToLower(item.Filename)] = i
	}

	queue := []string{rootPicturesDir}
	listOfFolderInfo := []*FolderInfo{}
	for len(queue) > 0 {
		currentItem := queue[0]
		queue = queue[1:]

		files, err := ioutil.ReadDir(currentItem)
		if err != nil {
			return nil, err
		}

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

	return listOfFolderInfo, nil
}

type FolderInfo struct {
	path          string
	lastMediaItem *photos.MediaItem
}
