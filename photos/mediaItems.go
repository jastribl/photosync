package photos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/jastribl/photosync/files"
)

const allMediaItemsCacheFile = "cache/allMediaItems.json"

func (m *Client) CacheAndReturnAllMediaItems() ([]*MediaItem, error) {
	var allMediaItems []*MediaItem
	for lastPageToken, dedupMap := "", map[string]bool{}; ; {
		mediaItems, err := m.getMediaItems(lastPageToken)
		if err != nil {
			return nil, err
		}
		for _, mediaItem := range mediaItems.MediaItems {
			if _, found := dedupMap[mediaItem.ID]; !found {
				allMediaItems = append(allMediaItems, mediaItem)
				dedupMap[mediaItem.ID] = true
			}
		}
		lastPageToken = mediaItems.NextPageToken
		if lastPageToken == "" {
			break
		}
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

func (m *Client) GetAllMediaItemsWithCache() ([]*MediaItem, error) {
	if !files.FileExists(allMediaItemsCacheFile) {
		allMediaItems, err := m.CacheAndReturnAllMediaItems()
		if err != nil {
			return nil, err
		}
		return allMediaItems, nil
	}

	var allMediaItems []*MediaItem

	bytes, err := ioutil.ReadFile(allMediaItemsCacheFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(bytes), &allMediaItems)
	if err != nil {
		return nil, err
	}

	return allMediaItems, err
}

func (m *Client) GetAllLowercaseFilenameToMediaItemMapWithCache() (map[string][]*MediaItem, error) {
	allMediaItems, err := m.GetAllMediaItemsWithCache()
	if err != nil {
		return nil, err
	}

	return MediaItemsToLowercaseFilenameMap(allMediaItems), nil
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
	if err != nil {
		return nil, err
	}
	if d.Error != nil {
		log.Fatal("got an error fetching media items: " + d.Error.Message + " " + d.Error.Status)
	}
	log.Printf("Got %d media items\n", len(d.MediaItems))

	return d, err
}

type SearchRequest struct {
	PageSize  int    `json:"pageSize"`
	PageToken string `json:"pageToken"`
	AlbumId   string `json:"albumId"`
}

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
	if err != nil {
		return nil, err
	}
	if d.Error != nil {
		log.Fatal("got an error fetching media items: " + d.Error.Message + " " + d.Error.Status)
	}

	return d, err
}

func MediaItemsToLowercaseFilenameMap(mediaItems []*MediaItem) map[string][]*MediaItem {
	lowercaseFilenamesToMediaItems := map[string][]*MediaItem{}
	for _, item := range mediaItems {
		lowerFilename := strings.ToLower(item.Filename)
		if list, ok := lowercaseFilenamesToMediaItems[lowerFilename]; ok {
			lowercaseFilenamesToMediaItems[lowerFilename] = append(list, item)
		} else {
			lowercaseFilenamesToMediaItems[lowerFilename] = []*MediaItem{item}
		}
	}

	return lowercaseFilenamesToMediaItems
}
