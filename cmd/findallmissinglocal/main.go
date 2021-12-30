package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/jastribl/photosync/config"
	"github.com/jastribl/photosync/files"
	"github.com/jastribl/photosync/photos"
)

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

	allLocalFileNamesMap := files.GetAllFileNamesInDirAsMap(
		cfg.RootPicturesDir,
		[]*regexp.Regexp{},
		[]*regexp.Regexp{},
	)
	allLocalFilesLowerCaseMap := map[string]int{}
	for fileName, count := range allLocalFileNamesMap {
		// log.Printf("adding to map: %s\n", strings.ToLower(fileName))
		allLocalFilesLowerCaseMap[strings.ToLower(fileName)] = count
	}

	allPhotosMediaItems, err := client.GetAllMediaItemsWithCache()
	if err != nil {
		log.Fatal(err)
	}
	lowerCaseFileNameToProductUrls := map[string][]string{}
	for _, mediaItem := range allPhotosMediaItems {
		lowerCaseFileName := strings.ToLower(mediaItem.Filename)
		if productUrlsFound, found := lowerCaseFileNameToProductUrls[lowerCaseFileName]; found {
			lowerCaseFileNameToProductUrls[lowerCaseFileName] = append(productUrlsFound, mediaItem.ProductULR)
		} else {
			lowerCaseFileNameToProductUrls[lowerCaseFileName] = []string{mediaItem.ProductULR}
		}
	}

	// todo: make these constants
	fileNameReplacements := []struct{ a, b string }{
		{".heic", ".jpg"},
		{".jpg", ".heic"},
	}

	// util function to check the map and decrement
	checkAndRemoveFilenameFromMap := func(key string) bool {
		numLeft, found := allLocalFilesLowerCaseMap[key]
		if found {
			// This means we found it, reduce the number from the map
			if numLeft == 1 {
				delete(allLocalFilesLowerCaseMap, key)
			} else {
				allLocalFilesLowerCaseMap[key] -= 1
			}
		}
		return found
	}

MEDIA_ITEM_LOOP:
	for _, mediaItem := range allPhotosMediaItems {
		lowerCaseFileName := strings.ToLower(mediaItem.Filename)
		if checkAndRemoveFilenameFromMap(lowerCaseFileName) {
			continue
		}

		for _, pair := range fileNameReplacements {
			replacementString := strings.ReplaceAll(lowerCaseFileName, pair.a, pair.b)
			if checkAndRemoveFilenameFromMap(replacementString) {
				continue MEDIA_ITEM_LOOP
			}
		}

		for i, productUrl := range lowerCaseFileNameToProductUrls[lowerCaseFileName] {
			fmt.Printf(
				"Missing locally one of: (%d) (%s) (%s): %s\n",
				i,
				mediaItem.MediaMetadata.CreationTime,
				lowerCaseFileName,
				productUrl,
			)
		}
	}
}
