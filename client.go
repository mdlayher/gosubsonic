package gosubsonic

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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
	return &SubsonicClient{
		Host:     host,
		Username: username,
		Password: password,
	}, nil
}

// Ping checks the connectivity of a Subsonic server
func (s SubsonicClient) Ping() error {
	// Nil error means that ping is successful
	if _, err := fetchJSON(s.makeURL("ping")); err != nil {
		return err
	}

	return nil
}

// makeURL Generates a URL for an API call using given parameters and method
func (s SubsonicClient) makeURL(method string) string {
	return fmt.Sprintf("http://%s/rest/%s.view?u=%s&p=%s&c=%s&v=%s&f=json",
		s.Host, method, s.Username, s.Password, CLIENT, APIVERSION)
}

// fetchJSON from specified URL and parse into ApiContainer
func fetchJSON(url string) (*ApiContainer, error) {
	// Make an API request
	res, err := http.Get(url)
	if err != nil {
		return nil, errors.New("gosubsonic: HTTP request failed: " + url)
	}

	// Read the entire response body, and defer it to be closed
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	// Unmarshal response JSON from API container
	var subRes ApiContainer
	err = json.Unmarshal([]byte(body), &subRes)
	if err != nil {
		return nil, errors.New("Failed to parse response JSON: " + url)
	}

	// Check for any errors in response object
	if subRes.Response.Error != (ApiError{}) {
		// Report error and code
		return nil, fmt.Errorf("gosubsonic: %d: %s", subRes.Response.Error.Code, subRes.Response.Error.Message)
	}

	// Return the response container
	return &subRes, nil
}
