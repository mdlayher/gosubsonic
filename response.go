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
