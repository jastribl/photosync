package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jastribl/photosync/config"
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

	allMediaItems, err := client.GetAllMediaItems(true)
	utils.FatalError(err)
	fmt.Printf("Old Cache Size: %d\n", len(allMediaItems))

	allMediaItems, err = client.CacheAllMediaItems()
	utils.FatalError(err)

	fmt.Printf("New Cache Size: %d\n", len(allMediaItems))
}
