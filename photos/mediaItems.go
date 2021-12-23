package photos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jastribl/photosync/files"
)

const allMediaItemsCacheFile = "cache/allMediaItems.json"

// CacheAllMediaItem
func (m *Client) CacheAllMediaItems() ([]*MediaItem, error) {
	allMediaItems, err := m.GetAllMediaItems(false)
	if err != nil {
		return nil, err
	}

	bytes, _ := json.MarshalIndent(allMediaItems, "", " ")

	file, err := os.Create(allMediaItemsCacheFile)
	if err != nil {
		return nil, err
	}
	file.Close()
	err = ioutil.WriteFile(allMediaItemsCacheFile, bytes, 0644)

	return allMediaItems, err
}

func (m *Client) getMediaItemsFromCache() ([]*MediaItem, error) {
	var allMediaItems []*MediaItem

	// todo: make this generic for other calls?
	bytes, err := ioutil.ReadFile(allMediaItemsCacheFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(bytes), &allMediaItems)
	if err != nil {
		return nil, err
	}
	log.Printf("Using cache to get %d media items\n", len(allMediaItems))
	return allMediaItems, err
}

// GetAllMediaItems gets all media items
func (m *Client) GetAllMediaItems(useCache bool) ([]*MediaItem, error) {
	if useCache && files.FileExists(allMediaItemsCacheFile) {
		return m.getMediaItemsFromCache()
	}

	var allMediaItems []*MediaItem
	lastPageToken := ""
	for {
		mediaItems, err := m.getMediaItems(lastPageToken)
		if err != nil {
			return nil, err
		}
		allMediaItems = append(allMediaItems, mediaItems.MediaItems...)
		lastPageToken = mediaItems.NextPageToken
		if lastPageToken == "" {
			break
		}
	}

	return allMediaItems, nil
}

func (m *Client) GetAllMediaItemsForAlbum(album *Album) ([]*MediaItem, error) {
	var allMediaItems []*MediaItem
	lastPageToken := ""
	for {
		mediaItems, err := m.searchMediaItems(album.ID, lastPageToken)
		if err != nil {
			return nil, err
		}
		allMediaItems = append(allMediaItems, mediaItems.MediaItems...)
		lastPageToken = mediaItems.NextPageToken
		if lastPageToken == "" {
			break
		}
	}

	return allMediaItems, nil
}

// GetMediaItems gets media items
func (m *Client) getMediaItems(pageToken string) (*MediaItems, error) {
	otherParams := ""
	if pageToken != "" {
		otherParams += fmt.Sprintf("&pageToken=%s", pageToken)
	}
	resp, err := m.httpClient.Get(fmt.Sprintf(
		"https://photoslibrary.googleapis.com/v1/mediaItems?pageSize=100%s",
		otherParams,
	))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	d := &MediaItems{}
	err = json.NewDecoder(resp.Body).Decode(d)

	return d, err
}

type SearchRequest struct {
	PageSize  int    `json:"pageSize"`
	PageToken string `json:"pageToken"`
	AlbumId   string `json:"albumId"`
}

// SeatchMediaItems gets media items
func (m *Client) searchMediaItems(albumID, pageToken string) (*MediaItems, error) {
	searchRequest := SearchRequest{
		PageSize:  100,
		PageToken: pageToken,
		AlbumId:   albumID,
	}
	jsonStr, err := json.Marshal(searchRequest)
	if err != nil {
		return nil, err
	}
	resp, err := m.httpClient.Post(
		"https://photoslibrary.googleapis.com/v1/mediaItems:search",
		"application/json",
		bytes.NewBuffer(jsonStr),
	)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	d := &MediaItems{}
	err = json.NewDecoder(resp.Body).Decode(d)

	return d, err
}
