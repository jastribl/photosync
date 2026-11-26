package suggest

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestParseFolderDate(t *testing.T) {
	cases := []struct {
		name string
		want string
		ok   bool
	}{
		{"2026-11-27", "2026-11-27", true},
		{"2026-11-27 Araf İstanbul", "2026-11-27", true},
		{"2026-11-27 - Araf İstanbul", "2026-11-27", true},
		{"2026-11-27  ", "2026-11-27", true},
		{"2026-13-40", "", false},
		{"2026-11-2", "", false},
		{"Nov 27 2026", "", false},
		{"2026-11-27Araf", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		if got, ok := ParseFolderDate(tc.name); got != tc.want || ok != tc.ok {
			t.Errorf("ParseFolderDate(%q) = (%q, %v), want (%q, %v)", tc.name, got, ok, tc.want, tc.ok)
		}
	}
}

func TestComposeFolderName(t *testing.T) {
	cases := []struct {
		date  string
		picks []string
		want  string
	}{
		{"2026-11-27", []string{"Araf İstanbul"}, "2026-11-27 - Araf İstanbul"},
		{"2026-11-27", []string{"Araf İstanbul", "Topdeck"}, "2026-11-27 - Araf İstanbul, Topdeck"},
		{"2026-11-27", []string{"Araf İstanbul", "Araf İstanbul", "Topdeck"}, "2026-11-27 - Araf İstanbul, Topdeck"},
		{"2026-11-27", []string{"  ", "Topdeck"}, "2026-11-27 - Topdeck"},
		{"2026-11-27", nil, ""},
		{"2026-11-27", []string{}, ""},
	}
	for _, tc := range cases {
		if got := ComposeFolderName(tc.date, tc.picks); got != tc.want {
			t.Errorf("ComposeFolderName(%q, %v) = %q, want %q", tc.date, tc.picks, got, tc.want)
		}
	}
}

func TestComposeFolderNameCapsLength(t *testing.T) {
	// Five long place names: the full join is far over the cap, so
	// trailing places are dropped until it fits — the earliest picks
	// (chronological order) are the ones kept.
	long := strings.Repeat("a", 60)
	got := ComposeFolderName("2026-11-27", []string{long, long + "b", long + "c", long + "d", long + "e"})
	if len(got) > MaxFolderNameBytes {
		t.Errorf("ComposeFolderName = %d bytes, want at most %d", len(got), MaxFolderNameBytes)
	}
	if !strings.HasPrefix(got, "2026-11-27 - "+long) {
		t.Errorf("ComposeFolderName = %q, want it to start with the first pick", got)
	}
	if strings.Contains(got, long+"e") {
		t.Errorf("ComposeFolderName = %q, want the last pick dropped", got)
	}
	// A short name that already fits is untouched.
	short := ComposeFolderName("2026-11-27", []string{"Araf İstanbul", "Topdeck"})
	if short != "2026-11-27 - Araf İstanbul, Topdeck" {
		t.Errorf("ComposeFolderName = %q, want the full name unchanged", short)
	}
}

func TestComposeFolderNameTruncatesSingleLongPlace(t *testing.T) {
	// One place, far too long on its own: the result is the date
	// prefix plus a truncation of the place, still within the cap.
	got := ComposeFolderName("2026-11-27", []string{strings.Repeat("a", 400)})
	if len(got) > MaxFolderNameBytes {
		t.Errorf("ComposeFolderName = %d bytes, want at most %d", len(got), MaxFolderNameBytes)
	}
	if !strings.HasPrefix(got, "2026-11-27 - a") {
		t.Errorf("ComposeFolderName = %q, want the date prefix and the start of the place", got)
	}
	// Multi-byte runes must never be split: "İstanbul" is 9 bytes in
	// UTF-8, so a cut that lands inside one must back off to a valid
	// boundary.
	got = ComposeFolderName("2026-11-27", []string{strings.Repeat("İstanbul", 40)})
	if len(got) > MaxFolderNameBytes {
		t.Errorf("ComposeFolderName = %d bytes, want at most %d", len(got), MaxFolderNameBytes)
	}
	if !utf8.ValidString(got) {
		t.Errorf("ComposeFolderName = %q, want valid UTF-8", got)
	}
	if !strings.HasPrefix(got, "2026-11-27 - İstanbul") {
		t.Errorf("ComposeFolderName = %q, want the date prefix and the start of the place", got)
	}
}
