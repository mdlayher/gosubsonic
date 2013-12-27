package gosubsonic

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// Constants to pass with each API request
const (
	CLIENT     = "gosubsonic-0.1"
	APIVERSION = "1.8.0"
)

// Subsonic client required fields
type SubsonicClient struct {
	Host     string
	Port     int
	Username string
	Password string
}

// Generate a URL for an API call using given parameters and method
func (s SubsonicClient) makeURL(method string) string {
	return fmt.Sprintf("http://%s:%d/rest/%s.view?u=%s&p=%s&c=%s&v=%s&f=json",
		s.Host, s.Port, method, s.Username, s.Password, CLIENT, APIVERSION)
}

// Fetch a list of all artists known to Subsonic
func (s SubsonicClient) FetchArtists() ([]Artist, error) {
	// List of artists to return
	artists := make([]Artist, 0)

	// Request a list of artists from API
	res, err := http.Get(s.makeURL("getArtists"))
	if err != nil {
		return artists, errors.New("HTTP request 'getArtists' failed")
	}

	// Read the entire response body, and defer it to be closed
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	// Unmarshal response JSON from API container
	var subRes ApiContainer
	json.Unmarshal(body, &subRes)

	// Check for any errors in response object
	if subRes.Response.Error != (ApiError{}) {
		// Report error and code
		return artists, errors.New(fmt.Sprintf("%d: %s", subRes.Response.Error.Code, subRes.Response.Error.Message))
	}

	// Iterate all indices to get artist lists inside
	for _, i := range subRes.Response.Artists.Index {
		// Iterate all artists and append to list
		for _, a := range i.Artist {
			artists = append(artists[:], a)
		}
	}

	// Return list of artists
	return artists, nil
}

// Fetch artist by ID from Subsonic
func (s SubsonicClient) GetArtist(id int) (Artist, error) {
	// Request artist from API, by ID
	res, err := http.Get(s.makeURL("getArtist") + "&id=" + strconv.FormatInt(int64(id), 10))
	if err != nil {
		return Artist{}, errors.New("HTTP request 'getArtist' failed")
	}

	// Read the entire response body, and defer it to be closed
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	// Unmarshal response JSON from API container
	var subRes ApiContainer
	json.Unmarshal(body, &subRes)

	// Check for any errors in response object
	if subRes.Response.Error != (ApiError{}) {
		// Report error and code
		return Artist{}, errors.New(fmt.Sprintf("%d: %s", subRes.Response.Error.Code, subRes.Response.Error.Message))
	}

	// Return artist
	return subRes.Response.Artist, nil
}
