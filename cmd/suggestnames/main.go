// Command suggestnames proposes names for sortlocal's date folders from
// the places near where each day's photos were taken, and renames the
// folders with the confirmed picks.
//
// Usage:
//
//	go run cmd/suggestnames/main.go [flags] <root dir>
//
// The root dir holds YYYY-MM-DD folders (as sortlocal produces them).
// Folders already named "YYYY-MM-DD - <description>" are handled the
// same way: the date prefix is the day, and the description part is
// replaced by the new picks.
//
// For each folder, photos are grouped into clusters (photos near each
// other in space and time are one stop — one restaurant, one sight),
// and one Google Places Nearby Search is made per cluster. You pick
// from a shortlist per cluster; the proposed name is
// "YYYY-MM-DD - <picked names joined with ', '>".
//
// A proposed name is capped at 255 bytes (the macOS APFS limit for a
// single path component): trailing places are dropped until it fits,
// and a single over-long place is truncated at a UTF-8 rune boundary,
// so the rename can never fail on length.
//
// The Places API needs an API key: pass --api-key, set
// GOOGLE_PLACES_API_KEY, or put "places-api-key" in config/config.json
// (see config/example-config.json). --stub answers with canned places
// and no network, for offline testing. --dry-run prints suggestions and
// proposed names without renaming (it still calls Places unless
// --stub). --yes picks the nearest candidate per cluster and renames
// without prompting.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jastribl/photosync/suggest"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	apiKey := flag.String("api-key", "", "Google Places API key (or set GOOGLE_PLACES_API_KEY)")
	stub := flag.Bool("stub", false, "use canned places with no network calls")
	dryRun := flag.Bool("dry-run", false, "print suggestions and proposed names without renaming")
	autoYes := flag.Bool("yes", false, "pick the nearest candidate per cluster and rename without prompting")
	timeGap := flag.Duration("time-gap", 45*time.Minute, "photos more than this apart in time start a new cluster")
	distanceM := flag.Float64("distance-m", 100, "photos more than this far (meters) from the cluster center start a new cluster")
	radiusM := flag.Float64("radius-m", 200, "Places search radius in meters around each cluster")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] <root dir of date folders>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}

	var client suggest.PlacesClient
	if *stub {
		client = suggest.NewStubClient()
	} else {
		key := resolveAPIKey(*apiKey)
		if key == "" {
			log.Fatal("no Places API key: pass --api-key, set GOOGLE_PLACES_API_KEY, or put \"places-api-key\" in config/config.json (or use --stub for offline testing)")
		}
		client = suggest.NewRESTClient(key)
	}

	cfg := suggest.Config{
		RootDir:   args[0],
		TimeGap:   *timeGap,
		DistanceM: *distanceM,
		RadiusM:   *radiusM,
		DryRun:    *dryRun,
		AutoYes:   *autoYes,
		Places:    client,
	}
	if _, err := suggest.Run(cfg, os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

// resolveAPIKey finds the Places API key in priority order: the
// --api-key flag, then GOOGLE_PLACES_API_KEY, then the "places-api-key"
// field of config/config.json (the repo's gitignored config file, read
// from the current directory like the other tools do). A missing or
// unreadable config file is not an error — it just means no key there.
func resolveAPIKey(flagKey string) string {
	if flagKey != "" {
		return flagKey
	}
	if key := os.Getenv("GOOGLE_PLACES_API_KEY"); key != "" {
		return key
	}
	data, err := os.ReadFile("config/config.json")
	if err != nil {
		return ""
	}
	var cfg struct {
		PlacesAPIKey string `json:"places-api-key"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ""
	}
	return cfg.PlacesAPIKey
}
