package gosubsonic

import (
	"time"
)

// APIContainer represents the top-level response from Subsonic
type APIContainer struct {
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

	// License - returned only in GetLicense
	License SubsonicLicense

	// MusicFolders - returned only in GetMusicFolders
	MusicFolders MusicFolderContainer

	// Indexes - returned only in GetIndexes
	Indexes IndexesContainer

	// Directory - returned only in GetMusicDirectory
	Directory MusicDirectoryContainer
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

// MusicFolderContainer represents the container for one or more MusicFolders
type MusicFolderContainer struct {
	MusicFolder interface{}
}

// MusicFolder represents a top-level music folders of Subsonic
type MusicFolder struct {
	ID   int64
	Name string
}

// IndexesContainer represents the container for a slice of Index structs
type IndexesContainer struct {
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

// MusicDirectoryContainer represents the container for a slice of SubsonicDirectory structs
type MusicDirectoryContainer struct {
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
