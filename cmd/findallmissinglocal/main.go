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

	allLocalFilenamesMap := files.GetAllFilenamesInDirAsMap(
		cfg.RootPicturesDir,
		[]*regexp.Regexp{},
		[]*regexp.Regexp{},
	)
	allLocalFilesLowerCaseMap := map[string]int{}
	for filename, count := range allLocalFilenamesMap {
		allLocalFilesLowerCaseMap[strings.ToLower(filename)] = count
	}

	allPhotosMediaItems, err := client.GetAllMediaItemsWithCache()
	if err != nil {
		log.Fatal(err)
	}
	lowercaseFilenameToProductUrls := map[string][]string{}
	for _, mediaItem := range allPhotosMediaItems {
		lowercaseFilename := strings.ToLower(mediaItem.Filename)
		if productUrlsFound, found := lowercaseFilenameToProductUrls[lowercaseFilename]; found {
			lowercaseFilenameToProductUrls[lowercaseFilename] = append(productUrlsFound, mediaItem.ProductULR)
		} else {
			lowercaseFilenameToProductUrls[lowercaseFilename] = []string{mediaItem.ProductULR}
		}
	}

	// todo: make these constants
	filenameReplacements := []struct{ a, b string }{
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
		lowercaseFilename := strings.ToLower(mediaItem.Filename)
		if checkAndRemoveFilenameFromMap(lowercaseFilename) {
			continue
		}

		for _, pair := range filenameReplacements {
			replacementString := strings.ReplaceAll(lowercaseFilename, pair.a, pair.b)
			if checkAndRemoveFilenameFromMap(replacementString) {
				continue MEDIA_ITEM_LOOP
			}
		}

		for i, productUrl := range lowercaseFilenameToProductUrls[lowercaseFilename] {
			fmt.Printf(
				"Missing locally one of: (%d) (%s) (%s): %s\n",
				i,
				mediaItem.MediaMetadata.CreationTime,
				lowercaseFilename,
				productUrl,
			)
		}
	}
}
