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

	allMediaItems, err := client.GetAllMediaItems(true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Old Cache Size: %d\n", len(allMediaItems))

	allMediaItems, err = client.CacheAllMediaItems()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("New Cache Size: %d\n", len(allMediaItems))
}
