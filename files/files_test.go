package files

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
)

func TestStrMatchesAnyAndNotAny(t *testing.T) {
	deny := []*regexp.Regexp{
		regexp.MustCompile(".*[pP]ictures [fF]rom .*$"),
		regexp.MustCompile("^Wendy$"),
	}
	allow := []*regexp.Regexp{
		regexp.MustCompile("^Photos from Michael$"),
	}

	tests := []struct {
		name        string
		s           string
		denyRegexs  []*regexp.Regexp
		allowRegexs []*regexp.Regexp
		want        bool
	}{
		{name: "no deny match", s: "Trip 2021", denyRegexs: deny, allowRegexs: allow, want: false},
		{name: "deny match", s: "Pictures from John", denyRegexs: deny, allowRegexs: allow, want: true},
		{name: "deny match lowercase pictures", s: "pictures from John", denyRegexs: deny, allowRegexs: allow, want: true},
		{name: "photos does not match the pictures pattern", s: "photos from John", denyRegexs: deny, allowRegexs: allow, want: false},
		{name: "exact deny match", s: "Wendy", denyRegexs: deny, allowRegexs: allow, want: true},
		{name: "deny match overridden by allow", s: "Photos from Michael", denyRegexs: []*regexp.Regexp{regexp.MustCompile(".*[pP]hotos [fF]rom .*$")}, allowRegexs: allow, want: false},
		{name: "allow alone does not match", s: "Photos from Michael", denyRegexs: []*regexp.Regexp{}, allowRegexs: allow, want: false},
		{name: "no regexes", s: "anything", denyRegexs: []*regexp.Regexp{}, allowRegexs: []*regexp.Regexp{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrMatchesAnyAndNotAny(tt.s, tt.denyRegexs, tt.allowRegexs); got != tt.want {
				t.Errorf("StrMatchesAnyAndNotAny(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

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
	writeFile("Ignored/d.jpg")

	deny := []*regexp.Regexp{regexp.MustCompile("^Ignored$")}
	got := GetAllFilenamesInDir(dir+"/", deny, []*regexp.Regexp{})
	want := []string{"a.jpg", "b.jpg", "c.png"}
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

	got := GetAllLowercaseFilenamesInDir(dir+"/", []*regexp.Regexp{}, []*regexp.Regexp{})
	want := []string{"img_001.jpg", "photo.png"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetAllLowercaseFilenamesInDir() = %v, want %v", got, want)
	}
}
