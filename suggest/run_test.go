package suggest

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jastribl/photosync/testutil"
)

func testConfig(root string) Config {
	return Config{
		RootDir:   root,
		TimeGap:   45 * time.Minute,
		DistanceM: 100,
		RadiusM:   200,
		Places:    NewStubClient(),
	}
}

type testPhoto struct {
	name    string
	lat     float64
	lng     float64
	minutes int
}

// writeDayFolder creates root/folder with one generated JPEG per
// testPhoto, taken that many minutes after midnight on the folder's
// date, and returns the folder path.
func writeDayFolder(t *testing.T, root, folder string, photos ...testPhoto) string {
	t.Helper()
	dir := filepath.Join(root, folder)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	day, _ := time.Parse("2006-01-02", folder[:10])
	for _, p := range photos {
		taken := day.Add(time.Duration(p.minutes) * time.Minute)
		testutil.WritePhoto(t, dir, p.name, p.lat, p.lng, taken)
	}
	return dir
}

func TestRunAutoYesRenamesAndDedupes(t *testing.T) {
	root := t.TempDir()
	// Two clusters: afternoon at a park, evening at a restaurant.
	// --yes picks the first (nearest) stub candidate for each; both
	// pick "Stub Restaurant", so the name dedupes to one entry.
	writeDayFolder(t, root, "2026-11-27",
		testPhoto{"a.jpg", 41.0200, 28.9900, 15 * 60},
		testPhoto{"b.jpg", 41.0200, 28.9900, 15*60 + 20},
		testPhoto{"c.jpg", 41.0082, 28.9784, 19 * 60},
		testPhoto{"d.jpg", 41.0082, 28.9784, 19*60 + 30})

	cfg := testConfig(root)
	cfg.AutoYes = true
	var out bytes.Buffer
	summary, err := Run(cfg, strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if summary.Renamed != 1 || summary.Skipped != 0 || summary.NoGPS != 0 {
		t.Errorf("summary = %+v, want {Renamed:1 Skipped:0 NoGPS:0}", summary)
	}
	if _, err := os.Stat(filepath.Join(root, "2026-11-27 Stub Restaurant")); err != nil {
		t.Errorf("expected folder to be renamed, output:\n%s", out.String())
	}
	if _, err := os.Stat(filepath.Join(root, "2026-11-27")); !os.IsNotExist(err) {
		t.Error("old folder name should be gone")
	}
}

func TestRunInteractivePickAndConfirm(t *testing.T) {
	root := t.TempDir()
	writeDayFolder(t, root, "2026-11-27",
		testPhoto{"a.jpg", 41.0082, 28.9784, 19 * 60})

	cfg := testConfig(root)
	var out bytes.Buffer
	// Pick #2 (Stub Cafe), then confirm the rename.
	summary, err := Run(cfg, strings.NewReader("2\ny\n"), &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if summary.Renamed != 1 {
		t.Errorf("summary = %+v, want Renamed 1; output:\n%s", summary, out.String())
	}
	if _, err := os.Stat(filepath.Join(root, "2026-11-27 Stub Cafe")); err != nil {
		t.Errorf("expected folder '2026-11-27 Stub Cafe', output:\n%s", out.String())
	}
}

func TestRunInteractiveCustomNameAndDecline(t *testing.T) {
	root := t.TempDir()
	writeDayFolder(t, root, "2026-11-27",
		testPhoto{"a.jpg", 41.0082, 28.9784, 19 * 60})

	cfg := testConfig(root)
	var out bytes.Buffer
	// Type a custom name, then decline the rename.
	summary, err := Run(cfg, strings.NewReader("Araf İstanbul\nn\n"), &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if summary.Renamed != 0 || summary.Skipped != 1 {
		t.Errorf("summary = %+v, want {Renamed:0 Skipped:1}", summary)
	}
	if _, err := os.Stat(filepath.Join(root, "2026-11-27")); err != nil {
		t.Errorf("folder should be unchanged, output:\n%s", out.String())
	}
}

func TestRunYesAtClusterPromptIsNotAName(t *testing.T) {
	root := t.TempDir()
	writeDayFolder(t, root, "2026-11-27",
		testPhoto{"a.jpg", 41.0082, 28.9784, 19 * 60})

	cfg := testConfig(root)
	var out bytes.Buffer
	// "y" at the cluster prompt is meant for the rename confirmation,
	// not a place name: the cluster is skipped, not named "y".
	summary, err := Run(cfg, strings.NewReader("y\n"), &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if summary.Renamed != 0 || summary.Skipped != 1 {
		t.Errorf("summary = %+v, want {Renamed:0 Skipped:1}", summary)
	}
	if _, err := os.Stat(filepath.Join(root, "2026-11-27")); err != nil {
		t.Errorf("folder should be unchanged, output:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Not a place name") {
		t.Errorf("expected a 'Not a place name' note, output:\n%s", out.String())
	}
}

func TestRunDryRunDoesNotRename(t *testing.T) {
	root := t.TempDir()
	writeDayFolder(t, root, "2026-11-27",
		testPhoto{"a.jpg", 41.0082, 28.9784, 19 * 60})

	cfg := testConfig(root)
	cfg.AutoYes = true
	cfg.DryRun = true
	var out bytes.Buffer
	summary, err := Run(cfg, strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if summary.Renamed != 1 {
		t.Errorf("summary = %+v, want Renamed 1 (would-be)", summary)
	}
	if _, err := os.Stat(filepath.Join(root, "2026-11-27")); err != nil {
		t.Errorf("dry run must not rename, output:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Would rename") {
		t.Errorf("expected a 'Would rename' line, output:\n%s", out.String())
	}
}

func TestRunNoGPSFolderSkipped(t *testing.T) {
	root := t.TempDir()
	writeDayFolder(t, root, "2026-11-27",
		testPhoto{"a.jpg", 0, 0, 19 * 60})

	cfg := testConfig(root)
	cfg.AutoYes = true
	var out bytes.Buffer
	summary, err := Run(cfg, strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if summary.NoGPS != 1 || summary.Renamed != 0 {
		t.Errorf("summary = %+v, want {NoGPS:1 Renamed:0}", summary)
	}
	if _, err := os.Stat(filepath.Join(root, "2026-11-27")); err != nil {
		t.Errorf("folder should be unchanged, output:\n%s", out.String())
	}
}

func TestRunAlreadyNamedFolderGetsNewDescription(t *testing.T) {
	root := t.TempDir()
	writeDayFolder(t, root, "2026-11-27 Old description",
		testPhoto{"a.jpg", 41.0082, 28.9784, 19 * 60})

	cfg := testConfig(root)
	cfg.AutoYes = true
	var out bytes.Buffer
	summary, err := Run(cfg, strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if summary.Renamed != 1 {
		t.Errorf("summary = %+v, want Renamed 1", summary)
	}
	if _, err := os.Stat(filepath.Join(root, "2026-11-27 Stub Restaurant")); err != nil {
		t.Errorf("expected the description to be replaced, output:\n%s", out.String())
	}
}
