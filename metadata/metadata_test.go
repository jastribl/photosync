package metadata

import (
	"reflect"
	"testing"
	"time"

	"github.com/jastribl/photosync/testutil"
)

func TestParseMdlsOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   map[string]string
	}{
		{
			name: "parses attributes and skips nulls and junk lines",
			output: "kMDItemLatitude = 37.7749\n" +
				"kMDItemLongitude = -122.4194\n" +
				"kMDItemContentCreationDate = 2021-06-01 10:00:00 +0000\n" +
				"kMDItemSomething = (null)\n" +
				"not a key value line\n",
			want: map[string]string{
				"kMDItemLatitude":            "37.7749",
				"kMDItemLongitude":           "-122.4194",
				"kMDItemContentCreationDate": "2021-06-01 10:00:00 +0000",
			},
		},
		{
			name:   "empty output gives empty map",
			output: "",
			want:   map[string]string{},
		},
		{
			name:   "values with equals signs keep the rest",
			output: "kMDItemFoo = a=b\n",
			want:   map[string]string{"kMDItemFoo": "a=b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseMdlsOutput(tt.output); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseMdlsOutput() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestExtractReadsGPSAndDateTime(t *testing.T) {
	dir := t.TempDir()
	taken := time.Date(2026, 11, 27, 19, 30, 0, 0, time.UTC)
	path := testutil.WritePhoto(t, dir, "a.jpg", 41.0082, 28.9784, taken)

	md := Extract(path, "a.jpg")
	if !md.HasGPS {
		t.Fatal("expected GPS data")
	}
	if diff := md.Lat - 41.0082; diff > 0.0001 || diff < -0.0001 {
		t.Errorf("Lat = %v, want ~41.0082", md.Lat)
	}
	if diff := md.Long - 28.9784; diff > 0.0001 || diff < -0.0001 {
		t.Errorf("Long = %v, want ~28.9784", md.Long)
	}
	if !md.HasDateTime {
		t.Fatal("expected a taken time")
	}
	if got := md.DateTime.Format("2006-01-02 15:04:05"); got != "2026-11-27 19:30:00" {
		t.Errorf("DateTime = %v, want 2026-11-27 19:30:00", got)
	}
}

func TestExtractWithoutGPS(t *testing.T) {
	dir := t.TempDir()
	taken := time.Date(2026, 11, 27, 19, 30, 0, 0, time.UTC)
	path := testutil.WritePhoto(t, dir, "a.jpg", 0, 0, taken)

	md := Extract(path, "a.jpg")
	if md.HasGPS {
		t.Errorf("expected no GPS data, got lat=%v long=%v", md.Lat, md.Long)
	}
	if !md.HasDateTime {
		t.Error("expected a taken time")
	}
}

func TestExtractPlainJPEG(t *testing.T) {
	dir := t.TempDir()
	path := testutil.WritePhoto(t, dir, "a.jpg", 0, 0, time.Time{})

	md := Extract(path, "a.jpg")
	if md.HasGPS || md.HasDateTime {
		t.Errorf("expected empty metadata, got %+v", md)
	}
}

func TestExtractMissingFile(t *testing.T) {
	md := Extract("/nonexistent/a.jpg", "a.jpg")
	if md.HasGPS || md.HasDateTime {
		t.Errorf("expected empty metadata for a missing file, got %+v", md)
	}
}
