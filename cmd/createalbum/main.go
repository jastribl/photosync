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

	args := os.Args[1:]
	title := args[0]
	fmt.Println("Creating new album: '" + title + "'")

	album, err := client.CreateAlbum(title)
	if err != nil {
		log.Fatal(err)
	}
	if album == nil {
		log.Fatal("Error creating album")
	}
	log.Printf("%#v\n", album)
}
