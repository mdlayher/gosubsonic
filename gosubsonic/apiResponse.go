package gosubsonic

// Top level response from Subsonic
type ApiContainer struct {
	Response ApiStatusResponse `json:"subsonic-response"`
}

// Status response from Subsonic
type ApiStatusResponse struct {
	Error   ApiError
	Xmlns   string
	Status  string
	Version string
	Artists ApiArtist
	Artist  Artist
}

// Any errors reported by API
type ApiError struct {
	Message string
	Code    int
}

// List of indices containing artists
type ApiArtist struct {
	Index []ApiArtistIndex
}

// Index containing list of artists
type ApiArtistIndex struct {
	Name   string
	Artist []Artist
}
