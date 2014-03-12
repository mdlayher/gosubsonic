package gosubsonic

import (
	"time"
)

// apiContainer represents the top-level response from Subsonic
type apiContainer struct {
	Response APIStatus `json:"subsonic-response"`
}

// APIError represents any errors reported by Subsonic
type APIError struct {
	Code    int
	Message string
}

// APIStatus represents the current status of Subsonic
type APIStatus struct {
	// Common fields
	Status  string
	Version string
	Xmlns   string

	// API error - returned only when an error occurs
	Error APIError

	// license - returned only in GetLicense
	License License

	// musicFolders - returned only in GetMusicFolders
	MusicFolders apiMusicFolderContainer

	// indexes - returned only in GetIndexes
	Indexes apiIndexesContainer

	// directory - returned only in GetMusicDirectory
	Directory interface{}

	// nowPlaying - returned only in GetNowPlaying
	NowPlaying interface{}
}

// License represents the license status of Subsonic
type License struct {
	// Raw values
	DateRaw string `json:"date"`
	Email   string
	Key     string
	Valid   bool

	// Parsed values
	Date time.Time
}

// apiMusicFolderContainer represents the container for one or more MusicFolders
type apiMusicFolderContainer struct {
	MusicFolder interface{}
}

// MusicFolder represents a top-level music folders of Subsonic
type MusicFolder struct {
	ID   int64
	Name string
}

// apiIndexesContainer represents the container for a slice of Index structs
type apiIndexesContainer struct {
	Index []Index
}

// Index represents a group in the Subsonic index
type Index struct {
	// Raw values
	Name      string
	ArtistRaw interface{} `json:"artist"`

	// Artist - generated from raw interfaces
	Artist []IndexArtist
}

// IndexArtist represents an artist in the Subsonic index
type IndexArtist struct {
	ID   int64
	Name string
}

// apiMusicDirectoryContainer represents the container for a slice of Directory structs
type apiMusicDirectoryContainer struct {
	Child interface{}
}

// Content is a container used to contain the Media and Directory structs residing in this Directory
type Content struct {
	Directories []Directory
	Media       []Media
}

// Media represents a media item from Subsonic
type Media struct {
	// Raw values
	ID                    int64
	Album                 string
	AlbumID               int64
	Artist                string
	ArtistID              int64
	BitRate               int64
	ContentType           string
	CoverArt              int64
	CreatedRaw            string `json:"created"`
	DiscNumber            int64
	DurationRaw           int64 `json:"duration"`
	Genre                 string
	IsVideo               bool
	Parent                int64
	Path                  string
	Size                  int64
	Suffix                string
	Title                 string
	Track                 int64
	TranscodedContentType string
	TranscodedSuffix      string
	Type                  string
	Year                  int64

	// Parsed values
	Created  time.Time
	Duration time.Duration
}

// Directory represents a media directory from Subsonic
type Directory struct {
	// Raw values
	ID         int64
	Album      string
	Artist     string
	CoverArt   int64
	CreatedRaw string `json:"created"`
	Parent     int64
	Title      string

	// Parsed values
	Created time.Time
}

// apiNowPlayingContainer represents the container for a slice of NowPlaying structs
type apiNowPlayingContainer struct {
	Entry interface{}
}

// NowPlaying represents a now playing entry from Subsonic
type NowPlaying struct {
	// Raw values
	ID          int64
	Album       string
	AlbumID     int64
	Artist      string
	ArtistID    int64
	BitRate     int64
	ContentType string
	CoverArt    int64
	CreatedRaw  string `json:"created"`
	DiscNumber  int64
	DurationRaw int64
	Genre       string
	IsDir       bool
	IsVideo     bool
	MinutesAgo  int64
	Parent      int64
	Path        string
	PlayerID    int64
	Size        int64
	Suffix      string
	Title       string
	Track       int64
	Username    string
	Year        int64

	// Parsed values
	Created  time.Time
	Duration time.Duration
}
