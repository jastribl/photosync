package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

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
	rootPicturesDir := args[0]

	allLocalFilenames := files.GetAllFilenamesInDir(
		rootPicturesDir,
		[]*regexp.Regexp{},
		[]*regexp.Regexp{},
	)

	var pickerSessionID string
	_, err = os.Stat(savedSessionCacheFile)
	cacheDoesNotExist := errors.Is(err, os.ErrNotExist)
	if cacheDoesNotExist || !yesNo("Would you like to re-use your previous picker session?") {
		pickerSession, err := client.NewPickerSession()
		if err != nil {
			log.Fatal(err)
		}

		pickerSession, err = client.WaitForSession(pickerSession)
		if err != nil {
			log.Fatal(err)
		}

		pickerSessionID = pickerSession.ID
		f, err := os.Create(savedSessionCacheFile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		w.WriteString(pickerSessionID)
		w.Flush()
	} else if !cacheDoesNotExist {
		// here we know we failed the top condition because we wanted to use the cache
		out, err := ioutil.ReadFile(savedSessionCacheFile)
		if err != nil {
			log.Fatal(err)
		}
		pickerSessionID = string(out)
	}

	log.Printf("About to fetch media items from session %s", pickerSessionID)
	mediaItems, err := client.GetAllMediaItemsForSessionID(pickerSessionID)
	if err != nil {
		log.Fatal(err)
	}

	allFilenamesLowerCaseToMediaItems := make(map[string][]*photos.MediaItem)
	for _, mediaItem := range mediaItems {
		lowercaseFilename := strings.ToLower(mediaItem.Filename)
		tempItems, found := allFilenamesLowerCaseToMediaItems[lowercaseFilename]
		if !found {
			tempItems = []*photos.MediaItem{}
		}
		tempItems = append(tempItems, mediaItem)
		allFilenamesLowerCaseToMediaItems[lowercaseFilename] = tempItems
	}

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

// todo: make this a util
func yesNo(q string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {

		fmt.Printf("%s[Y]es|[N]o?", q)
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
