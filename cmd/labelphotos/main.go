package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/jastribl/photosync/config"
	"github.com/jastribl/photosync/labelling"
	"github.com/jastribl/photosync/photos"
)

const savedSessionCacheFile = "cache/savedSessionID.txt"

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
	createLabels := false
	if len(args) > 2 && args[2] == "--create" {
		createLabels = true
	}
	fmt.Println("Running for the following input")
	fmt.Println("Root picture dir: '" + rootPicturesDir + "'")

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
	// mediaItems, err := client.GetAllMediaItemsForSessionID("d8671d27-fe5f-47c1-924a-7e2682b12741")
	if err != nil {
		log.Fatal(err)
	}

	album, err := client.GetAlbumWithTitle(albumName)
	if err != nil {
		log.Fatal(err)
	}
	if album == nil {
		log.Fatalln("Album not found with name '" + albumName + "'")
	}

	listOfFolderInfo := labelling.GetTopLevelFolderInfo(rootPicturesDir, client, mediaItems)

	for i, folderInfo := range listOfFolderInfo {
		if folderInfo.NumMediaItems == 0 {
			log.Printf("No pictures from '%s' in the album '%s', skipping\n", folderInfo.Path, albumName)
			continue
		}
		if folderInfo.Path == rootPicturesDir {
			log.Println("Not adding label for root dir")
			continue
		}
		if labelling.ShouldIgnoreFolder(folderInfo.Path) {
			log.Printf("Not adding label for path: %s\n", folderInfo.Path)
			continue
		}
		labelText := folderInfo.Path[len(rootPicturesDir) : len(folderInfo.Path)-1]
		if i != len(listOfFolderInfo)-1 {
			var lastMediaItemOfNextFolder *photos.MediaItem = nil
			nextFolderPath := ""
			for j := i + 1; lastMediaItemOfNextFolder == nil && j < len(listOfFolderInfo); j += 1 {
				nextFolderInfo := listOfFolderInfo[j]
				nextFolderPath = nextFolderInfo.Path
				lastMediaItemOfNextFolder = nextFolderInfo.LastMediaItem
			}
			if lastMediaItemOfNextFolder == nil {
				log.Printf(
					"Got no first pic of next folder for '%s', so we must label the front of the album instead\n",
					labelText,
				)
				goto LABEL_BEGINNING_OF_ALBUM
			}

			afterMediaID := lastMediaItemOfNextFolder.ID
			if afterMediaID == "" {
				log.Fatalln("Got empty media id for: " + lastMediaItemOfNextFolder.Filename)
			}
			fmt.Printf(
				"Adding '%s' after '%s' (last pic of folder '%s') \n",
				labelText,
				lastMediaItemOfNextFolder.Filename,
				nextFolderPath,
			)
			if !createLabels {
				continue
			}
			_, err := client.AddTextEnrichmentToAlbum(album.ID, lastMediaItemOfNextFolder, labelText)
			if err != nil {
				log.Fatal(err)
			}
			continue
		}

	LABEL_BEGINNING_OF_ALBUM:
		fmt.Printf("Adding '%s' at the beginning of the album\n", labelText)
		if !createLabels {
			continue
		}
		_, err := client.AddTextEnrichmentToAlbum(album.ID, nil, labelText)
		if err != nil {
			log.Fatal(err)
		}
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
