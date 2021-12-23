package photos

// VideoProcessingStatus is an enum for video processing status
type VideoProcessingStatus string

const (
	// UNSPECIFIED means unspecified
	UNSPECIFIED VideoProcessingStatus = "UNSPECIFIED"
	// PROCESSING means processing
	PROCESSING VideoProcessingStatus = "PROCESSING"
	// READY means ready
	READY VideoProcessingStatus = "READY"
	// FAILED means failed
	FAILED VideoProcessingStatus = "FAILED"
)

// Photo is the structure to hold photo data
type Photo struct {
	CameraMake      string  `json:"cameraMake"`
	CameraModel     string  `json:"cameraModel"`
	FocalLength     float64 `json:"focalLength"`
	ApertureFNumber float64 `json:"apertureFNumber"`
	IsoEquivalent   int     `json:"isoEquivalent"`
	ExposureTime    string  `json:"exposureTime"`
}

// Video is the structure to hold video data
type Video struct {
	CameraMake  string                `json:"cameraMake"`
	CameraModel string                `json:"cameraModel"`
	Fps         float64               `json:"fps"`
	Status      VideoProcessingStatus `json:"status"`
}

// MediaMetadata is the structure to hold media metadata
type MediaMetadata struct {
	CreationTime string `json:"creationTime"`
	Width        string `json:"width"`
	Height       string `json:"height"`
	Photo        *Photo `json:"photo,omitempty"`
	Video        *Video `json:"video,omitempty"`
}

// ContributorInfo is the structure to hold constributor info
type ContributorInfo struct {
	ProfilePictureBaseURL string `json:"profilePictureBaseUrl"`
	DisplayName           string `json:"displayName"`
}

// MediaItem is the structure to hold a media item
type MediaItem struct {
	ID              string           `json:"id"`
	Description     *string          `json:"description,omitempty"`
	ProductULR      string           `json:"productUrl"`
	BaseURL         string           `json:"baseUrl"`
	MimeType        string           `json:"mimeType"`
	MediaMetadata   MediaMetadata    `json:"mediaMetadata"`
	ContributorInfo *ContributorInfo `json:"contributorInfo,omitempty"`
	Filename        string           `json:"filename"`
}

// MediaItems is the structure to hold media items
type MediaItems struct {
	MediaItems    []*MediaItem   `json:"mediaItems"`
	NextPageToken string         `json:"nextPageToken"`
	Error         *ErrorResponse `json:"error"`
}

type SharedAlbumOptions struct {
	IsCollaborative bool `json:"isCollaborative"`
	IsCommentable   bool `json:"isCommentable"`
}

type ShareInfo struct {
	SharedAlbumOptions SharedAlbumOptions `json:"sharedAlbumOptions"`
	ShareableUrl       string             `json:"shareableUrl"`
	ShareToken         string             `json:"shareToken"`
	IsJoined           bool               `json:"isJoined"`
	IsOwned            bool               `json:"isOwned"`
	IsJoinable         bool               `json:"isJoinable"`
}

type Album struct {
	ID                    string    `json:"id"`
	Title                 string    `json:"title"`
	ProductULR            string    `json:"productUrl"`
	IsWritable            bool      `json:"isWriteable"`
	ShareInfo             ShareInfo `json:"shareInfo"`
	MediaItemsCount       string    `json:"mediaItemsCount"`
	CoverPhotoBaseUrl     string    `json:"coverPhotoBaseUrl"`
	CoverPhotoMediaItemId string    `json:"coverPhotoMediaItemId"`
}

type Albums struct {
	Albums        []*Album `json:"albums"`
	NextPageToken string   `json:"nextPageToken"`
}
