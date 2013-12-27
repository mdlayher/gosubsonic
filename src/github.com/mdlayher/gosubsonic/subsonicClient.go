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

// Fetch JSON from specified URL and parse into ApiContainer
func fetchJSON(url string) (ApiContainer, error) {
	// Make an API request
	res, err := http.Get(url)
	if err != nil {
		return ApiContainer{}, errors.New("HTTP request failed: " + url)
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
		return ApiContainer{}, errors.New(fmt.Sprintf("%d: %s", subRes.Response.Error.Code, subRes.Response.Error.Message))
	}

	// Return the response container
	return subRes, nil
}

// Fetch a list of all artists known to Subsonic
func (s SubsonicClient) FetchArtists() ([]Artist, error) {
	// List of artists to return
	artists := make([]Artist, 0)

	// Query for list of all artists
	res, err := fetchJSON(s.makeURL("getArtists"))
	if err != nil {
		return artists, err
	}

	// Iterate all indices to get artist lists inside
	for _, i := range res.Response.Artists.Index {
		// Iterate all artists and append to list
		for _, a := range i.Artist {
			a.Client = s
			artists = append(artists[:], a)
		}
	}

	// Return list of artists
	return artists, nil
}

// Fetch artist by ID from Subsonic
func (s SubsonicClient) GetArtist(id int) (Artist, error) {
	// Request artist from API, by ID
	res, err := fetchJSON(s.makeURL("getArtist") + "&id=" + strconv.FormatInt(int64(id), 10))
	if err != nil {
		return Artist{}, err
	}

	artist := res.Response.Artist
	artist.Client = s

	// Return artist
	return artist, nil
}
