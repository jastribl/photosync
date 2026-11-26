package suggest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Place is one candidate venue near a cluster of photos, as returned by
// Google Places Nearby Search.
type Place struct {
	ID          string
	Name        string
	Address     string
	PrimaryType string
	Types       []string
	Lat         float64
	Long        float64
}

// PlacesClient is the one call suggestnames needs: the places near a
// point. The interface keeps the real API and the offline stub
// interchangeable, and lets tests drive the flow without a network.
type PlacesClient interface {
	SearchNearby(ctx context.Context, lat, lng, radiusM float64) ([]Place, error)
}

// nearbySearchTypes is what a day's stops are: somewhere to eat, or
// somewhere worth visiting. They go in a single call, so one call per
// cluster covers every kind of stop.
var nearbySearchTypes = []string{
	"restaurant",
	"cafe",
	"tourist_attraction",
	"park",
	"museum",
	"hiking_area",
}

// proFieldMask is the Places (New) field mask for the Pro pricing tier:
// name, address, location, and types. Rating, opening hours, reviews,
// and photos would re-price the call to Enterprise, so they are not
// requested — the shortlist does not need them.
const proFieldMask = "places.id,places.displayName,places.formattedAddress,places.location,places.types,places.primaryType"

const (
	placesBaseURL  = "https://places.googleapis.com"
	maxResults     = 5
	searchEndpoint = "/v1/places:searchNearby"
)

// restClient talks to the Places API (New) over plain REST: one POST per
// call, API key in a header, no SDK and no new dependencies.
type restClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewRESTClient returns a PlacesClient backed by the live Places API
// (New). apiKey is a Google Cloud API key with the Places API (New)
// enabled. Billing is not required for the free-tier calls this tool
// makes (verified live Nov 2026 without billing); it would only matter
// past the free caps.
//
// The transport forces HTTP/1.1: Go's HTTP/2 over the egress proxy on
// this dev machine hangs (verified Nov 2026 — the same call over
// HTTP/1.1 returns in under a second), and the Places API is
// HTTP/1.1-compatible, so this costs nothing anywhere.
func NewRESTClient(apiKey string) PlacesClient {
	return &restClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy:           http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{NextProtos: []string{"http/1.1"}},
			},
		},
		baseURL: placesBaseURL,
	}
}

type nearbySearchRequest struct {
	IncludedTypes       []string `json:"includedTypes"`
	MaxResultCount      int      `json:"maxResultCount"`
	LocationRestriction struct {
		Circle struct {
			Center struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"center"`
			Radius float64 `json:"radius"`
		} `json:"circle"`
	} `json:"locationRestriction"`
	RankPreference string `json:"rankPreference"`
}

type nearbySearchResponse struct {
	Places []struct {
		ID          string `json:"id"`
		DisplayName struct {
			Text string `json:"text"`
		} `json:"displayName"`
		FormattedAddress string `json:"formattedAddress"`
		Location         struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location"`
		Types       []string `json:"types"`
		PrimaryType string   `json:"primaryType"`
	} `json:"places"`
}

func (c *restClient) SearchNearby(ctx context.Context, lat, lng, radiusM float64) ([]Place, error) {
	var req nearbySearchRequest
	req.IncludedTypes = nearbySearchTypes
	req.MaxResultCount = maxResults
	req.LocationRestriction.Circle.Center.Latitude = lat
	req.LocationRestriction.Circle.Center.Longitude = lng
	req.LocationRestriction.Circle.Radius = radiusM
	req.RankPreference = "DISTANCE"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+searchEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Goog-Api-Key", c.apiKey)
	httpReq.Header.Set("X-Goog-FieldMask", proFieldMask)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("places search nearby: HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	var parsed nearbySearchResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("places search nearby: cannot parse response: %w", err)
	}
	places := make([]Place, 0, len(parsed.Places))
	for _, p := range parsed.Places {
		places = append(places, Place{
			ID:          p.ID,
			Name:        p.DisplayName.Text,
			Address:     p.FormattedAddress,
			PrimaryType: p.PrimaryType,
			Types:       p.Types,
			Lat:         p.Location.Latitude,
			Long:        p.Location.Longitude,
		})
	}
	return places, nil
}

// stubClient returns canned places near the requested point, with no
// network. It exists so the tool and its tests can run end to end
// offline (--stub); the places are fixed, only their positions move with
// the requested centroid.
type stubClient struct{}

// NewStubClient returns a PlacesClient that answers every search with
// three canned places near the requested point.
func NewStubClient() PlacesClient { return &stubClient{} }

func (s *stubClient) SearchNearby(ctx context.Context, lat, lng, radiusM float64) ([]Place, error) {
	return []Place{
		{ID: "stub-1", Name: "Stub Restaurant", Address: "1 Stub Street", PrimaryType: "restaurant", Types: []string{"restaurant", "food"}, Lat: lat + 0.0004, Long: lng},
		{ID: "stub-2", Name: "Stub Cafe", Address: "2 Stub Street", PrimaryType: "cafe", Types: []string{"cafe", "food"}, Lat: lat, Long: lng + 0.0011},
		{ID: "stub-3", Name: "Stub Park", Address: "3 Stub Street", PrimaryType: "park", Types: []string{"park", "tourist_attraction"}, Lat: lat - 0.002, Long: lng},
	}, nil
}
