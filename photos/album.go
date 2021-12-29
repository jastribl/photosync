package photos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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
	var album Album
	err := m.postJson("https://photoslibrary.googleapis.com/v1/albums", createRequest, &album)
	if err != nil {
		return nil, err
	}
	return &album, nil
}

// GetAlbums gets albums
func (m *Client) GetAlbums(pageToken string) (*Albums, error) {
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

// GetAllAlbums gets all albums
func (m *Client) GetAllAlbums() ([]*Album, error) {
	var allAlbums []*Album
	lastPageToken := ""
	for {
		albums, err := m.GetAlbums(lastPageToken)
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

type MediaItemsListPostStruct struct {
	MediaItemIDs []string `json:"mediaItemIds"`
}

func (m *Client) AddMediaItemsToAlbum(albumID string, mediaItems []*MediaItem) error {
	mediaItemIDs := []string{}
	for _, item := range mediaItems {
		mediaItemIDs = append(mediaItemIDs, item.ID)
	}
	postStruct := MediaItemsListPostStruct{
		MediaItemIDs: mediaItemIDs,
	}
	jsonStr, err := json.Marshal(postStruct)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(
		"https://photoslibrary.googleapis.com/v1/albums/%s:batchAddMediaItems",
		albumID,
	)
	resp, err := m.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	var target map[string]interface{}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&target)
	if err != nil {
		return err
	}
	return nil
}

// GetAlbumWithTitleContains returns the album with the title containing
func (m *Client) GetAlbumWithTitleContains(title string) (*Album, error) {
	albums, err := m.GetAllAlbums()
	if err != nil {
		return nil, err
	}

	for _, album := range albums {
		if strings.Contains(album.Title, title) {
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
	albumID, afterMediaId, text string,
) (*AddEnrichmentResponse, error) {
	var position string
	if afterMediaId == "" {
		position = "FIRST_IN_ALBUM"
	} else {
		position = "AFTER_MEDIA_ITEM"
	}
	jsonStr, err := json.Marshal(&AddEnrichmentToAlbumRequest{
		NewEnrichmentItem: &NewEnrichmentItem{
			TextEnrichment: &TextEnrichment{
				Text: text,
			},
		},
		AlbumPosition: &AlbumPosition{
			Position:            position,
			RelativeMediaItemId: afterMediaId,
		},
	})
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(
		"https://photoslibrary.googleapis.com/v1/albums/%s:addEnrichment",
		albumID,
	)
	resp, err := m.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	response := &AddEnrichmentResponse{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.Fatal(err)
	}

	return response, nil

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
