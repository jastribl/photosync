package main

import (
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bradfitz/latlong"
	"github.com/evanoberholster/imagemeta"
	utils "github.com/jastribl/photosync/cmd"
	"github.com/jastribl/photosync/config"
	"github.com/jastribl/photosync/files"
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

	allLocalFilenames := files.GetAllFilenamesInDir(
		rootPicturesDir,
		[]*regexp.Regexp{},
		[]*regexp.Regexp{},
	)

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

type fileMetadata struct {
	lat, long   float64
	hasGPS      bool
	dateTime    time.Time
	hasDateTime bool
}

func localDateForFile(filePath string, filename string, creationTime string) string {
	utcDate := creationTime[:10]

	metadata := extractMetadata(filePath, filename)

	if metadata.hasGPS {
		timezoneName := latlong.LookupZoneName(metadata.lat, metadata.long)
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

	if metadata.hasDateTime {
		log.Printf("Warning: no GPS data for '%s', using EXIF DateTime\n", filename)
		return metadata.dateTime.Format("2006-01-02")
	}

	log.Printf("Warning: no GPS or EXIF DateTime for '%s', using UTC date (may be incorrect)\n", filename)
	return utcDate
}

func extractMetadata(filePath string, filename string) fileMetadata {
	f, err := os.Open(filePath)
	if err != nil {
		return extractMetadataMdls(filePath)
	}
	defer f.Close()

	e, err := imagemeta.Decode(f)
	if err != nil {
		log.Printf("imagemeta unsupported for '%s': %s, trying mdls fallback\n", filename, err)
		return extractMetadataMdls(filePath)
	}

	metadata := fileMetadata{}
	metadata.lat, metadata.long = e.GPS.Latitude(), e.GPS.Longitude()
	metadata.hasGPS = metadata.lat != 0 || metadata.long != 0
	dt := e.SelectedDate()
	if !dt.IsZero() {
		metadata.dateTime = dt
		metadata.hasDateTime = true
	}
	return metadata
}

func extractMetadataMdls(filePath string) fileMetadata {
	metadata := fileMetadata{}

	out, err := exec.Command("mdls",
		"-name", "kMDItemLatitude",
		"-name", "kMDItemLongitude",
		"-name", "kMDItemContentCreationDate",
		filePath,
	).Output()
	if err != nil {
		return metadata
	}

	attrs := parseMdlsOutput(string(out))

	latStr, hasLat := attrs["kMDItemLatitude"]
	longStr, hasLong := attrs["kMDItemLongitude"]
	if hasLat && hasLong {
		lat, latErr := strconv.ParseFloat(latStr, 64)
		long, longErr := strconv.ParseFloat(longStr, 64)
		if latErr == nil && longErr == nil {
			metadata.lat, metadata.long, metadata.hasGPS = lat, long, true
		}
	}

	if dtStr, ok := attrs["kMDItemContentCreationDate"]; ok {
		dt, err := time.Parse("2006-01-02 15:04:05 +0000", dtStr)
		if err == nil {
			metadata.dateTime, metadata.hasDateTime = dt, true
		}
	}

	return metadata
}

func parseMdlsOutput(output string) map[string]string {
	attrs := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if value == "(null)" {
			continue
		}
		attrs[key] = value
	}
	return attrs
}
