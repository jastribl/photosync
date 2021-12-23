package main

import (
	"log"
	"os"
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
	allFileNames, err := files.GetAllFileNamesInDirAsMap(
		rootPicturesDir,
		cfg.PicturePathSubstringsToIgnore,
		cfg.FileNamesToIgnoreMap,
	)
	if err != nil {
		log.Fatal(err)
	}

	mediaItmes, err := client.GetAllMediaItems(true)
	if err != nil {
		log.Fatal(err)
	}

	freeBefore, _ := time.Parse("2006-01-02", cfg.FreeBeforeDate)

	for _, mediaItem := range mediaItmes {
		filename := mediaItem.Filename
		_, found := allFileNames[filename]
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
