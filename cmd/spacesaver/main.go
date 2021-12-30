package main

import (
	"log"
	"os"
	"regexp"
	"strings"
	"time"

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

	rootPicturesDir := cfg.RootPicturesDir
	allFilenames := files.GetAllFilenamesInDirAsMap(
		rootPicturesDir,
		cfg.PicturePathRegexsToIgnore,
		[]*regexp.Regexp{},
	)
	allFilenamesLowerCaseMap := map[string]int{}
	for filename, count := range allFilenames {
		allFilenamesLowerCaseMap[strings.ToLower(filename)] = count
	}

	mediaItmes, err := client.GetAllMediaItemsWithCache()
	if err != nil {
		log.Fatal(err)
	}

	freeBefore, _ := time.Parse("2006-01-02", cfg.FreeBeforeDate)

	for _, mediaItem := range mediaItmes {
		filename := mediaItem.Filename
		_, found := allFilenamesLowerCaseMap[strings.ToLower(filename)]
		if !found {
			timeOfImage, err := time.Parse(
				"2006-01-02T15:04:05Z",
				mediaItem.MediaMetadata.CreationTime,
			)
			if err != nil {
				log.Fatal(err)
			}
			if timeOfImage.After(freeBefore) {
				log.Println(mediaItem.ProductULR)
			}
		}
	}
}
