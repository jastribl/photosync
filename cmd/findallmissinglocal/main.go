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
	if err != nil {
		log.Fatal(err)
	}

	allPhotosMediaItems, err := client.GetAllMediaItems()
	if err != nil {
		log.Fatal(err)
	}

	for _, mediaItem := range allPhotosMediaItems {
		if _, ok := allLocalFileNamesMap[mediaItem.Filename]; ok {
			continue
		}
		if _, ok := allLocalFileNamesMap[strings.ReplaceAll(mediaItem.Filename, ".HEIC", ".jpg")]; ok {
			continue
		}
		if _, ok := allLocalFileNamesMap[strings.ReplaceAll(mediaItem.Filename, ".HEIC", ".jpg")]; ok {
			continue
		}
		if _, ok := allLocalFileNamesMap[strings.ReplaceAll(mediaItem.Filename, ".HEIC", ".heic")]; ok {
			continue
		}
		if _, ok := allLocalFileNamesMap[strings.ReplaceAll(mediaItem.Filename, ".heic", ".jpg")]; ok {
			continue
		}
		fmt.Printf("Missing locally: (%s): %s\n", mediaItem.Filename, mediaItem.ProductULR)
	}
}
