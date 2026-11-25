package main

import (
	"fmt"
	"log"
	"os"
	"strings"

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
	if !strings.HasSuffix(rootPicturesDir, "/") {
		rootPicturesDir += "/"
	}
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

	// folderItemsByLabel maps each label back to its folder's picked
	// photos, so the run can tell whether a folder has any photos in the
	// album at all (see the skip rule below).
	folderItemsByLabel := map[string][]*photos.MediaItem{}
	for _, folderInfo := range listOfFolderInfo {
		if folderInfo.NumMediaItems == 0 || folderInfo.Path == rootPicturesDir {
			continue
		}
		labelText := folderInfo.Path[len(rootPicturesDir) : len(folderInfo.Path)-1]
		folderItemsByLabel[labelText] = folderInfo.MediaItems
	}

	// knownMissing collects the IDs of photos proven not to be in the
	// album: each one failed as an anchor with the not-in-album error.
	// A folder whose picked photos are all in here has nothing in the
	// album, so its label is skipped instead of added. Labels run
	// oldest-first and each label's chain starts with the next folder's
	// photos, so a folder is probed by the labels before it — no
	// reordering needed. One limit: the first folder is never probed
	// (no earlier label). Every other folder is fully probed — its
	// whole chain is tried — so a skip is conclusive, not a guess.
	// Dry runs never scan: they make no API calls, so nothing is probed
	// and every label is shown.
	knownMissing := map[string]bool{}
	failed := 0
	skipped := 0
	for _, placement := range plan {
		if !createLabels {
			describePlacement(placement)
			continue
		}
		if labelling.FolderProvenEmpty(folderItemsByLabel[placement.Text], knownMissing) {
			log.Printf("Skipping label '%s': no photos from this folder are in the album", placement.Text)
			skipped++
			continue
		}
		if err := addLabelWithFallback(client, album.ID, placement, knownMissing); err != nil {
			log.Printf("unable to add label '%s': %v", placement.Text, err)
			failed++
		}
	}
	if failed > 0 {
		log.Fatalf("%d of %d labels failed", failed, len(plan))
	}
	log.Printf("Added %d of %d labels", len(plan)-failed-skipped, len(plan))
	if skipped > 0 {
		log.Printf("Skipped %d label(s): no photos from those folders are in the album", skipped)
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

// addLabelWithFallback tries each anchor candidate in order. A label with
// no anchors is placed at the front of the album — that is its real
// placement (the last folder has no later folder to anchor after), not a
// fallback. A label with anchors that all fail is not added at all: putting
// it at the front would attach a day's label to the wrong place, so the
// error is reported and the caller moves on to the next label.
//
// Every anchor that fails with the not-in-album error is recorded in
// knownMissing: that failure proves the photo is not in the album, and
// the caller uses it to skip later labels whose whole folder is proven
// missing. Any other error is inconclusive and records nothing.
//
// Logging is one line per label at most: a single "Added label ..." line
// when the label lands on its first anchor or at the front, a single
// warning when a fallback anchor is the one that worked, a single error
// from the caller when the label could not be added at all. Failed
// attempts are not logged individually.
func addLabelWithFallback(client *photos.Client, albumID string, placement labelling.LabelPlacement, knownMissing map[string]bool) error {
	if len(placement.Anchors) == 0 {
		_, err := client.AddTextEnrichmentToAlbum(albumID, nil, placement.Text)
		if err != nil {
			return err
		}
		log.Printf("Added label '%s' at the beginning of the album", placement.Text)
		return nil
	}
	var lastErr error
	for i, anchor := range placement.Anchors {
		_, err := client.AddTextEnrichmentToAlbum(albumID, anchor, placement.Text)
		if err == nil {
			if i > 0 {
				log.Printf("Label '%s' fell back to anchor '%s'", placement.Text, anchor.Filename)
			} else {
				log.Printf("Added label '%s' after '%s'", placement.Text, anchor.Filename)
			}
			return nil
		}
		if photos.IsInvalidMediaItemID(err) {
			knownMissing[anchor.ID] = true
		}
		lastErr = err
	}
	return lastErr
}
