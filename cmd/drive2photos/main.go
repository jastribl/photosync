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

	args := os.Args[1:]
	rootPicturesDir := args[0]
	albumName := args[1]
	fmt.Println("Running for the following input")
	fmt.Println("Root picture dir: '" + rootPicturesDir + "'")
	fmt.Println("Album Name: '" + albumName + "'")

	log.Println("Getting all drive filenames")
	allDriveLowercaseFilenamesMap := files.GetAllLowercaseFilenamesInDirAsMap(
		rootPicturesDir,
		// 2021
		[]*regexp.Regexp{ // FOLDER_DENY_REGEXS
			regexp.MustCompile(".*[pP]ictures [fF]rom .*$"),
			regexp.MustCompile(".*[pP]hotos [fF]rom .*$"),
			regexp.MustCompile("^Wendy$"),
		},
		[]*regexp.Regexp{ // FOLDER_ALLOW_REGEXS
			regexp.MustCompile("^Photos from Michael$"),
		},
	)

	log.Println("Getting album")
	album, err := client.GetAlbumWithTitle(albumName)
	if err != nil {
		log.Fatal(err)
	}
	if album == nil {
		log.Fatal("Album not found with name '" + albumName + "'")
	}

	log.Println("Getting album media items")
	albumMediaItems, err := client.GetAllMediaItemsForAlbum(album)
	if err != nil {
		log.Fatal(err)
	}
	allAlbumFilenamesLowerCaseToMediaItems := photos.MediaItemsToLowercaseFilenameMap(albumMediaItems)

	log.Println("Getting all media items")
	allPhotosLowerCaseFilenamesToMediaItems, err := client.GetAllLowercaseFilenameToMediaItemMapWithCache()
	if err != nil {
		log.Fatal(err)
	}

	numExtra := 0
MEDIA_ITEM_LOOP:
	for filenameLowerCase, mediaItems := range allAlbumFilenamesLowerCaseToMediaItems {
		filenameLowerCase = strings.ToLower(filenameLowerCase)

		// Check if filename is in drive folder already
		if _, ok := allDriveLowercaseFilenamesMap[filenameLowerCase]; ok {
			continue
		}

		// Check the same but for replaced filenames
		for _, pair := range files.FILE_NAME_REPLACEMENTS {
			if _, ok := allDriveLowercaseFilenamesMap[strings.ReplaceAll(filenameLowerCase, pair.A, pair.B)]; ok {
				continue MEDIA_ITEM_LOOP
			}
		}

		// We never found the media item
		for i, mediaItem := range mediaItems {
			fmt.Printf(
				"Photos extra file (date: %s) (%d): %s - %s\n",
				mediaItem.MediaMetadata.CreationTime,
				i,
				filenameLowerCase,
				mediaItem.ProductULR,
			)
		}
		numExtra += 1
	}

	numMissing := 0
	for filenameLowerCase := range allDriveLowercaseFilenamesMap {
		filenameLowerCaseHEIC := strings.ReplaceAll(filenameLowerCase, ".jpg", ".heic")
		// Check if the Google Photos album contains the file
		if _, ok := allAlbumFilenamesLowerCaseToMediaItems[filenameLowerCase]; ok {
			continue
		}

		// Also check with swapping extension
		if _, ok := allAlbumFilenamesLowerCaseToMediaItems[filenameLowerCaseHEIC]; ok {
			continue
		}

		if mediaItems, ok := allPhotosLowerCaseFilenamesToMediaItems[filenameLowerCase]; ok {
			// Check if we have a media item for this file name - if so print that out so we can add it
			for i, mediaItem := range mediaItems {
				fmt.Printf("Link to missing in Photos: (%s) (%d): %s\n", filenameLowerCase, i, mediaItem.ProductULR)
			}
		} else if mediaItems, ok := allPhotosLowerCaseFilenamesToMediaItems[filenameLowerCaseHEIC]; ok {
			// Check for the same thing but with extensions swapped
			for i, mediaItem := range mediaItems {
				fmt.Printf("Link to missing in Photos: (%s) (%d): %s\n", filenameLowerCaseHEIC, i, mediaItem.ProductULR)
			}
		} else {
			// Otherwise we're missing the file and don't know where to find it
			fmt.Printf("Google Photos missing file: (%s)\n", filenameLowerCase)
		}
		numMissing += 1
	}

	log.Printf("Num Extra: %d\n", numExtra)
	log.Printf("Num Missing: %d\n", numMissing)
}
