package main

import (
	"fmt"
	"log"
	"os"
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

	args := os.Args[1:]
	rootPicturesDir := args[0]
	fmt.Println("Running for the following input")
	fmt.Println("Root picture dir: '" + rootPicturesDir + "'")

	allLocalFileNamesMap, err := files.GetAllFileNamesInDirAsMap(
		rootPicturesDir,
		[]string{},
		map[string]bool{},
	)
	allLocalFilesLowerCaseMap := map[string]int{}
	for fileName, count := range allLocalFileNamesMap {
		allLocalFilesLowerCaseMap[strings.ToLower(fileName)] = count
	}
	if err != nil {
		log.Fatal(err)
	}

	allPhotosMediaItems, err := client.GetAllMediaItems(true)
	if err != nil {
		log.Fatal(err)
	}

	fileNameReplacements := []struct{ a, b string }{
		{".heic", ".jpg"},
		{".jpg", ".heic"},
	}

	for _, mediaItem := range allPhotosMediaItems {
		lowerCaseFileName := strings.ToLower(mediaItem.Filename)
		if _, ok := allLocalFilesLowerCaseMap[mediaItem.Filename]; ok {
			continue
		}
		found := false
		for _, pair := range fileNameReplacements {
			if _, ok := allLocalFilesLowerCaseMap[strings.ReplaceAll(lowerCaseFileName, pair.a, pair.b)]; ok {
				found = true
				break
			}
		}
		if found {
			continue
		}

		fmt.Printf(
			"Missing locally: (%s) (%s): %s\n",
			mediaItem.MediaMetadata.CreationTime,
			lowerCaseFileName,
			mediaItem.ProductULR,
		)
	}
}
