package main

import (
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
	allFilenamesLowerCaseToMediaItems, err := client.GetAllLowercaseFilenameToMediaItemMapWithCache()
	if err != nil {
		log.Fatal(err)
	}

	args := os.Args[1:]
	rootPicturesDir := args[0]

	allLocalFilenames := files.GetAllFilenamesInDir(
		rootPicturesDir,
		[]*regexp.Regexp{},
		[]*regexp.Regexp{},
	)

	for _, filename := range allLocalFilenames {
		if mediaItems, found := allFilenamesLowerCaseToMediaItems[strings.ToLower(filename)]; found {
			if len(mediaItems) > 1 {
				log.Printf("Found multiple media items for '%s', leaving untouched\n", filename)
				continue
			}

			newFolderPath := rootPicturesDir + "/" + mediaItems[0].MediaMetadata.CreationTime[:10] + "/"
			if !files.FileExists(newFolderPath) {
				err := os.Mkdir(newFolderPath, 0777)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Created folder: %s\n", newFolderPath)
			}
			log.Printf("Moved '%s' into folder '%s'\n", filename, newFolderPath)
			err := os.Rename(rootPicturesDir+"/"+filename, newFolderPath+filename)
			if err != nil {
				log.Fatal(err)
			}
			continue
		}

		log.Printf("Unable to find Google drive photo '%s', leaving untouched\n", filename)
	}
}
