package gosubsonic

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Constants to pass with each API request
const (
	CLIENT     = "gosubsonic-git-master"
	APIVERSION = "1.8.0"
)

// Client represents the required parameters to connect to a Subsonic server
type Client struct {
	Host     string
	Username string
	Password string
}

// New creates a new Client using the specified parameters
func New(host string, username string, password string) (*Client, error) {
	// Generate a new Subsonic client
	client := Client{
		Host:     host,
		Username: username,
		Password: password,
	}

	// Attempt to ping the Subsonic server
	_, err := client.Ping()
	return &client, err
}

// -- System --

// Ping checks the connectivity of a Subsonic server
func (s Client) Ping() (*APIStatus, error) {
	// Nil error means that ping is successful
	res, err := fetchJSON(s.makeURL("ping"))
	if err != nil {
		return nil, err
	}

	return &res.Response, nil
}

// GetLicense retrieves details about the Subsonic server license
func (s Client) GetLicense() (*License, error) {
	// Retrieve license information from Subsonic
	res, err := fetchJSON(s.makeURL("getLicense"))
	if err != nil {
		return nil, err
	}

	// Check for a license in the response
	if &res.Response.License == nil {
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
func (s Client) GetMusicFolders() ([]MusicFolder, error) {
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
	// Unknown case
	default:
		return nil, errors.New("gosubsonic: failed to parse getMusicFolders response")
	}

	// Iterate each item
	for _, i := range iface {
		// Type hint to appropriate type
		if m, ok := i.(map[string]interface{}); ok {
			// Create a music folder from the map
			f := MusicFolder{
				// Note: ID is always an int64, so we can safely convert the float64
				ID:   int64(m["id"].(float64)),
				Name: m["name"].(string),
			}

			// Add folder to collection
			folders = append(folders, f)
		}
	}

	// Return output folders
	return folders, nil
}

// GetIndexes returns an indexed structure of all artists from Subsonic
func (s Client) GetIndexes(folderID int64, modified int64) ([]Index, error) {
	// Additional parameters for query
	query := ""

	// Check for a set folder ID (ID >= 0)
	if folderID >= 0 {
		query = query + "&musicFolderId=" + strconv.FormatInt(folderID, 10)
	}

	// Check for a modify time (modified >= 0)
	if modified >= 0 {
		query = query + "&ifModifiedSince=" + strconv.FormatInt(modified, 10)
	}

	// Retrieve indexes from Subsonic, with query parameters
	res, err := fetchJSON(s.makeURL("getIndexes") + query)
	if err != nil {
		return nil, err
	}

	// Generate new index with proper information
	outIndex := make([]Index, 0)

	// Iterate all raw Index structs
	for _, index := range res.Response.Indexes.Index {
		// Slice of IndexArtist structs to output
		artists := make([]IndexArtist, 0)

		// Slice of interfaces to parse out response
		iface := make([]interface{}, 0)

		// Parse response from interface{}, which may be one or more items
		switch index.ArtistRaw.(type) {
		// Single item
		case map[string]interface{}:
			iface = append(iface, index.ArtistRaw.(interface{}))
		// Multiple items
		case []interface{}:
			iface = index.ArtistRaw.([]interface{})
		// Unknown case
		default:
			return nil, errors.New("gosubsonic: failed to parse getIndexes response")
		}

		// Iterate each item
		for _, i := range iface {
			// Type hint to appropriate type
			if m, ok := i.(map[string]interface{}); ok {
				// Create a IndexArtist from map
				a := IndexArtist{
					// Note: ID is always an int64, so we can safely convert the float64
					ID:   int64(m["id"].(float64)),
					Name: m["name"].(string),
				}

				// Add artist to collection
				artists = append(artists, a)
			}
		}

		// Store artists collection in out index, nullify raw values
		index.ArtistRaw = nil
		index.Artist = artists
		outIndex = append(outIndex, index)
	}

	// Return output
	return outIndex, nil
}

// GetMusicDirectory returns a list of all content in a music directory
func (s Client) GetMusicDirectory(folderID int64) (*Content, error) {
	// Retrieve a list of files in a given directory from Subsonic
	res, err := fetchJSON(s.makeURL("getMusicDirectory") + "&id=" + strconv.FormatInt(folderID, 10))
	if err != nil {
		return nil, err
	}

	// Slice of Media and Directory structs to return
	media := make([]Media, 0)
	directories := make([]Directory, 0)

	// Slice of interfaces to parse out response
	iface := make([]interface{}, 0)

	// Check if missing "child" element, meaning the directory is empty
	if _, ok := res.Response.Directory.(apiMusicDirectoryContainer); !ok {
		return &Content{}, nil
	}

	// Parse response from interface{}, which may be one or more items
	ch := res.Response.Directory.(apiMusicDirectoryContainer).Child
	switch ch.(type) {
	// Single item
	case map[string]interface{}:
		iface = append(iface, ch.(interface{}))
	// Multiple items
	case []interface{}:
		iface = ch.([]interface{})
	// Unknown case
	default:
		return nil, errors.New("gosubsonic: failed to parse getMusicDirectory response")
	}

	// Iterate each item
	for _, i := range iface {
		// Type hint to appropriate type
		if m, ok := i.(map[string]interface{}); ok {
			// First, we have to work out some shared fields between directories and media

			// Subsonic problem: albums with numeric titles return as integers
			// Therefore, we have to check for a float64 as well
			var album string
			switch m["album"].(type) {
			// No album title
			case nil:
				break
			case string:
				album = m["album"].(string)
			case float64:
				album = strconv.FormatInt(int64(m["album"].(float64)), 10)
			default:
				return nil, errors.New("gosubsonic: unknown Album data type for getMusicDirectory")
			}

			// Same thing with song title
			var title string
			switch m["title"].(type) {
			case string:
				title = m["title"].(string)
			case float64:
				title = strconv.FormatInt(int64(m["title"].(float64)), 10)
			default:
				return nil, errors.New("gosubsonic: unknown Title data type for getMusicDirectory")
			}

			// Some albums may not have cover art, so we check individually for it
			var coverArt int64
			if c, ok := m["coverArt"].(float64); ok {
				coverArt = int64(c)
			}

			// Parse CreatedRaw into a time.Time struct
			created, err := time.Parse("2006-01-02T15:04:05", m["created"].(string))
			if err != nil {
				return nil, err
			}

			// Is this a directory?
			if b, ok := m["isDir"].(bool); b && ok {
				// Create a directory from the map
				d := Directory{
					// Note: ID is always an int64, so we can safely convert the float64
					ID:         int64(m["id"].(float64)),
					Album:      album,
					Artist:     m["artist"].(string),
					CoverArt:   coverArt,
					Created:    created,
					CreatedRaw: m["created"].(string),
					Parent:     int64(m["parent"].(float64)),
					Title:      title,
				}

				// Add directory to collection
				directories = append(directories, d)
			} else {
				// For the time being, we will not support video media, so skip it
				if b, ok := m["isVideo"].(bool); b && ok {
					continue
				}

				// If not, it's media
				med := Media{
					// Note: ID is always an int64, so we can safely convert the float64
					ID:          int64(m["id"].(float64)),
					Album:       album,
					ContentType: m["contentType"].(string),
					CoverArt:    coverArt,
					Created:     created,
					CreatedRaw:  m["created"].(string),
					DurationRaw: int64(m["duration"].(float64)),
					IsVideo:     m["isVideo"].(bool),
					Parent:      int64(m["parent"].(float64)),
					Path:        m["path"].(string),
					Suffix:      m["suffix"].(string),
					Title:       title,
					Type:        m["type"].(string),
				}

				// Subsonic is very inconsistent, so we have to check for optional items
				if a, ok := m["albumId"].(float64); ok {
					med.AlbumID = int64(a)
				}
				if a, ok := m["artist"].(string); ok {
					med.Artist = a
				}
				if a, ok := m["artistId"].(float64); ok {
					med.ArtistID = int64(a)
				}
				if d, ok := m["discNumber"].(float64); ok {
					med.DiscNumber = int64(d)
				}
				if g, ok := m["genre"].(string); ok {
					med.Genre = g
				}
				if t, ok := m["track"].(float64); ok {
					med.Track = int64(t)
				}
				if y, ok := m["year"].(float64); ok {
					med.Year = int64(y)
				}

				// Returned only in transcodes
				if t, ok := m["transcodedContentType"].(string); ok {
					med.TranscodedContentType = t
				}
				if t, ok := m["transcodedSuffix"].(string); ok {
					med.TranscodedSuffix = t
				}

				// Parse DurationRaw into a time.Duration struct
				d, err := time.ParseDuration(strconv.FormatInt(med.DurationRaw, 10) + "s")
				if err != nil {
					return nil, err
				}
				med.Duration = d

				// Add media to collection
				media = append(media, med)
			}
		}
	}

	// Return output content
	return &Content{Media: media, Directories: directories}, nil
}

// -- Album/song lists --

// GetNowPlaying returns a list of tracks which are currently being played
func (s Client) GetNowPlaying() ([]NowPlaying, error) {
	// Retreive all tracks currently playing from Subsonic
	res, err := fetchJSON(s.makeURL("getNowPlaying"))
	if err != nil {
		return nil, err
	}

	// Subsonic problem: when no songs are playing, the apiNowPlayingContainer will be an empty string
	// To work around this, we have to check if it's a string and bail out if so
	if _, ok := res.Response.NowPlaying.(string); ok {
		return nil, nil
	}

	// Slice of NowPlaying structs to return
	nowPlaying := make([]NowPlaying, 0)

	// Slice of interfaces to parse out response
	iface := make([]interface{}, 0)

	// Parse response from interface{}, which may be one or more items
	en := res.Response.NowPlaying.(apiNowPlayingContainer).Entry
	switch en.(type) {
	// Single item
	case map[string]interface{}:
		iface = append(iface, en.(interface{}))
	// Multiple items
	case []interface{}:
		iface = en.([]interface{})
	// Unknown case
	default:
		return nil, errors.New("gosubsonic: failed to parse getNowPlaying response")
	}

	// Iterate each item
	for _, i := range iface {
		// Type hint to appropriate type
		if m, ok := i.(map[string]interface{}); ok {
			// Create a now playing entry from the map
			n := NowPlaying{
				ID:          int64(m["id"].(float64)),
				AlbumID:     int64(m["albumId"].(float64)),
				Artist:      m["artist"].(string),
				ArtistID:    int64(m["artistId"].(float64)),
				BitRate:     int64(m["bitRate"].(float64)),
				ContentType: m["contentType"].(string),
				CreatedRaw:  m["created"].(string),
				DiscNumber:  int64(m["discNumber"].(float64)),
				DurationRaw: int64(m["duration"].(float64)),
				Genre:       m["genre"].(string),
				IsDir:       m["isDir"].(bool),
				IsVideo:     m["isVideo"].(bool),
				MinutesAgo:  int64(m["minutesAgo"].(float64)),
				Parent:      int64(m["parent"].(float64)),
				Path:        m["path"].(string),
				PlayerID:    int64(m["playerId"].(float64)),
				Size:        int64(m["size"].(float64)),
				Suffix:      m["suffix"].(string),
				Title:       m["title"].(string),
				Track:       int64(m["track"].(float64)),
				Year:        int64(m["year"].(float64)),
			}

			// Subsonic problem: albums with numeric titles return as integers
			// Therefore, we have to check for a float64 as well
			switch m["album"].(type) {
			case string:
				n.Album = m["album"].(string)
			case float64:
				n.Album = strconv.FormatInt(int64(m["album"].(float64)), 10)
			default:
				return nil, errors.New("gosubsonic: unknown Album data type for getNowPlaying")
			}

			// Some albums may not have cover art, so we check individually for it
			if c, ok := m["coverArt"].(float64); ok {
				n.CoverArt = int64(c)
			}

			// Parse CreatedRaw into a time.Time struct
			t, err := time.Parse("2006-01-02T15:04:05", n.CreatedRaw)
			if err != nil {
				return nil, err
			}
			n.Created = t

			// Parse DurationRaw into a time.Duration struct
			d, err := time.ParseDuration(strconv.FormatInt(n.DurationRaw, 10) + "s")
			if err != nil {
				return nil, err
			}
			n.Duration = d

			// Add now playing to collection
			nowPlaying = append(nowPlaying, n)
		}
	}

	// Return output entries
	return nowPlaying, nil
}

// -- Media retrieval --

// StreamOptions represents additional options for the Stream() method
type StreamOptions struct {
	MaxBitRate            int64
	Format                string
	TimeOffset            int64
	Size                  string
	EstimateContentLength bool
}

// Stream returns a io.ReadCloser which contains a processed media file stream, with an optional StreamOptions struct
func (s Client) Stream(id int64, options *StreamOptions) (io.ReadCloser, error) {
	// Check for no options, which will do a simple stream
	if options == nil {
		return fetchBinary(s.makeURL("stream") + "&id=" + strconv.FormatInt(id, 10))
	}

	// Check for additional options
	optStr := ""

	// maxBitRate
	if options.MaxBitRate > 0 {
		optStr = optStr + "&maxBitRate=" + strconv.FormatInt(options.MaxBitRate, 10)
	}

	// format
	if options.Format != "" {
		optStr = optStr + "&format=" + options.Format
	}

	// timeOffset
	if options.TimeOffset > 0 {
		optStr = optStr + "&timeOffset=" + strconv.FormatInt(options.TimeOffset, 10)
	}

	// size
	if options.Size != "" {
		optStr = optStr + "&size=" + options.Size
	}

	// estimateContentLength
	if options.EstimateContentLength {
		optStr = optStr + "&estimateContentLength=true"
	}

	// Stream with options
	return fetchBinary(s.makeURL("stream") + "&id=" + strconv.FormatInt(id, 10) + optStr)
}

// Download returns a io.ReadCloser which contains a raw, non-transcoded media file stream
func (s Client) Download(id int64) (io.ReadCloser, error) {
	return fetchBinary(s.makeURL("download") + "&id=" + strconv.FormatInt(id, 10))
}

// -- Functions --

// makeURL Generates a URL for an API call using given parameters and method
func (s Client) makeURL(method string) string {
	return fmt.Sprintf("http://%s/rest/%s.view?u=%s&p=%s&c=%s&v=%s&f=json",
		s.Host, method, s.Username, s.Password, CLIENT, APIVERSION)
}

// fetchBinary retrieves a binary stream from a specified URL and returns a io.ReadCloser on the stream
func fetchBinary(url string) (io.ReadCloser, error) {
	// Perform HTTP GET request
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("gosubsonic: HTTP request failed: %s - %s", err.Error(), url)
	}

	// Check for JSON content type, meaning file is not binary
	if strings.Contains(res.Header.Get("Content-Type"), "application/json") {
		// Read the entire response body, and defer it to be closed
		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()

		// Unmarshal response JSON from API container
		var subRes apiContainer
		err = json.Unmarshal(body, &subRes)
		if err != nil {
			return nil, fmt.Errorf("gosubsonic: failed to parse response JSON: %s - %s", err.Error(), url)
		}

		// Return the error
		return nil, fmt.Errorf("gosubsonic: %d: %s", subRes.Response.Error.Code, subRes.Response.Error.Message)
	}

	// Return response reader for body
	return res.Body, nil
}

// fetchJSON retrives JSON from a specified URL and parses it into an apiContainer
func fetchJSON(url string) (*apiContainer, error) {
	// Make an API request
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("gosubsonic: HTTP request failed: %s - %s", err.Error(), url)
	}

	// Read the entire response body, and defer it to be closed
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	// Unmarshal response JSON from API container
	var subRes apiContainer
	err = json.Unmarshal(body, &subRes)
	if err != nil {
		return nil, fmt.Errorf("gosubsonic: failed to parse response JSON: %s - %s", err.Error(), url)
	}

	// Check for any errors in response object
	if subRes.Response.Error != (APIError{}) {
		// Report error and code
		return nil, fmt.Errorf("gosubsonic: %d: %s", subRes.Response.Error.Code, subRes.Response.Error.Message)
	}

	// Return the response container
	return &subRes, nil
}
