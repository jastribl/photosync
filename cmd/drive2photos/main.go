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
	albumName := args[1]
	fmt.Println("Running for the following input")
	fmt.Println("Root picture dir: '" + rootPicturesDir + "'")
	fmt.Println("Album Name: '" + albumName + "'")

	log.Println("Getting all drive filenames")
	allDriveFileNames, err := files.GetAllFileNamesInDirAsMap(
		rootPicturesDir,
		[]string{ // folderDenyRegexs
			".*[pP]ictures [fF]rom .*$",
			".*[pP]hotos [fF]rom .*$",
			"^Wendy$",
		},
		[]string{ // folderAllowRegexs
			"^Photos from Michael$",
		},
		map[string]bool{},
	)
	allDriveLowercaseFilenamesMap := map[string]int{}
	for fileName, numberOfThatFile := range allDriveFileNames {
		allDriveLowercaseFilenamesMap[strings.ToLower(fileName)] = numberOfThatFile
	}
	utils.FatalError(err)

	log.Println("Getting album")
	album, err := client.GetAlbumWithTitleContains(albumName)
	utils.FatalError(err)
	if album == nil {
		log.Fatal("Album not found with name '" + albumName + "'")
	}

	log.Println("Getting album media items")
	albumMediaItems, err := client.GetAllMediaItemsForAlbum(album)
	utils.FatalError(err)
	allAlbumFileNamesLowerCaseToMediaItems := mediaItemsToLowercaseFilenameMap(albumMediaItems)

	log.Println("Getting all media items")
	allMediaItems, err := client.GetAllMediaItems(true)
	utils.FatalError(err)
	allMediaItemLowerCaseFilenamesToMediaItems := mediaItemsToLowercaseFilenameMap(allMediaItems)

	numExtra := 0
	for fileNameLowerCase, mediaItems := range allAlbumFileNamesLowerCaseToMediaItems {
		fileNameLowerCase = strings.ToLower(fileNameLowerCase)

		// Check if filename is in drive folder already
		if _, ok := allDriveLowercaseFilenamesMap[fileNameLowerCase]; ok {
			continue
		}

		// Check the same but for replaced filenames
		fileNameReplacements := []struct{ a, b string }{
			{".heic", ".jpg"},
			{".jpg", ".heic"},
		}
		found := false
		for _, pair := range fileNameReplacements {
			if _, ok := allDriveLowercaseFilenamesMap[strings.ReplaceAll(fileNameLowerCase, pair.a, pair.b)]; ok {
				found = true
				break
			}
		}

		if !found {
			for i, mediaItem := range mediaItems {
				fmt.Printf(
					"Photos extra file (date: %s) (%d): %s - %s\n",
					mediaItem.MediaMetadata.CreationTime,
					i,
					fileNameLowerCase,
					mediaItem.ProductULR,
				)
			}
			numExtra += 1
		}
	}

	numMissing := 0
	for fileNameLowerCase := range allDriveLowercaseFilenamesMap {
		fileNameLowerCaseHEIC := strings.ReplaceAll(fileNameLowerCase, ".jpg", ".heic")
		// Check if the Google Photos album contains the file
		if _, ok := allAlbumFileNamesLowerCaseToMediaItems[fileNameLowerCase]; ok {
			continue
		}

		// Also check with swapping extension
		if _, ok := allAlbumFileNamesLowerCaseToMediaItems[fileNameLowerCaseHEIC]; ok {
			continue
		}

		if mediaItems, ok := allMediaItemLowerCaseFilenamesToMediaItems[fileNameLowerCase]; ok {
			// Check if we have a media item for this file name - if so print that out so we can add it
			for i, mediaItem := range mediaItems {
				fmt.Printf("Link to missing in Photos: (%s) (%d): %s\n", fileNameLowerCase, i, mediaItem.ProductULR)
			}
		} else if mediaItems, ok := allMediaItemLowerCaseFilenamesToMediaItems[fileNameLowerCaseHEIC]; ok {
			// Check for the same thing but with extensions swapped
			for i, mediaItem := range mediaItems {
				fmt.Printf("Link to missing in Photos: (%s) (%d): %s\n", fileNameLowerCaseHEIC, i, mediaItem.ProductULR)
			}
		} else {
			// Otherwise we're missing the file and don't know where to find it
			fmt.Printf("Google Photos missing file: (%s)\n", fileNameLowerCase)
		}
		numMissing += 1
	}

	log.Printf("Num Extra: %d\n", numExtra)
	log.Printf("Num Missing: %d\n", numMissing)
}

func mediaItemsToLowercaseFilenameMap(mediaItems []*photos.MediaItem) map[string][]*photos.MediaItem {
	lowerCaseFilenamesToMediaItems := map[string][]*photos.MediaItem{}
	for _, item := range mediaItems {
		lowerFilename := strings.ToLower(item.Filename)
		if list, ok := lowerCaseFilenamesToMediaItems[lowerFilename]; ok {
			lowerCaseFilenamesToMediaItems[lowerFilename] = append(list, item)
		} else {
			lowerCaseFilenamesToMediaItems[lowerFilename] = []*photos.MediaItem{item}
		}
	}

	return lowerCaseFilenamesToMediaItems
}
