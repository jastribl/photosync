// Package testutil holds helpers shared by the repo's tests. It is
// test-only infrastructure: WritePhoto builds small valid JPEGs with
// EXIF GPS and DateTimeOriginal baked in, so tests can exercise the
// metadata and clustering code without real photos.
package testutil

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// WritePhoto writes a small valid JPEG at dir/name whose EXIF carries the
// given GPS position (decimal degrees) and taken time, and returns the
// file path. Pass lat == 0 && lng == 0 for "no GPS", and a zero taken
// time for "no taken time".
func WritePhoto(t *testing.T, dir, name string, lat, lng float64, taken time.Time) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 30, B: 30, A: 255})
		}
	}
	var jpg bytes.Buffer
	if err := jpeg.Encode(&jpg, img, nil); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}
	data := jpg.Bytes()
	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		t.Fatalf("encoded JPEG does not start with SOI")
	}
	hasGPS := lat != 0 || lng != 0
	hasTaken := !taken.IsZero()
	if hasGPS || hasTaken {
		exif := buildEXIF(lat, lng, hasGPS, taken, hasTaken)
		seg := []byte{0xFF, 0xE1}
		segLen := 2 + 6 + len(exif)
		seg = append(seg, byte(segLen>>8), byte(segLen))
		seg = append(seg, []byte("Exif\x00\x00")...)
		seg = append(seg, exif...)
		out := append([]byte{}, data[:2]...)
		out = append(out, seg...)
		out = append(out, data[2:]...)
		data = out
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("os.WriteFile(%s): %v", path, err)
	}
	return path
}

// buildEXIF assembles a little-endian TIFF block carrying DateTimeOriginal
// and/or GPS coordinates, as found in a JPEG APP1 Exif segment.
func buildEXIF(lat, lng float64, hasGPS bool, taken time.Time, hasTaken bool) []byte {
	ifd0Entries := 1 // ExifOffset
	if hasGPS {
		ifd0Entries = 2 // + GPSInfo
	}
	ifd0Len := 2 + 12*ifd0Entries + 4
	exifIFDOff := 8 + ifd0Len
	exifLen := 0
	if hasTaken {
		exifLen = 2 + 12 + 4
	}
	gpsIFDOff := exifIFDOff + exifLen
	gpsLen := 0
	if hasGPS {
		gpsLen = 2 + 12*4 + 4
	}
	dataOff := gpsIFDOff + gpsLen

	dtOff := 0
	latRatsOff := 0
	lngRatsOff := 0
	if hasTaken {
		dtOff = dataOff
	}
	if hasGPS {
		latRatsOff = dataOff
		if hasTaken {
			latRatsOff += 20
		}
		lngRatsOff = latRatsOff + 24
	}

	var b bytes.Buffer
	// TIFF header: little-endian, magic 42, IFD0 at offset 8.
	b.Write([]byte("II"))
	binary.Write(&b, binary.LittleEndian, uint16(42))
	binary.Write(&b, binary.LittleEndian, uint32(8))

	// IFD0.
	binary.Write(&b, binary.LittleEndian, uint16(ifd0Entries))
	writeEntry(&b, 0x8769, 4, 1, uint32(exifIFDOff)) // ExifOffset
	if hasGPS {
		writeEntry(&b, 0x8825, 4, 1, uint32(gpsIFDOff)) // GPSInfo
	}
	binary.Write(&b, binary.LittleEndian, uint32(0))

	// Exif IFD: DateTimeOriginal.
	if hasTaken {
		binary.Write(&b, binary.LittleEndian, uint16(1))
		dtStr := taken.Format("2006:01:02 15:04:05") + "\x00"
		writeEntry(&b, 0x9003, 2, uint32(len(dtStr)), uint32(dtOff))
		binary.Write(&b, binary.LittleEndian, uint32(0))
	}

	// GPS IFD.
	if hasGPS {
		binary.Write(&b, binary.LittleEndian, uint16(4))
		latRef := "N\x00"
		if lat < 0 {
			latRef = "S\x00"
		}
		writeEntryInline(&b, 0x0001, 2, 2, []byte(latRef))
		writeEntry(&b, 0x0002, 5, 3, uint32(latRatsOff))
		lngRef := "E\x00"
		if lng < 0 {
			lngRef = "W\x00"
		}
		writeEntryInline(&b, 0x0003, 2, 2, []byte(lngRef))
		writeEntry(&b, 0x0004, 5, 3, uint32(lngRatsOff))
		binary.Write(&b, binary.LittleEndian, uint32(0))
	}

	// Overflow data: DateTimeOriginal string, then GPS rationals.
	if hasTaken {
		b.WriteString(taken.Format("2006:01:02 15:04:05") + "\x00")
	}
	if hasGPS {
		writeRationals(&b, lat)
		writeRationals(&b, lng)
	}
	return b.Bytes()
}

// writeEntry writes one 12-byte IFD entry whose value is a 4-byte
// uint32 stored inline (for LONG counts of 1, or an offset).
func writeEntry(b *bytes.Buffer, tag, typ uint16, count, value uint32) {
	binary.Write(b, binary.LittleEndian, tag)
	binary.Write(b, binary.LittleEndian, typ)
	binary.Write(b, binary.LittleEndian, count)
	binary.Write(b, binary.LittleEndian, value)
}

// writeEntryInline writes one 12-byte IFD entry whose value is at most
// 4 bytes, stored inline and zero-padded.
func writeEntryInline(b *bytes.Buffer, tag, typ uint16, count uint32, value []byte) {
	binary.Write(b, binary.LittleEndian, tag)
	binary.Write(b, binary.LittleEndian, typ)
	binary.Write(b, binary.LittleEndian, count)
	padded := make([]byte, 4)
	copy(padded, value)
	b.Write(padded)
}

// writeRationals writes degrees/minutes/seconds as three rationals.
func writeRationals(b *bytes.Buffer, decimal float64) {
	abs := math.Abs(decimal)
	deg := uint32(abs)
	minF := (abs - float64(deg)) * 60
	min := uint32(minF)
	secF := (minF - float64(min)) * 60
	secNum := uint32(math.Round(secF * 100))
	writeRational(b, deg, 1)
	writeRational(b, min, 1)
	writeRational(b, secNum, 100)
}

func writeRational(b *bytes.Buffer, num, den uint32) {
	binary.Write(b, binary.LittleEndian, num)
	binary.Write(b, binary.LittleEndian, den)
}
