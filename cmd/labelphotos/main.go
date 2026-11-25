package main

import (
	"fmt"
	"log"
	"os"

	utils "github.com/jastribl/photosync/cmd"
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

	// The photos to label still come from a Picker session: since the
	// March 2025 Library API change, the app can only list media items it
	// created itself, so it cannot see phone-uploaded photos in the album
	// via the API — the user picking them is the only way the app learns
	// about them. See labelling.BuildLabelPlan for how the placements are
	// derived from the picked set.
	pickerSessionID, err := utils.GetPickerSessionID(client, savedSessionCacheFile)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("About to fetch media items from session %s", pickerSessionID)
	mediaItems, err := client.GetAllMediaItemsForSessionID(pickerSessionID)
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

	// Build the whole label plan before writing anything, so a failure
	// part-way through cannot leave the remaining labels unplanned.
	plan := labelling.BuildLabelPlan(listOfFolderInfo, rootPicturesDir)
	if len(plan) == 0 {
		log.Println("No labels to add")
		return
	}

	failed := 0
	for _, placement := range plan {
		if !createLabels {
			describePlacement(placement)
			continue
		}
		if err := addLabelWithFallback(client, album.ID, placement); err != nil {
			log.Printf("FAILED to add label '%s': %v", placement.Text, err)
			failed++
		}
	}
	if failed > 0 {
		log.Fatalf("%d of %d labels failed", failed, len(plan))
	}
}

func describePlacement(placement labelling.LabelPlacement) {
	if len(placement.Anchors) > 0 {
		fmt.Printf(
			"Would add '%s' after '%s'",
			placement.Text,
			placement.Anchors[0].Filename,
		)
		if len(placement.Anchors) > 1 {
			fmt.Printf(" (with %d fallback anchor(s))", len(placement.Anchors)-1)
		}
		fmt.Println()
		return
	}
	fmt.Printf("Would add '%s' at the beginning of the album\n", placement.Text)
}

// addLabelWithFallback tries each anchor candidate in order and finally the
// front of the album. A label whose preferred anchor photo is not in the
// album no longer aborts the run: it falls back to the next folder's photo,
// and only fails if no candidate at all works.
func addLabelWithFallback(client *photos.Client, albumID string, placement labelling.LabelPlacement) error {
	candidates := append([]*photos.MediaItem{}, placement.Anchors...)
	candidates = append(candidates, nil) // nil = FIRST_IN_ALBUM
	var lastErr error
	for i, anchor := range candidates {
		_, err := client.AddTextEnrichmentToAlbum(albumID, anchor, placement.Text)
		if err == nil {
			if i > 0 {
				if anchor == nil {
					log.Printf("Label '%s' fell back to the front of the album", placement.Text)
				} else {
					log.Printf("Label '%s' fell back to anchor '%s'", placement.Text, anchor.Filename)
				}
			}
			return nil
		}
		lastErr = err
	}
	return lastErr
}
