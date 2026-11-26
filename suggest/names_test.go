package suggest

import "testing"

func TestParseFolderDate(t *testing.T) {
	cases := []struct {
		name string
		want string
		ok   bool
	}{
		{"2026-11-27", "2026-11-27", true},
		{"2026-11-27 Araf İstanbul", "2026-11-27", true},
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
		{"2026-11-27", []string{"Araf İstanbul"}, "2026-11-27 Araf İstanbul"},
		{"2026-11-27", []string{"Araf İstanbul", "Topdeck"}, "2026-11-27 Araf İstanbul, Topdeck"},
		{"2026-11-27", []string{"Araf İstanbul", "Araf İstanbul", "Topdeck"}, "2026-11-27 Araf İstanbul, Topdeck"},
		{"2026-11-27", []string{"  ", "Topdeck"}, "2026-11-27 Topdeck"},
		{"2026-11-27", nil, ""},
		{"2026-11-27", []string{}, ""},
	}
	for _, tc := range cases {
		if got := ComposeFolderName(tc.date, tc.picks); got != tc.want {
			t.Errorf("ComposeFolderName(%q, %v) = %q, want %q", tc.date, tc.picks, got, tc.want)
		}
	}
}
