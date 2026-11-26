// Package metadata extracts the GPS coordinates and taken time of a photo
// file. It is shared by sortlocal (which sorts photos into date folders)
// and suggestnames (which clusters a day's photos by location and time).
//
// The primary path reads EXIF via the imagemeta library, which handles
// JPEG, HEIC, and other formats. On macOS there is an mdls fallback for
// files imagemeta cannot read (e.g. HEICs whose EXIF it cannot parse);
// on other systems the fallback simply finds no mdls binary and returns
// empty metadata.
package metadata

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/evanoberholster/imagemeta"
)

// Metadata is the GPS position and taken time of one photo file, as far
// as they could be determined. HasGPS / HasDateTime say whether each
// field is meaningful.
type Metadata struct {
	Lat, Long   float64
	HasGPS      bool
	DateTime    time.Time
	HasDateTime bool
}

// Extract reads the metadata of the file at filePath. filename is only
// used in log lines. A file that cannot be read, or whose metadata cannot
// be parsed, yields an empty Metadata rather than an error: callers treat
// missing GPS / time as "unknown" and fall back.
func Extract(filePath, filename string) Metadata {
	f, err := os.Open(filePath)
	if err != nil {
		return extractMetadataMdls(filePath)
	}
	defer f.Close()

	e, err := imagemeta.Decode(f)
	if err != nil {
		log.Printf("imagemeta unsupported for '%s': %s, trying mdls fallback\n", filename, err)
		return extractMetadataMdls(filePath)
	}

	metadata := Metadata{}
	metadata.Lat, metadata.Long = e.GPS.Latitude(), e.GPS.Longitude()
	metadata.HasGPS = metadata.Lat != 0 || metadata.Long != 0
	dt := e.SelectedDate()
	if !dt.IsZero() {
		metadata.DateTime = dt
		metadata.HasDateTime = true
	}
	return metadata
}

func extractMetadataMdls(filePath string) Metadata {
	metadata := Metadata{}

	out, err := exec.Command("mdls",
		"-name", "kMDItemLatitude",
		"-name", "kMDItemLongitude",
		"-name", "kMDItemContentCreationDate",
		filePath,
	).Output()
	if err != nil {
		return metadata
	}

	attrs := parseMdlsOutput(string(out))

	latStr, hasLat := attrs["kMDItemLatitude"]
	longStr, hasLong := attrs["kMDItemLongitude"]
	if hasLat && hasLong {
		lat, latErr := strconv.ParseFloat(latStr, 64)
		long, longErr := strconv.ParseFloat(longStr, 64)
		if latErr == nil && longErr == nil {
			metadata.Lat, metadata.Long, metadata.HasGPS = lat, long, true
		}
	}

	if dtStr, ok := attrs["kMDItemContentCreationDate"]; ok {
		dt, err := time.Parse("2006-01-02 15:04:05 +0000", dtStr)
		if err == nil {
			metadata.DateTime, metadata.HasDateTime = dt, true
		}
	}

	return metadata
}

func parseMdlsOutput(output string) map[string]string {
	attrs := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if value == "(null)" {
			continue
		}
		attrs[key] = value
	}
	return attrs
}
