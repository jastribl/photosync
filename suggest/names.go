package suggest

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// MaxFolderNameBytes is the longest a folder name may be, in bytes.
// macOS (APFS, like HFS+ before it) caps a single path component at
// NAME_MAX = 255. APFS counts the limit in UTF-16 code units in some
// accounts and UTF-8 bytes in others; a 255-byte cap is safe under
// either reading, because a string's UTF-8 byte length is always >=
// its UTF-16 code-unit count. The rename would fail with
// ENAMETOOLONG past the cap, so ComposeFolderName never returns a
// longer name.
const MaxFolderNameBytes = 255

// folderNameRe matches a date folder name: a YYYY-MM-DD prefix, with an
// optional " - <description>" (or " <description>") after it. The
// description part is what this tool proposes and replaces; the date
// prefix is the day and never changes.
var folderNameRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})(?:\s+.*)?$`)

// ParseFolderDate returns the YYYY-MM-DD prefix of a folder name, and
// whether the name has a valid one. "2026-11-27", "2026-11-27 - Araf
// İstanbul" and "2026-11-27 Araf İstanbul" all parse to "2026-11-27";
// anything else does not parse.
func ParseFolderDate(folderName string) (string, bool) {
	m := folderNameRe.FindStringSubmatch(folderName)
	if m == nil {
		return "", false
	}
	if _, err := time.Parse("2006-01-02", m[1]); err != nil {
		return "", false
	}
	return m[1], true
}

// ComposeFolderName builds the proposed folder name from the day's date
// prefix and the picked place names, in cluster (chronological) order:
// "YYYY-MM-DD - place1, place2, place3". Duplicate names are kept
// once, at their first position: two clusters at the same restaurant
// still name it once. It returns "" when there are no picks, which
// means "do not rename". The result is capped at MaxFolderNameBytes:
// trailing places are dropped one at a time until it fits, and a
// single place that is still too long is truncated at a UTF-8 rune
// boundary, so the rename can never fail on length.
func ComposeFolderName(datePrefix string, picks []string) string {
	seen := map[string]bool{}
	names := []string{}
	for _, pick := range picks {
		name := strings.TrimSpace(pick)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	if len(names) == 0 {
		return ""
	}
	for len(names) > 1 {
		candidate := datePrefix + " - " + strings.Join(names, ", ")
		if len(candidate) <= MaxFolderNameBytes {
			return candidate
		}
		names = names[:len(names)-1]
	}
	candidate := datePrefix + " - " + names[0]
	if len(candidate) <= MaxFolderNameBytes {
		return candidate
	}
	return truncateToBytes(candidate, MaxFolderNameBytes)
}

// truncateToBytes shortens s to at most maxBytes bytes without splitting
// a multi-byte UTF-8 rune, and trims any trailing separator left by the
// cut. It assumes len(s) > maxBytes.
func truncateToBytes(s string, maxBytes int) string {
	s = s[:maxBytes]
	for !utf8.ValidString(s) {
		s = s[:len(s)-1]
	}
	return strings.TrimRight(s, " ,")
}
