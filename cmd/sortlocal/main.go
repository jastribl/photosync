package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/bradfitz/latlong"
	utils "github.com/jastribl/photosync/cmd"
	"github.com/jastribl/photosync/config"
	"github.com/jastribl/photosync/files"
	"github.com/jastribl/photosync/metadata"
	"github.com/jastribl/photosync/photos"
)

const savedSessionCacheFile = "cache/savedSessionIDForSorting.txt"

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
	rootPicturesDir := strings.TrimRight(args[0], "/") + "/"

	allLocalFilenames := files.GetAllFilenamesInDir(rootPicturesDir)

	pickerSessionID, err := utils.GetPickerSessionID(client, savedSessionCacheFile)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("About to fetch media items from session %s", pickerSessionID)
	mediaItems, err := client.GetAllMediaItemsForSessionID(pickerSessionID)
	if err != nil {
		log.Fatal(err)
	}

	allFilenamesLowerCaseToMediaItems := photos.MediaItemsToLowercaseFilenameMap(mediaItems)

	for _, filename := range allLocalFilenames {
		if mediaItems, found := allFilenamesLowerCaseToMediaItems[strings.ToLower(filename)]; found {
			if len(mediaItems) > 1 {
				log.Printf("Found multiple media items for '%s', leaving untouched\n", filename)
				continue
			}

			filePath := rootPicturesDir + filename
			localDate := localDateForFile(filePath, filename, mediaItems[0].MediaMetadata.CreationTime)

			newFolderPath := rootPicturesDir + localDate + "/"
			if !files.FileExists(newFolderPath) {
				err := os.Mkdir(newFolderPath, 0777)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Created folder: %s\n", newFolderPath)
			}
			log.Printf("Moved '%s' into folder '%s'\n", filename, newFolderPath)
			err = os.Rename(filePath, newFolderPath+filename)
			if err != nil {
				log.Fatal(err)
			}
			continue
		}

		log.Printf("Unable to find Google Photos photo '%s', leaving untouched\n", filename)
	}
}

func localDateForFile(filePath string, filename string, creationTime string) string {
	utcDate := creationTime[:10]

	md := metadata.Extract(filePath, filename)

	if md.HasGPS {
		timezoneName := latlong.LookupZoneName(md.Lat, md.Long)
		location, err := time.LoadLocation(timezoneName)
		if err != nil {
			log.Printf("Warning: could not load timezone '%s' for '%s': %s\n", timezoneName, filename, err)
			return utcDate
		}
		t, err := time.Parse(time.RFC3339, creationTime)
		if err != nil {
			log.Printf("Warning: could not parse creation time for '%s': %s\n", filename, err)
			return utcDate
		}
		localDate := t.In(location).Format("2006-01-02")
		log.Printf("Resolved '%s' to timezone '%s', local date: %s\n", filename, timezoneName, localDate)
		return localDate
	}

	if md.HasDateTime {
		log.Printf("Warning: no GPS data for '%s', using EXIF DateTime\n", filename)
		return md.DateTime.Format("2006-01-02")
	}

	log.Printf("Warning: no GPS or EXIF DateTime for '%s', using UTC date (may be incorrect)\n", filename)
	return utcDate
}
