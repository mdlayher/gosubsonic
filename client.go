package gosubsonic

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Constants to pass with each API request
const (
	CLIENT     = "gosubsonic-git-master"
	APIVERSION = "1.8.0"
)

// SubsonicClient represents the required parameters to connect to a Subsonic server
type SubsonicClient struct {
	Host     string
	Username string
	Password string
}

// New creates a new SubsonicClient using the specified parameters
func New(host string, username string, password string) (*SubsonicClient, error) {
	// Generate a new Subsonic client
	client := SubsonicClient{
		Host:     host,
		Username: username,
		Password: password,
	}

	// Attempt to ping the Subsonic server
	return &client, client.Ping()
}

// -- System --

// Ping checks the connectivity of a Subsonic server
func (s SubsonicClient) Ping() error {
	// Nil error means that ping is successful
	if _, err := fetchJSON(s.makeURL("ping")); err != nil {
		return err
	}

	return nil
}

// GetLicense retrieves details about the Subsonic server license
func (s SubsonicClient) GetLicense() (*SubsonicLicense, error) {
	// Retrieve license information from Subsonic
	res, err := fetchJSON(s.makeURL("getLicense"))
	if err != nil {
		return nil, err
	}

	// Check for a license in the response
	if res.Response.License == (SubsonicLicense{}) {
		return nil, errors.New("gosubsonic: no license found")
	}

	// Parse raw date into a time.Time struct, using the special Go date for parsing
	// reference: http://golang.org/pkg/time/#Parse
	t, err := time.Parse("2006-01-02T15:04:05", res.Response.License.DateRaw)
	if err != nil {
		return nil, err
	}
	res.Response.License.Date = t

	return &res.Response.License, nil
}

// -- Browsing --

// GetMusicFolders returns the configured top-level music folders
func (s SubsonicClient) GetMusicFolders() ([]MusicFolder, error) {
	// Retrieve top-level music folders from Subsonic
	res, err := fetchJSON(s.makeURL("getMusicFolders"))
	if err != nil {
		return nil, err
	}

	// Slice of MusicFolders to return
	folders := make([]MusicFolder, 0)

	// Slice of interfaces to parse out response
	iface := make([]interface{}, 0)

	// Parse response from interface{}, which may be one or more items
	mf := res.Response.MusicFolders.MusicFolder
	switch mf.(type) {
	// Single item
	case map[string]interface{}:
		iface = append(iface, mf.(interface{}))
	// Multiple items
	case []interface{}:
		iface = mf.([]interface{})
	}

	// Iterate each item
	for _, i := range iface {
		// Type hint to appropriate type
		if m, ok := i.(map[string]interface{}); ok {
			// Create a music folder from the map
			f := MusicFolder{
				// Note: ID is always an int64, so we can safely convert the float64
				ID: int64(m["id"].(float64)),
				Name: m["name"].(string),
			}

			// Add folder to collection
			folders = append(folders, f)
		}
	}

	// Return output folders
	return folders, nil
}

// -- Functions --

// makeURL Generates a URL for an API call using given parameters and method
func (s SubsonicClient) makeURL(method string) string {
	return fmt.Sprintf("http://%s/rest/%s.view?u=%s&p=%s&c=%s&v=%s&f=json",
		s.Host, method, s.Username, s.Password, CLIENT, APIVERSION)
}

// fetchJSON from specified URL and parse into APIContainer
func fetchJSON(url string) (*APIContainer, error) {
	// Make an API request
	res, err := http.Get(url)
	if err != nil {
		return nil, errors.New("gosubsonic: HTTP request failed: " + url)
	}

	// Read the entire response body, and defer it to be closed
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	// Unmarshal response JSON from API container
	var subRes APIContainer
	err = json.Unmarshal(body, &subRes)
	if err != nil {
		return nil, errors.New("Failed to parse response JSON: " + url)
	}

	// Check for any errors in response object
	if subRes.Response.Error != (APIError{}) {
		// Report error and code
		return nil, fmt.Errorf("gosubsonic: %d: %s", subRes.Response.Error.Code, subRes.Response.Error.Message)
	}

	// Return the response container
	return &subRes, nil
}
