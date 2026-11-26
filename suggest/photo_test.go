package suggest

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jastribl/photosync/testutil"
)

func TestLoadPhotosReadsGPSAndTime(t *testing.T) {
	dir := t.TempDir()
	day := time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)
	testutil.WritePhoto(t, dir, "a.jpg", 41.0082, 28.9784, day.Add(19*time.Hour+30*time.Minute))
	testutil.WritePhoto(t, dir, "b.jpg", 0, 0, day.Add(19*time.Hour+32*time.Minute))

	photos, err := LoadPhotos(dir)
	if err != nil {
		t.Fatalf("LoadPhotos: %v", err)
	}
	if len(photos) != 2 {
		t.Fatalf("expected 2 photos, got %d", len(photos))
	}
	if !photos[0].HasGPS || photos[0].Lat != 41.0082 {
		t.Errorf("first photo = %+v, want GPS (41.0082, 28.9784)", photos[0])
	}
	if got := photos[0].Taken.Format("15:04"); got != "19:30" {
		t.Errorf("first photo taken = %s, want 19:30", got)
	}
	if photos[1].HasGPS {
		t.Errorf("second photo should have no GPS, got %+v", photos[1])
	}

	// Inheritance from files: b is 2 minutes after a, so it takes a's spot.
	if !InheritGPS(photos) {
		t.Fatal("expected GPS to be present in folder")
	}
	if !photos[1].HasGPS || photos[1].Lat != 41.0082 {
		t.Errorf("second photo after inheritance = %+v, want a's position", photos[1])
	}
}

func TestLoadPhotosSkipsDSStoreAndSubdirs(t *testing.T) {
	dir := t.TempDir()
	day := time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)
	testutil.WritePhoto(t, dir, "a.jpg", 41.0082, 28.9784, day)
	if err := os.WriteFile(filepath.Join(dir, ".DS_Store"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}

	photos, err := LoadPhotos(dir)
	if err != nil {
		t.Fatalf("LoadPhotos: %v", err)
	}
	if len(photos) != 1 {
		t.Errorf("expected 1 photo, got %d", len(photos))
	}
}
