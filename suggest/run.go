package suggest

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config is everything Run needs: the root dir of date folders, the
// clustering and search knobs, and how interactive to be.
type Config struct {
	RootDir   string
	TimeGap   time.Duration
	DistanceM float64
	RadiusM   float64
	DryRun    bool
	AutoYes   bool
	Places    PlacesClient
}

// Summary counts what Run did, for the closing line. In a dry run,
// Renamed counts the folders that would have been renamed.
type Summary struct {
	Renamed int
	Skipped int
	NoGPS   int
}

// Run walks every date folder under RootDir, proposes a name for each
// from the places near its photos, and renames the folders that get at
// least one pick. It prints one line per pick and one per rename, and a
// short summary at the end. stdin/stdout are parameters (rather than
// os.Stdin/os.Stdout) so tests can drive the prompts.
func Run(cfg Config, stdin io.Reader, stdout io.Writer) (Summary, error) {
	rootDir := cfg.RootDir
	if !strings.HasSuffix(rootDir, "/") {
		rootDir += "/"
	}
	reader := bufio.NewReader(stdin)
	var summary Summary

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return summary, err
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".DS_Store" {
			continue
		}
		name := entry.Name()
		date, ok := ParseFolderDate(name)
		if !ok {
			fmt.Fprintf(stdout, "Skipping '%s': not a date folder\n", name)
			summary.Skipped++
			continue
		}
		folderPath := filepath.Join(rootDir, name)
		if renamed, err := processFolder(cfg, reader, stdout, folderPath, name, date); err != nil {
			return summary, err
		} else {
			summary.Renamed += renamed.Renamed
			summary.Skipped += renamed.Skipped
			summary.NoGPS += renamed.NoGPS
		}
	}

	if cfg.DryRun {
		fmt.Fprintf(stdout, "Dry run: nothing was renamed. Would rename %d folder(s), skipped %d, %d with no GPS\n", summary.Renamed, summary.Skipped, summary.NoGPS)
	} else {
		fmt.Fprintf(stdout, "Renamed %d folder(s), skipped %d, %d with no GPS\n", summary.Renamed, summary.Skipped, summary.NoGPS)
	}
	return summary, nil
}

func processFolder(cfg Config, reader *bufio.Reader, stdout io.Writer, folderPath, folderName, date string) (Summary, error) {
	var summary Summary

	photos, err := LoadPhotos(folderPath)
	if err != nil {
		return summary, err
	}
	if len(photos) == 0 {
		fmt.Fprintf(stdout, "Skipping '%s': no photos in folder\n", folderName)
		summary.Skipped++
		return summary, nil
	}
	InheritGPS(photos)
	if !HasAnyGPS(photos) {
		fmt.Fprintf(stdout, "Skipping '%s': no GPS data for any photo in folder\n", folderName)
		summary.NoGPS++
		return summary, nil
	}

	clusters := ClusterPhotos(photos, cfg.TimeGap, cfg.DistanceM)
	picks := []string{}
	for i, cluster := range clusters {
		pick, err := pickForCluster(cfg, reader, stdout, folderName, i+1, len(clusters), cluster)
		if err != nil {
			return summary, err
		}
		if pick != "" {
			fmt.Fprintf(stdout, "Picked '%s' for '%s'\n", pick, folderName)
			picks = append(picks, pick)
		}
	}

	proposed := ComposeFolderName(date, picks)
	if proposed == "" {
		fmt.Fprintf(stdout, "No picks for '%s': leaving unchanged\n", folderName)
		summary.Skipped++
		return summary, nil
	}
	if proposed == folderName {
		fmt.Fprintf(stdout, "Folder '%s' already has that name: leaving unchanged\n", folderName)
		summary.Skipped++
		return summary, nil
	}
	if cfg.DryRun {
		fmt.Fprintf(stdout, "Would rename '%s' to '%s'\n", folderName, proposed)
		summary.Renamed++
		return summary, nil
	}
	if !cfg.AutoYes {
		answer := prompt(reader, stdout, fmt.Sprintf("Rename '%s' to '%s'? [y/n] ", folderName, proposed))
		if !isYes(answer) {
			fmt.Fprintf(stdout, "Not renaming '%s'\n", folderName)
			summary.Skipped++
			return summary, nil
		}
	}
	if err := os.Rename(folderPath, filepath.Join(filepath.Dir(folderPath), proposed)); err != nil {
		return summary, err
	}
	fmt.Fprintf(stdout, "Renamed '%s' to '%s'\n", folderName, proposed)
	summary.Renamed++
	return summary, nil
}

// pickForCluster asks the Places API about one cluster and returns the
// picked place name, or "" if the cluster is skipped. An API error is
// not fatal: the cluster is skipped with a note and the run continues.
func pickForCluster(cfg Config, reader *bufio.Reader, stdout io.Writer, folderName string, index, total int, cluster Cluster) (string, error) {
	places, err := cfg.Places.SearchNearby(context.Background(), cluster.CenterLat, cluster.CenterLong, cfg.RadiusM)
	if err != nil {
		fmt.Fprintf(stdout, "Cluster %d/%d in '%s': places lookup failed (%v), skipping cluster\n", index, total, folderName, err)
		return "", nil
	}
	if len(places) == 0 {
		fmt.Fprintf(stdout, "Cluster %d/%d in '%s': no places found nearby, skipping cluster\n", index, total, folderName)
		return "", nil
	}
	if cfg.AutoYes {
		return places[0].Name, nil
	}
	fmt.Fprintf(stdout, "Cluster %d/%d in '%s': %d photo(s), %s-%s, near (%.5f, %.5f):\n",
		index, total, folderName, len(cluster.Photos),
		cluster.Start.Format("15:04"), cluster.End.Format("15:04"),
		cluster.CenterLat, cluster.CenterLong)
	for i, place := range places {
		dist := int(haversineM(cluster.CenterLat, cluster.CenterLong, place.Lat, place.Long) + 0.5)
		kind := place.PrimaryType
		if kind == "" && len(place.Types) > 0 {
			kind = place.Types[0]
		}
		fmt.Fprintf(stdout, "  %d. %s (%s, %d m) — %s\n", i+1, place.Name, kind, dist, place.Address)
	}
	answer := prompt(reader, stdout, "Pick [1-5], 0 to skip, or type a name: ")
	if answer == "" {
		return "", nil
	}
	if n, err := strconv.Atoi(answer); err == nil {
		if n == 0 {
			return "", nil
		}
		if n >= 1 && n <= len(places) {
			return places[n-1].Name, nil
		}
		fmt.Fprintf(stdout, "Out of range, skipping cluster\n")
		return "", nil
	}
	// A bare y/n here is almost certainly meant for the rename
	// confirmation below, not a place name — never accept it as one.
	// Otherwise a user answering "y" out of habit would name the
	// cluster "y".
	switch strings.ToLower(answer) {
	case "y", "yes", "n", "no":
		fmt.Fprintf(stdout, "Not a place name — skipping cluster (use the rename prompt to confirm)\n")
		return "", nil
	}
	return answer, nil
}

// prompt prints a prompt and reads one line of input, trimmed. On EOF
// or a read error it returns "".
func prompt(reader *bufio.Reader, stdout io.Writer, text string) string {
	fmt.Fprint(stdout, text)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return ""
	}
	return strings.TrimSpace(line)
}

func isYes(answer string) bool {
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}
