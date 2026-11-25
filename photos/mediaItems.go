package photos

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GetAllMediaItemsForAlbum is kept intentionally even though no current
// tool calls it: it is the only way in this codebase to list the contents
// of an arbitrary album. labelphotos and sortlocal only see items from a
// Picker session (i.e. items the user just picked), not an album's full
// contents, so removing this would lose that capability entirely.
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
func (m *Client) GetMediaItemsFromSession(sessionID, pageToken string) (*MediaItems, error) {
	otherParams := ""
	if pageToken != "" {
		otherParams += fmt.Sprintf("&pageToken=%s", pageToken)
	}

	resp, err := m.httpClient.Get(fmt.Sprintf(
		"https://photospicker.googleapis.com/v1/mediaItems?pageSize=100&sessionId=%s%s",
		sessionID,
		otherParams,
	))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	d := &PickedMediaItems{}
	err = json.NewDecoder(resp.Body).Decode(d)
	if err != nil {
		return nil, err
	}
	var mediaItems []*MediaItem
	for _, pickedMediaItem := range d.MediaItems {
		mediaItems = append(mediaItems, &MediaItem{
			ID:       pickedMediaItem.ID,
			Filename: pickedMediaItem.MediaFile.Filename,
			MediaMetadata: MediaMetadata{
				CreationTime: pickedMediaItem.CreateTime,
			},
		})
	}
	return &MediaItems{
		MediaItems:    mediaItems,
		NextPageToken: d.NextPageToken,
	}, nil
}

func (m *Client) GetAllMediaItemsForSessionID(sessionID string) ([]*MediaItem, error) {
	var allMediaItems []*MediaItem
	lastPageToken := ""
	for {
		mediaItems, err := m.GetMediaItemsFromSession(sessionID, lastPageToken)
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
	d := &MediaItems{}
	err := m.postJson(
		"https://photoslibrary.googleapis.com/v1/mediaItems:search",
		searchRequest,
		d,
	)
	if err != nil {
		return nil, err
	}
	if d.Error != nil {
		return nil, fmt.Errorf(
			"got an error fetching media items: %s %s",
			d.Error.Message,
			d.Error.Status,
		)
	}

	return d, nil
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
