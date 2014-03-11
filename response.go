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
	Message string
	Code    int
}

// APIStatus represents the current status of Subsonic
type APIStatus struct {
	// Common fields
	Xmlns   string
	Status  string
	Version string

	// API error - returned only when an error occurs
	Error APIError

	// license - returned only in GetLicense
	license SubsonicLicense

	// musicFolders - returned only in GetMusicFolders
	musicFolders apiMusicFolderContainer

	// indexes - returned only in GetIndexes
	indexes apiIndexesContainer

	// directory - returned only in GetMusicDirectory
	directory apiMusicDirectoryContainer

	// nowPlaying - returned only in GetNowPlaying
	nowPlaying apiNowPlayingContainer
}

// SubsonicLicense represents the license status of Subsonic
type SubsonicLicense struct {
	// Raw values
	Valid   bool
	Email   string
	DateRaw string `json:"date"`
	Key     string

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
	Index []SubsonicIndex
}

// SubsonicIndex represents a group in the Subsonic index
type SubsonicIndex struct {
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

// apiMusicDirectoryContainer represents the container for a slice of SubsonicDirectory structs
type apiMusicDirectoryContainer struct {
	Child interface{}
}

// SubsonicDirectory represents a media directory from Subsonic
type SubsonicDirectory struct {
	// Raw values
	ID         int64
	Title      string
	CreatedRaw string `json:"created"`
	Album      string
	Parent     int64
	IsDir      bool
	Artist     string
	CoverArt   int64

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
	Genre       string
	Album       string
	IsDir       bool
	ContentType string
	IsVideo     bool
	ID          int64
	Title       string
	Username    string
	CreatedRaw  string `json:"created"`
	ArtistID    int64
	Path        string
	Year        int64
	Artist      string
	MinutesAgo  int64
	AlbumID     int64
	Track       int64
	Parent      int64
	DiscNumber  int64
	Suffix      string
	Size        int64
	DurationRaw int64
	PlayerID    int64
	BitRate     int64
	CoverArt    int64

	// Parsed values
	Created  time.Time
	Duration time.Duration
}
