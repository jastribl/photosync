package main

import (
	"reflect"
	"testing"
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
