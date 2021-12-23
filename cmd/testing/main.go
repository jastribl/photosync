package main

import (
	"fmt"
	"log"
	"os"

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

	album, err := client.GetAlbumWithTitleContains(albumName)
	if err != nil {
		log.Fatal(err)
	}

	if album == nil {
		log.Fatalln("Album not found with name '" + albumName + "'")
	}

	mediaItems, err := client.GetAllMediaItemsForAlbum(album)
	if err != nil {
		log.Fatal(err)
	}

	mediaItemFileNames := []string{}
	for _, item := range mediaItems {
		mediaItemFileNames = append(mediaItemFileNames, item.Filename)
	}

	allDriveFileNames, err := files.GetAllFileNamesInDir(
		rootPicturesDir,
		[]string{},
		map[string]bool{},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("mediaItemFileNames:")
	for _, name := range mediaItemFileNames {
		fmt.Println(name)
	}
	fmt.Println("allDriveFileNames:")
	for _, name := range allDriveFileNames {
		fmt.Println(name)
	}
	return

	if len(mediaItemFileNames) != len(allDriveFileNames) {
		log.Fatalln("not the same length")
	}

	for i, mediaItemFileName := range mediaItemFileNames {
		if mediaItemFileName != allDriveFileNames[i] {
			log.Fatal("not matching")
		}
	}
	log.Println("All good!")
}
