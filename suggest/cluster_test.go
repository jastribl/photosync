package suggest

import (
	"testing"
	"time"
)

func photoAt(name string, lat, lng float64, taken time.Time) Photo {
	return Photo{Name: name, Lat: lat, Long: lng, HasGPS: true, Taken: taken}
}

func TestClusterPhotosSplitsOnTimeGap(t *testing.T) {
	// Same place, two time blocks 3 hours apart: two clusters.
	day := time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)
	photos := []Photo{
		photoAt("a.jpg", 41.0082, 28.9784, day.Add(15*time.Hour)),
		photoAt("b.jpg", 41.0082, 28.9784, day.Add(15*time.Hour+20*time.Minute)),
		photoAt("c.jpg", 41.0082, 28.9784, day.Add(19*time.Hour)),
		photoAt("d.jpg", 41.0082, 28.9784, day.Add(19*time.Hour+30*time.Minute)),
	}
	clusters := ClusterPhotos(photos, 45*time.Minute, 100)
	if len(clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(clusters))
	}
	if len(clusters[0].Photos) != 2 || len(clusters[1].Photos) != 2 {
		t.Errorf("expected 2 photos per cluster, got %d and %d", len(clusters[0].Photos), len(clusters[1].Photos))
	}
}

func TestClusterPhotosSplitsOnDistance(t *testing.T) {
	// Same time block, two places ~1.5 km apart: two clusters.
	day := time.Date(2026, 11, 27, 19, 0, 0, 0, time.UTC)
	photos := []Photo{
		photoAt("a.jpg", 41.0082, 28.9784, day),
		photoAt("b.jpg", 41.0082, 28.9784, day.Add(10*time.Minute)),
		photoAt("c.jpg", 41.0200, 28.9900, day.Add(20*time.Minute)),
		photoAt("d.jpg", 41.0200, 28.9900, day.Add(30*time.Minute)),
	}
	clusters := ClusterPhotos(photos, 45*time.Minute, 100)
	if len(clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(clusters))
	}
}

func TestClusterPhotosHikeStaysSplitByTime(t *testing.T) {
	// Photos spread along a trail over hours: the time gap keeps them
	// from merging into one giant cluster even though each step is
	// within the distance threshold of the last.
	day := time.Date(2026, 11, 27, 9, 0, 0, 0, time.UTC)
	photos := []Photo{
		photoAt("a.jpg", 41.0000, 28.9700, day),
		photoAt("b.jpg", 41.0010, 28.9710, day.Add(50*time.Minute)),
		photoAt("c.jpg", 41.0020, 28.9720, day.Add(100*time.Minute)),
	}
	clusters := ClusterPhotos(photos, 45*time.Minute, 100)
	if len(clusters) != 3 {
		t.Fatalf("expected 3 clusters, got %d", len(clusters))
	}
}

func TestClusterCenterIsMedian(t *testing.T) {
	day := time.Date(2026, 11, 27, 19, 0, 0, 0, time.UTC)
	photos := []Photo{
		photoAt("a.jpg", 41.0000, 28.9700, day),
		photoAt("b.jpg", 41.0010, 28.9710, day.Add(5*time.Minute)),
		// One fix noticeably off to one side: the median stays on the
		// middle photo, where the mean would be pulled toward it.
		photoAt("c.jpg", 41.0100, 28.9800, day.Add(10*time.Minute)),
	}
	clusters := ClusterPhotos(photos, 45*time.Minute, 10000)
	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(clusters))
	}
	if clusters[0].CenterLat != 41.0010 || clusters[0].CenterLong != 28.9710 {
		t.Errorf("center = (%v, %v), want (41.0010, 28.9710)", clusters[0].CenterLat, clusters[0].CenterLong)
	}
}

func TestInheritGPS(t *testing.T) {
	day := time.Date(2026, 11, 27, 19, 0, 0, 0, time.UTC)
	photos := []Photo{
		photoAt("a.jpg", 41.0082, 28.9784, day),
		{Name: "b.jpg", Taken: day.Add(2 * time.Minute)},
		photoAt("c.jpg", 41.0090, 28.9790, day.Add(60*time.Minute)),
	}
	if !InheritGPS(photos) {
		t.Fatal("expected GPS to be present")
	}
	if !photos[1].HasGPS || photos[1].Lat != 41.0082 || photos[1].Long != 28.9784 {
		t.Errorf("photo b should inherit a's position, got %+v", photos[1])
	}
}

func TestInheritGPSNone(t *testing.T) {
	day := time.Date(2026, 11, 27, 19, 0, 0, 0, time.UTC)
	photos := []Photo{
		{Name: "a.jpg", Taken: day},
		{Name: "b.jpg", Taken: day.Add(time.Minute)},
	}
	if InheritGPS(photos) {
		t.Error("expected no GPS in folder")
	}
	if HasAnyGPS(photos) {
		t.Error("expected HasAnyGPS to be false")
	}
}

func TestHaversineM(t *testing.T) {
	// ~111 m per 0.001 degrees of latitude.
	d := haversineM(41.0082, 28.9784, 41.0092, 28.9784)
	if d < 100 || d > 125 {
		t.Errorf("haversine = %v m, want ~111 m", d)
	}
	if got := haversineM(41.0082, 28.9784, 41.0082, 28.9784); got != 0 {
		t.Errorf("haversine of a point with itself = %v, want 0", got)
	}
}
