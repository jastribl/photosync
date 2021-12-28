package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jastribl/photosync/config"
	"github.com/jastribl/photosync/files"
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
	fmt.Println("Running for the following input")
	fmt.Println("Root picture dir: '" + rootPicturesDir + "'")

	allLocalFileNames, err := files.GetAllFileNamesInDir(
		rootPicturesDir,
		[]string{},
		map[string]bool{},
	)
	utils.FatalError(err)

	allPhotosMediaItems, err := client.GetAllMediaItems(true)
	utils.FatalError(err)

	allPhotosLowerCaseFilenamesToMedia := map[string][]*photos.MediaItem{}
	for _, mediaItem := range allPhotosMediaItems {
		lowerCaseFileName := strings.ToLower(mediaItem.Filename)
		if list, ok := allPhotosLowerCaseFilenamesToMedia[lowerCaseFileName]; ok {
			allPhotosLowerCaseFilenamesToMedia[lowerCaseFileName] = append(list, mediaItem)
		} else {
			allPhotosLowerCaseFilenamesToMedia[lowerCaseFileName] = []*photos.MediaItem{mediaItem}
		}
	}

	// fileNameReplacements := []struct{ a, b string }{
	// 	{".heic", ".jpg"},
	// 	{".jpg", ".heic"},
	// }

	for _, localFilename := range allLocalFileNames {
		lowerCaseLocalFilename := strings.ToLower(localFilename)
		if items, ok := allPhotosLowerCaseFilenamesToMedia[lowerCaseLocalFilename]; ok {
			if len(items) > 1 {
				fmt.Printf("Found multiple media (%d) for filename %s\n", len(items), lowerCaseLocalFilename)
				for i, item := range items {
					log.Printf("Item %d: %s\n", i, item.ProductULR)
				}
			}
		} else {
			fmt.Printf("Photos missing: %s\n", lowerCaseLocalFilename)
		}
	}

	// for _, mediaItem := range allPhotosMediaItems {
	// 	lowerCaseFileName := strings.ToLower(mediaItem.Filename)
	// 	if _, ok := allLocalFilesLowerCaseMap[mediaItem.Filename]; ok {
	// 		continue
	// 	}
	// 	found := false
	// 	for _, pair := range fileNameReplacements {
	// 		if _, ok := allLocalFilesLowerCaseMap[strings.ReplaceAll(lowerCaseFileName, pair.a, pair.b)]; ok {
	// 			found = true
	// 			break
	// 		}
	// 	}
	// 	if found {
	// 		continue
	// 	}

	// 	fmt.Printf(
	// 		"Missing locally: (%s) (%s): %s\n",
	// 		mediaItem.MediaMetadata.CreationTime,
	// 		lowerCaseFileName,
	// 		mediaItem.ProductULR,
	// 	)
	// }
}