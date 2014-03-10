package gosubsonic

// ApiContainer represents the top-level response from Subsonic
type ApiContainer struct {
	Response ApiStatusResponse `json:"subsonic-response"`
}

// ApiError represents any errors reported by Subsonic
type ApiError struct {
	Message string
	Code    int
}

// ApiStatusResponse represents the current status of Subsonic
type ApiStatusResponse struct {
	Error   ApiError
	Xmlns   string
	Status  string
	Version string
}
