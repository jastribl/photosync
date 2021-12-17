package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jastribl/photosync/config"
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

	rootPicturesDir := cfg.RootPicturesDir
	allFileNames, err := getAllFileNamesInDir(
		cfg,
		rootPicturesDir,
		cfg.PicturePathSubstringsToIgnore,
	)
	if err != nil {
		log.Fatal(err)
	}

	mediaItmes, err := client.GetAllMediaItems()
	if err != nil {
		log.Fatal(err)
	}

	freeBefore, _ := time.Parse("2006-01-02", cfg.FreeBeforeDate)

	for _, mediaItem := range mediaItmes {
		filename := mediaItem.Filename
		_, found := allFileNames[filename]
		if !found {
			timeOfImage, err := time.Parse(
				"2006-01-02T15:04:05Z",
				mediaItem.MediaMetadata.CreationTime,
			)
			if err != nil {
				log.Fatal(err)
			}
			if timeOfImage.After(freeBefore) {
				log.Println(mediaItem.ProductULR)
			}
		}
	}
}

func getAllFileNamesInDir(cfg *config.Config, rootDir string, folderSubstringsToIgnore []string) (
	map[string]bool,
	error,
) {
	toReturn := make(map[string]bool)

	queue := []string{rootDir}
	for len(queue) > 0 {
		nextItem := queue[0]
		queue = queue[1:]

		files, err := ioutil.ReadDir(nextItem)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if file.IsDir() {
				if !strContainsAnySubstrings(file.Name(), folderSubstringsToIgnore) {
					queue = append(queue, nextItem+file.Name()+"/")
				}
			} else if _, found := cfg.FileNamesToIgnoreMap[file.Name()]; found {
				continue
			} else if file.Name() == ".DS_Store" {
				log.Fatal("wat")
			} else if _, hasKey := toReturn[file.Name()]; !hasKey {
				toReturn[file.Name()] = true
			}
		}
	}

	return toReturn, nil
}

func strContainsAnySubstrings(s string, anySubstrings []string) bool {
	s = strings.ToLower(s)
	for _, test := range anySubstrings {
		if strings.Contains(s, test) {
			return true
		}
	}
	return false
}
