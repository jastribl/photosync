package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jastribl/photosync/config"
	"github.com/jastribl/photosync/labelling"
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
	createLabels := false
	if len(args) > 2 && args[2] == "--create" {
		createLabels = true
	}
	fmt.Println("Running for the following input")
	fmt.Println("Root picture dir: '" + rootPicturesDir + "'")
	fmt.Println("Album Name: '" + albumName + "'")

	album, err := client.GetAlbumWithTitleContains(albumName)
	if err != nil {
		log.Fatal(err)
	}

	if album == nil {
		log.Fatalln("Album not found with name '" + albumName + "'")
	}

	listOfFolderInfo := labelling.GetTopLevelFolderInfo(rootPicturesDir, client, album)

	for i, folderInfo := range listOfFolderInfo {
		if folderInfo.Path == rootPicturesDir {
			log.Println("Not adding label for root")
			continue
		}
		if labelling.ShouldIgnoreFolder(folderInfo.Path) {
			log.Printf("Not adding label for path: %s\n", folderInfo.Path)
			continue
		}
		labelText := folderInfo.Path[len(rootPicturesDir) : len(folderInfo.Path)-1]
		if i == len(listOfFolderInfo)-1 {
			fmt.Printf("Adding '%s' at the beginning of the album\n", labelText)
			if !createLabels {
				break
			}
			_, err := client.AddTextEnrichmentToAlbum(album.ID, nil, labelText)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			var lastMediaItemOfNextFolder *photos.MediaItem = nil
			nextFolderPath := ""
			for j := i + 1; lastMediaItemOfNextFolder == nil && j < len(listOfFolderInfo); j += 1 {
				nextFolderInfo := listOfFolderInfo[j]
				nextFolderPath = nextFolderInfo.Path
				lastMediaItemOfNextFolder = nextFolderInfo.LastMediaItem
			}
			if lastMediaItemOfNextFolder == nil {
				log.Fatalln("Got no first pic of next folder for folder")
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
				break
			}
			_, err := client.AddTextEnrichmentToAlbum(album.ID, lastMediaItemOfNextFolder, labelText)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
