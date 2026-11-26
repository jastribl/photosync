// Package suggest proposes folder names for a day's photos from the
// places near where they were taken. It is the naming step of the
// labelling workflow: sortlocal sorts photos into YYYY-MM-DD folders,
// suggestnames proposes "YYYY-MM-DD - <places>" names for them, and
// labelphotos then puts those folder names into albums as labels.
//
// Places data comes from the Google Places API (New) Nearby Search —
// one call per GPS cluster, never per photo, and never with image
// recognition. See places.go for the API shape and cluster.go for how
// photos are grouped into clusters before any call is made. Clusters
// are in time order (oldest first), so the picks — and the folder name
// built from them — list the day's stops in the order they were visited.
package suggest

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jastribl/photosync/metadata"
)

// Photo is one local photo file with the position and taken time read
// from its EXIF data. Taken falls back to the file's modification time
// when the photo has no EXIF date, so clustering always has a time to
// work with; HasGPS says whether Lat/Long are real.
type Photo struct {
	Path    string
	Name    string
	Lat     float64
	Long    float64
	HasGPS  bool
	Taken   time.Time
	HasEXIF bool
}

// LoadPhotos reads every file in dir (top level only, .DS_Store skipped)
// and returns one Photo per file, in filename order.
func LoadPhotos(dir string) ([]Photo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	photos := []Photo{}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == ".DS_Store" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		md := metadata.Extract(path, entry.Name())
		photo := Photo{
			Path:    path,
			Name:    entry.Name(),
			Lat:     md.Lat,
			Long:    md.Long,
			HasGPS:  md.HasGPS,
			HasEXIF: md.HasDateTime,
		}
		if md.HasDateTime {
			photo.Taken = md.DateTime
		} else if info, err := entry.Info(); err == nil {
			photo.Taken = info.ModTime()
		}
		photos = append(photos, photo)
	}
	sort.Slice(photos, func(i, j int) bool { return photos[i].Name < photos[j].Name })
	return photos, nil
}

// InheritGPS gives every photo without GPS the position of the photo
// nearest to it in time within the same folder that does have GPS: a
// meal photo taken two minutes after a geotagged one at the same table
// is at the same place. Photos keep their own position when they have
// one. It returns whether any photo in the folder has GPS at all (after
// inheritance, either all do or none do).
func InheritGPS(photos []Photo) bool {
	hasAny := false
	for _, p := range photos {
		if p.HasGPS {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return false
	}
	for i := range photos {
		if photos[i].HasGPS {
			continue
		}
		best := -1
		var bestGap time.Duration
		for j := range photos {
			if !photos[j].HasGPS {
				continue
			}
			gap := photos[i].Taken.Sub(photos[j].Taken)
			if gap < 0 {
				gap = -gap
			}
			if best == -1 || gap < bestGap {
				best, bestGap = j, gap
			}
		}
		if best != -1 {
			photos[i].Lat, photos[i].Long, photos[i].HasGPS = photos[best].Lat, photos[best].Long, true
		}
	}
	return true
}

// HasAnyGPS reports whether any photo in the list has GPS coordinates.
func HasAnyGPS(photos []Photo) bool {
	for _, p := range photos {
		if p.HasGPS {
			return true
		}
	}
	return false
}
