package suggest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const cannedPlacesJSON = `{
  "places": [
    {
      "id": "ChIJaraf123",
      "displayName": {"text": "Araf İstanbul", "languageCode": "tr"},
      "formattedAddress": "Balo Sk. No:32, Kadıköy",
      "location": {"latitude": 41.0082, "longitude": 28.9784},
      "types": ["restaurant", "food", "establishment"],
      "primaryType": "restaurant"
    },
    {
      "id": "ChIJcafe456",
      "displayName": {"text": "Coffee Art", "languageCode": "en"},
      "formattedAddress": "Moda Cd. No:10, Kadıköy",
      "location": {"latitude": 41.0090, "longitude": 28.9790},
      "types": ["cafe", "food", "establishment"],
      "primaryType": "cafe"
    }
  ]
}`

func TestRESTClientRequestShapeAndParsing(t *testing.T) {
	var gotMethod, gotPath, gotKey, gotMask string
	var gotBody nearbySearchRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotKey = r.Header.Get("X-Goog-Api-Key")
		gotMask = r.Header.Get("X-Goog-FieldMask")
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &gotBody); err != nil {
			t.Errorf("request body is not valid JSON: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cannedPlacesJSON))
	}))
	defer server.Close()

	client := &restClient{apiKey: "test-key", httpClient: server.Client(), baseURL: server.URL}
	places, err := client.SearchNearby(context.Background(), 41.0082, 28.9784, 200)
	if err != nil {
		t.Fatalf("SearchNearby: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if gotPath != "/v1/places:searchNearby" {
		t.Errorf("path = %s, want /v1/places:searchNearby", gotPath)
	}
	if gotKey != "test-key" {
		t.Errorf("X-Goog-Api-Key = %q, want test-key", gotKey)
	}
	if gotMask != proFieldMask {
		t.Errorf("X-Goog-FieldMask = %q, want %q", gotMask, proFieldMask)
	}
	if strings.Contains(gotMask, "rating") || strings.Contains(gotMask, "regularOpeningHours") {
		t.Errorf("field mask must stay on the Pro tier, got %q", gotMask)
	}
	if len(gotBody.IncludedTypes) != len(nearbySearchTypes) {
		t.Errorf("includedTypes = %v, want %v", gotBody.IncludedTypes, nearbySearchTypes)
	}
	if gotBody.MaxResultCount != 5 {
		t.Errorf("maxResultCount = %d, want 5", gotBody.MaxResultCount)
	}
	if gotBody.RankPreference != "DISTANCE" {
		t.Errorf("rankPreference = %q, want DISTANCE", gotBody.RankPreference)
	}
	if gotBody.LocationRestriction.Circle.Center.Latitude != 41.0082 ||
		gotBody.LocationRestriction.Circle.Center.Longitude != 28.9784 ||
		gotBody.LocationRestriction.Circle.Radius != 200 {
		t.Errorf("circle = %+v, want center (41.0082, 28.9784) radius 200", gotBody.LocationRestriction.Circle)
	}

	if len(places) != 2 {
		t.Fatalf("expected 2 places, got %d", len(places))
	}
	if places[0].Name != "Araf İstanbul" || places[0].PrimaryType != "restaurant" || places[0].ID != "ChIJaraf123" {
		t.Errorf("first place = %+v", places[0])
	}
	if places[0].Address != "Balo Sk. No:32, Kadıköy" {
		t.Errorf("first place address = %q", places[0].Address)
	}
	if places[0].Lat != 41.0082 || places[0].Long != 28.9784 {
		t.Errorf("first place location = (%v, %v)", places[0].Lat, places[0].Long)
	}
	if places[1].Name != "Coffee Art" || places[1].PrimaryType != "cafe" {
		t.Errorf("second place = %+v", places[1])
	}
}

func TestRESTClientNonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"code": 400, "message": "API key not valid."}}`))
	}))
	defer server.Close()

	client := &restClient{apiKey: "bad-key", httpClient: server.Client(), baseURL: server.URL}
	_, err := client.SearchNearby(context.Background(), 41.0, 28.9, 200)
	if err == nil {
		t.Fatal("expected an error for HTTP 400")
	}
	if !strings.Contains(err.Error(), "400") || !strings.Contains(err.Error(), "API key not valid") {
		t.Errorf("error = %v, want it to carry the status and body", err)
	}
}

func TestRESTClientForcesHTTP11(t *testing.T) {
	// The egress proxy on the dev machine hangs on Go's HTTP/2, so the
	// client must negotiate HTTP/1.1 only. See NewRESTClient.
	client, ok := NewRESTClient("key").(*restClient)
	if !ok {
		t.Fatal("NewRESTClient did not return a *restClient")
	}
	tr, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("restClient transport is not an *http.Transport")
	}
	if tr.TLSClientConfig == nil || len(tr.TLSClientConfig.NextProtos) != 1 || tr.TLSClientConfig.NextProtos[0] != "http/1.1" {
		t.Errorf("transport NextProtos = %v, want [http/1.1]", tr.TLSClientConfig.NextProtos)
	}
}

func TestStubClient(t *testing.T) {
	places, err := NewStubClient().SearchNearby(context.Background(), 41.0082, 28.9784, 200)
	if err != nil {
		t.Fatalf("stub SearchNearby: %v", err)
	}
	if len(places) != 3 {
		t.Fatalf("expected 3 stub places, got %d", len(places))
	}
	if places[0].Name != "Stub Restaurant" || places[1].Name != "Stub Cafe" || places[2].Name != "Stub Park" {
		t.Errorf("stub places = %v", places)
	}
	for _, p := range places {
		if d := haversineM(41.0082, 28.9784, p.Lat, p.Long); d > 500 {
			t.Errorf("stub place %q is %v m from the centroid, want it nearby", p.Name, d)
		}
	}
}
