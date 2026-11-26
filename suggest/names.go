package suggest

import (
	"regexp"
	"strings"
	"time"
)

// folderNameRe matches a date folder name: a YYYY-MM-DD prefix, with an
// optional " <description>" after it. The description part is what this
// tool proposes and replaces; the date prefix is the day and never
// changes.
var folderNameRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})(?:\s+.*)?$`)

// ParseFolderDate returns the YYYY-MM-DD prefix of a folder name, and
// whether the name has a valid one. "2026-11-27" and "2026-11-27 Araf
// İstanbul" both parse to "2026-11-27"; anything else does not parse.
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
// prefix and the picked place names, in cluster (chronological) order.
// Duplicate names are kept once, at their first position: two clusters
// at the same restaurant still name it once. It returns "" when there
// are no picks, which means "do not rename".
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
	return datePrefix + " " + strings.Join(names, ", ")
}
