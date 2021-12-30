package main

import (
	"fmt"
	"log"
	"os"
	"regexp"

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

	allLowercaseLocalFilenames := files.GetAllLowercaseFilenamesInDir(
		rootPicturesDir,
		[]*regexp.Regexp{
			regexp.MustCompile(".*[pP]ictures [fF]rom .*$"),
			regexp.MustCompile(".*[pP]hotos [fF]rom .*$"),
		},
		[]*regexp.Regexp{},
	)

	allPhotosLowerCaseFilenamesToMedia, err := client.GetAllLowercaseFilenameToMediaItemMapWithCache()
	if err != nil {
		log.Fatal(err)
	}

	for _, lowercaseLocalFilename := range allLowercaseLocalFilenames {
		if items, ok := allPhotosLowerCaseFilenamesToMedia[lowercaseLocalFilename]; ok {
			if len(items) > 1 {
				fmt.Printf("Found multiple media (%d) for filename %s\n", len(items), lowercaseLocalFilename)
				for i, item := range items {
					log.Printf("Item %d: %s\n", i, item.ProductULR)
				}
			}
		} else {
			fmt.Printf("Photos missing: %s\n", lowercaseLocalFilename)
		}
	}
}
