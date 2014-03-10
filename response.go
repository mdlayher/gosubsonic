package gosubsonic

// APIContainer represents the top-level response from Subsonic
type APIContainer struct {
	Response APIStatusResponse `json:"subsonic-response"`
}

// APIError represents any errors reported by Subsonic
type APIError struct {
	Message string
	Code    int
}

// APIStatusResponse represents the current status of Subsonic
type APIStatusResponse struct {
	Error   APIError
	Xmlns   string
	Status  string
	Version string
}
