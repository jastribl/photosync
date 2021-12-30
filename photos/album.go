package photos

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type RequestAlbum struct {
	Title string `json:"title"`
}

type AlbumCreateRequest struct {
	Album RequestAlbum `json:"album"`
}

func (m *Client) CreateAlbum(title string) (*Album, error) {
	createRequest := AlbumCreateRequest{
		Album: RequestAlbum{
			Title: title,
		},
	}
	response := &Album{}
	err := m.postJson(
		"https://photoslibrary.googleapis.com/v1/albums",
		createRequest,
		response,
	)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (m *Client) getAlbums(pageToken string) (*Albums, error) {
	pageTokenPart := ""
	if pageToken != "" {
		pageTokenPart = fmt.Sprintf("&pageToken=%s", pageToken)
	}
	resp, err := m.httpClient.Get(fmt.Sprintf(
		"https://photoslibrary.googleapis.com/v1/albums?pageSize=50%s",
		pageTokenPart,
	))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	d := &Albums{}
	err = json.NewDecoder(resp.Body).Decode(d)

	return d, err
}

func (m *Client) GetAllAlbums() ([]*Album, error) {
	var allAlbums []*Album
	lastPageToken := ""
	for {
		albums, err := m.getAlbums(lastPageToken)
		if err != nil {
			return nil, err
		}
		allAlbums = append(allAlbums, albums.Albums...)
		lastPageToken = albums.NextPageToken
		if lastPageToken == "" {
			break
		}
	}

	return allAlbums, nil
}

func (m *Client) GetAlbumWithTitle(title string) (*Album, error) {
	albums, err := m.GetAllAlbums()
	if err != nil {
		return nil, err
	}

	for _, album := range albums {
		if album.Title == title {
			return album, nil
		}
	}

	return nil, nil
}

type AlbumPosition struct {
	Position                 string `json:"position,omitempty"`
	RelativeEnrichmentItemId string `json:"relativeEnrichmentItemId,omitempty"`
	RelativeMediaItemId      string `json:"relativeMediaItemId,omitempty"`
}

type LatLng struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

type Location struct {
	Latlng       *LatLng `json:"latlng,omitempty"`
	LocationName string  `json:"locationName,omitempty"`
}

type LocationEnrichment struct {
	Location *Location `json:"location,omitempty"`
}

type MapEnrichment struct {
	Destination *Location `json:"destination,omitempty"`
	Origin      *Location `json:"origin,omitempty"`
}

type TextEnrichment struct {
	Text string `json:"text,omitempty"`
}

type NewEnrichmentItem struct {
	LocationEnrichment *LocationEnrichment `json:"locationEnrichment,omitempty"`
	MapEnrichment      *MapEnrichment      `json:"mapEnrichment,omitempty"`
	TextEnrichment     *TextEnrichment     `json:"textEnrichment,omitempty"`
}

type AddEnrichmentToAlbumRequest struct {
	AlbumPosition     *AlbumPosition     `json:"albumPosition,omitempty"`
	NewEnrichmentItem *NewEnrichmentItem `json:"newEnrichmentItem,omitempty"`
}

func (m *Client) AddTextEnrichmentToAlbum(
	albumID string,
	afterMediaItem *MediaItem,
	labelText string,
) (*AddEnrichmentResponse, error) {
	var position string
	if afterMediaItem == nil {
		position = "FIRST_IN_ALBUM"
	} else {
		position = "AFTER_MEDIA_ITEM"
	}

	for sleepSeconds := 1; ; sleepSeconds *= 2 {
		if sleepSeconds > 10 {
			sleepSeconds = 10
		}
		request := AddEnrichmentToAlbumRequest{
			NewEnrichmentItem: &NewEnrichmentItem{
				TextEnrichment: &TextEnrichment{
					Text: labelText,
				},
			},
			AlbumPosition: &AlbumPosition{
				Position: position,
			},
		}
		if afterMediaItem != nil {
			request.AlbumPosition.RelativeMediaItemId = afterMediaItem.ID
		}
		response := &AddEnrichmentResponse{}
		err := m.postJson(
			fmt.Sprintf("https://photoslibrary.googleapis.com/v1/albums/%s:addEnrichment", albumID),
			request,
			response,
		)
		if err != nil {
			return nil, err
		}
		if response.Error == nil {
			return response, nil
		}
		if response.Error.Status == "RESOURCE_EXHAUSTED" {
			// this means we need to retry after some time
			log.Println("Hit API Quota Limit, retrying after a short sleep...")
			time.Sleep(time.Duration(sleepSeconds) * time.Second)
			continue
		}

		return nil, fmt.Errorf(
			"got an error other than resource exhausted: %s - %s",
			response.Error.Message,
			response.Error.Status,
		)
	}
}

type EnrichmentItem struct {
	ID string `json:"id"`
}

type ErrorDetails struct {
	Type     string            `json:"type"`
	Metadata map[string]string `json:"metadata"`
	Reason   string            `json:"reason"`
}

type ErrorResponse struct {
	Code    int            `json:"code"`
	Details []ErrorDetails `json:"details"`
	Message string         `json:"message"`
	Status  string         `json:"status"`
}

type AddEnrichmentResponse struct {
	EnrichmentItem *EnrichmentItem `json:"enrichmentItem"`
	Error          *ErrorResponse  `json:"error"`
}
