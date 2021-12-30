package main

import (
	"fmt"
	"log"
	"os"

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

	allMediaItems, err := client.GetAllMediaItemsWithCache()
	if err != nil {
		log.Fatal(err)
	}
	oldCacheSize := len(allMediaItems)
	fmt.Printf("Old Cache Size: %d\n", oldCacheSize)

	allMediaItems, err = client.CacheAndReturnAllMediaItems()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Old Cache Size: %d\n", oldCacheSize)
	fmt.Printf("New Cache Size: %d\n", len(allMediaItems))
}
