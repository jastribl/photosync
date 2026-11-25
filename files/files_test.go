package files

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(existing, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if !FileExists(existing) {
		t.Errorf("FileExists(%q) = false, want true", existing)
	}
	if FileExists(filepath.Join(dir, "missing.txt")) {
		t.Errorf("FileExists(missing) = true, want false")
	}
	if FileExists(dir) {
		// FileExists only stats the path; a directory exists too.
		t.Logf("FileExists(%q) = true for a directory (stat only)", dir)
	}
}

func TestGetAllFilenamesInDir(t *testing.T) {
	dir := t.TempDir()
	writeFile := func(relPath string) {
		t.Helper()
		fullPath := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeFile("a.jpg")
	writeFile(".DS_Store")
	writeFile("Sub/b.jpg")
	writeFile("Sub/c.png")
	writeFile("Other/d.jpg")

	// No folders are skipped (the old deny/allow rules are gone);
	// .DS_Store is still skipped. Order is breadth-first, with
	// directory entries in sorted order.
	got := GetAllFilenamesInDir(dir + "/")
	want := []string{"a.jpg", "d.jpg", "b.jpg", "c.png"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetAllFilenamesInDir() = %v, want %v", got, want)
	}
}

func TestGetAllLowercaseFilenamesInDir(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"IMG_001.JPG", "Photo.PNG"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	got := GetAllLowercaseFilenamesInDir(dir + "/")
	want := []string{"img_001.jpg", "photo.png"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetAllLowercaseFilenamesInDir() = %v, want %v", got, want)
	}
}
