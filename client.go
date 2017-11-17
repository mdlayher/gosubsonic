package gosubsonic

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
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

// dataSource represents a data source for a Subsonic client (could be HTTP, mock, etc)
type dataSource interface {
	Get(string) (*apiContainer, error)
}

// Client represents the required parameters to connect to a Subsonic server
type Client struct {
	Host     string
	Username string
	Password string
	source   dataSource
}

// New creates a new Client using the specified parameters
func New(host string, username string, password string) (*Client, error) {
	// Generate a new Subsonic client
	client := Client{
		Host:     host,
		Username: username,
		Password: password,

		// Use HTTP as the data source
		source: httpDataSource{},
	}

	// Attempt to ping the Subsonic server
	_, err := client.Ping()
	return &client, err
}

// NewMock creates a new Client which receives mock data instead of connecting to a Subsonic server
func NewMock() (*Client, error) {
	// Generate a new mock client
	client := Client{
		Host: "__MOCK__",

		// Use mock data as the data source
		source: mockDataSource{},
	}

	// Initialize mock data
	if err := mockInit(client); err != nil {
		return nil, errors.New("gosubsonic: failed to initialize mock client")
	}

	return &client, nil
}

// -- System --

// Ping checks the connectivity of a Subsonic server
func (s Client) Ping() (*APIStatus, error) {
	// Nil error means that ping is successful
	res, err := s.source.Get(s.makeURL("ping"))
	if err != nil {
		return nil, err
	}

	return &res.Response, nil
}

// GetLicense retrieves details about the Subsonic server license
func (s Client) GetLicense() (*License, error) {
	// Retrieve license information from Subsonic
	res, err := s.source.Get(s.makeURL("getLicense"))
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
	res, err := s.source.Get(s.makeURL("getMusicFolders"))
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
	res, err := s.source.Get(s.makeURL("getIndexes") + query)
	if err != nil {
		return nil, err
	}

	// Generate new index with proper information
	outIndex := make([]Index, 0)

	// Slice of interfaces to parse out response
	iface := make([]interface{}, 0)

	// Parse response from interface{}, which may be one or more items
	idx := res.Response.Indexes.Index
	switch idx.(type) {
	// Single item
	case map[string]interface{}:
		iface = append(iface, idx.(interface{}))
	// Multiple items
	case []interface{}:
		iface = idx.([]interface{})
	// Unknown case
	default:
		return nil, errors.New("gosubsonic: failed to parse getIndexes response")
	}

	// Iterate each index item
	for _, i := range iface {
		m, ok := i.(map[string]interface{})
		if !ok {
			continue
		}

		// Create an index
		index := Index{
			Name: m["name"].(string),
			ArtistRaw: m["artist"],
		}

		// Slice of IndexArtist structs to output
		artists := make([]IndexArtist, 0)

		// Slice of interfaces to parse out response
		ifaceArtists := make([]interface{}, 0)

		// Parse response from interface{}, which may be one or more items
		switch index.ArtistRaw.(type) {
		// Single item
		case map[string]interface{}:
			ifaceArtists = append(ifaceArtists, index.ArtistRaw.(interface{}))
		// Multiple items
		case []interface{}:
			ifaceArtists = index.ArtistRaw.([]interface{})
		// Unknown case
		default:
			return nil, errors.New("gosubsonic: failed to parse getIndexes response")
		}

		// Iterate each item
		for _, ia := range ifaceArtists {
			// Type hint to appropriate type
			ma, ok := ia.(map[string]interface{})
			if !ok {
				continue
			}

			// Name
			name, err := ifaceToString(ma["name"])
			if err != nil {
				return nil, err
			}

			// Create a IndexArtist from map
			id, _ := strconv.ParseInt(ma["id"].(string), 0, 64)
			a := IndexArtist{
				// Note: ID is always an int64, so we can safely convert the float64
				ID:   id,
				Name: name,
			}

			// Add artist to collection
			artists = append(artists, a)
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
	res, err := s.source.Get(s.makeURL("getMusicDirectory") + "&id=" + strconv.FormatInt(folderID, 10))
	if err != nil {
		return nil, err
	}

	// Slice of Audio, Directory, Video structs to return
	audio := make([]Audio, 0)
	directories := make([]Directory, 0)
	video := make([]Video, 0)

	// Slice of interfaces to parse out response
	iface := make([]interface{}, 0)

	// Parse response from interface{}, which may be one or more items
	ch := res.Response.Directory.Child
	switch ch.(type) {
	// No items
	case nil:
		break
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

			// Artist
			artist, err := ifaceToString(m["artist"])
			if err != nil {
				return nil, err
			}

			// Album
			album, err := ifaceToString(m["album"])
			if err != nil {
				return nil, err
			}

			// Title
			title, err := ifaceToString(m["title"])
			if err != nil {
				return nil, err
			}

			// Some albums may not have cover art, so we check individually for it
			var coverArt int64
			if c, ok := m["coverArt"].(float64); ok {
				coverArt = int64(c)
			}

			// Parse CreatedRaw into a time.Time struct
			created, err := time.Parse("2006-01-02T15:04:05Z", m["created"].(string))
			if err != nil {
				return nil, err
			}

			// Is this a directory?
			if b, ok := m["isDir"].(bool); b && ok {
				id, _ := strconv.ParseInt(m["id"].(string), 0, 64)
				parentId, _ := strconv.ParseInt(m["parent"].(string), 0, 64)
				// Create a directory from the map
				d := Directory{
					// Note: ID is always an int64, so we can safely convert the float64
					ID:         id,
					Album:      album,
					Artist:     artist,
					CoverArt:   coverArt,
					Created:    created,
					CreatedRaw: m["created"].(string),
					Parent:     parentId,
					Title:      title,
				}

				// Add directory to collection
				directories = append(directories, d)
			} else {
				// If not a directory, this is a media item
				// Parse shared media field items
				var id int64
				if i, err := strconv.ParseInt(m["id"].(string), 0, 64); err==nil {
					id = i;
				}

				var bitRate int64
				if b, ok := m["bitRate"].(float64); ok {
					bitRate = int64(b)
				}

				var contentType string
				if c, ok := m["contentType"].(string); ok {
					contentType = c
				}

				var createdRaw string
				if c, ok := m["created"].(string); ok {
					createdRaw = c
				}

				var durationRaw int64
				if d, ok := m["duration"].(float64); ok {
					durationRaw = int64(d)
				}

				// Parse DurationRaw into a time.Duration struct
				duration, err := time.ParseDuration(strconv.FormatInt(durationRaw, 10) + "s")
				if err != nil {
					return nil, err
				}

				var parent int64
				if p, err := strconv.ParseInt(m["parent"].(string), 0, 64); err==nil {
					parent = p
				}

				var path string
				if p, ok := m["path"].(string); ok {
					path = html.UnescapeString(p)
				}

				var size int64
				if s, ok := m["size"].(float64); ok {
					size = int64(s)
				}

				var suffix string
				if s, ok := m["suffix"].(string); ok {
					suffix = s
				}

				var mType string
				if t, ok := m["type"].(string); ok {
					mType = t
				}

				// Returned only in transcodes
				var transcodedContentType string
				if t, ok := m["transcodedContentType"].(string); ok {
					transcodedContentType = t
				}

				var transcodedSuffix string
				if t, ok := m["transcodedSuffix"].(string); ok {
					transcodedSuffix = t
				}

				// Check if this item is a video
				if b, ok := m["isVideo"].(bool); b && ok {
					med := Video{
						ID:          id,
						BitRate:     bitRate,
						ContentType: contentType,
						CoverArt:    coverArt,
						Created:     created,
						CreatedRaw:  createdRaw,
						Duration:    duration,
						DurationRaw: durationRaw,
						Parent:      parent,
						Path:        path,
						Size:        size,
						Suffix:      suffix,
						Title:       title,
						TranscodedContentType: transcodedContentType,
						TranscodedSuffix:      transcodedSuffix,
					}

					// Add video to collection
					video = append(video, med)
				} else {
					// Else, this is an audio item
					med := Audio{
						// Note: ID is always an int64, so we can safely convert the float64
						ID:          id,
						Album:       album,
						Artist:      artist,
						BitRate:     bitRate,
						ContentType: contentType,
						CoverArt:    coverArt,
						Created:     created,
						CreatedRaw:  createdRaw,
						Duration:    duration,
						DurationRaw: durationRaw,
						Parent:      parent,
						Path:        path,
						Size:        size,
						Suffix:      suffix,
						Title:       title,
						Type:        mType,
						TranscodedContentType: transcodedContentType,
						TranscodedSuffix:      transcodedSuffix,
					}

					// Subsonic is very inconsistent, so we have to check for optional items
					if a, ok := m["albumId"].(float64); ok {
						med.AlbumID = int64(a)
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

					// Add audio to collection
					audio = append(audio, med)
				}
			}
		}
	}

	// Return output content
	return &Content{
		Audio:       audio,
		Directories: directories,
		Video:       video,
	}, nil
}

// -- Album/song lists --

// GetNowPlaying returns a list of tracks which are currently being played
func (s Client) GetNowPlaying() ([]NowPlaying, error) {
	// Retreive all tracks currently playing from Subsonic
	res, err := s.source.Get(s.makeURL("getNowPlaying"))
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
	en := res.Response.NowPlaying.(map[string]interface{})["entry"]
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
			// Artist
			artist, err := ifaceToString(m["artist"])
			if err != nil {
				return nil, err
			}

			// Album
			album, err := ifaceToString(m["album"])
			if err != nil {
				return nil, err
			}

			// Title
			title, err := ifaceToString(m["title"])
			if err != nil {
				return nil, err
			}

			// MusicID
			_musicID, err := strconv.Atoi(m["id"].(string))
			if err != nil {
				return nil, err
			}
			musicID := int64(_musicID)

			// AlbumID
			_albumID, err := strconv.Atoi(m["albumId"].(string))
			if err != nil {
				return nil, err
			}
			albumID := int64(_albumID)

			// Parent
			_parent, err := strconv.Atoi(m["parent"].(string))
			if err != nil {
				return nil, err
			}
			parent := int64(_parent)

			// Create a now playing entry from the map
			n := NowPlaying{
				ID:          musicID,
				AlbumID:     albumID,
				Album:       album,
				Artist:      artist,
				BitRate:     int64(m["bitRate"].(float64)),
				ContentType: m["contentType"].(string),
				CreatedRaw:  m["created"].(string),
				DiscNumber:  int64(m["discNumber"].(float64)),
				DurationRaw: int64(m["duration"].(float64)),
				Genre:       m["genre"].(string),
				IsDir:       m["isDir"].(bool),
				MinutesAgo:  int64(m["minutesAgo"].(float64)),
				Parent:      parent,
				Path:        m["path"].(string),
				PlayerID:    int64(m["playerId"].(float64)),
				Size:        int64(m["size"].(float64)),
				Suffix:      m["suffix"].(string),
				Title:       title,
				Track:       int64(m["track"].(float64)),
				Year:        int64(m["year"].(float64)),
			}

			// Some albums may not have cover art, so we check individually for it
			if c, ok := m["coverArt"].(float64); ok {
				n.CoverArt = int64(c)
			}

			// Parse CreatedRaw into a time.Time struct
			t, err := time.Parse("2006-01-02T15:04:05Z", n.CreatedRaw)
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

// GetCoverArt returns a io.ReadCloser which contains a cover art stream, scaled to the specified size
func (s Client) GetCoverArt(id int64, size int64) (io.ReadCloser, error) {
	// Check for a non-negative size for image scaling
	optStr := ""
	if size > 0 {
		optStr = optStr + "&size=" + strconv.FormatInt(size, 10)
	}

	return fetchBinary(s.makeURL("getCoverArt") + "&id=" + strconv.FormatInt(id, 10) + optStr)
}

// -- Media annotation --

// Scrobble triggers a "Now Playing" or "Submission" request to Last.fm, if configured
func (s Client) Scrobble(id int64, time int64, submission bool) error {
	// Build query string
	optStr := ""

	// time (time < 0 means no time)
	if time > 0 {
		optStr = optStr + "&time=" + strconv.FormatInt(time, 10)
	}

	// submission (true: Submission, false: NowPlaying)
	if submission {
		optStr = optStr + "&submission=true"
	} else {
		optStr = optStr + "&submission=false"
	}

	// Send a scrobble request to Subsonic
	_, err := s.source.Get(s.makeURL("scrobble") + "&id=" + strconv.FormatInt(id, 10) + optStr)
	return err
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

// httpDataSource represents a HTTP data source for a Subsonic client
type httpDataSource struct {
}

// Get retrieves JSON from HTTP with a specified URL, and parses it into an apiContainer
func (s httpDataSource) Get(url string) (*apiContainer, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("gosubsonic: HTTP request failed: %s - %s", err.Error(), url)
	}

	// Read the entire response body
	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Close response body
	if err := res.Body.Close(); err != nil {
		return nil, err
	}

	// Return apiContainer
	return processJSON(out)
}

// mockDataSource represents a mock data source for a Subsonic client
type mockDataSource struct {
}

// Get retrieves JSON from mock data with a specified URL, and parses it into an apiContainer
func (s mockDataSource) Get(url string) (*apiContainer, error) {
	// Get mock data from map
	res, ok := mockData[url]
	if !ok {
		return nil, fmt.Errorf("gosubsonic: No mock data: %s", url)
	}

	// Return apiContainer
	return processJSON(res)
}

// processJSON parses raw JSON into an apiContainer
func processJSON(body []byte) (*apiContainer, error) {
	// Unmarshal response JSON from API container
	var subRes apiContainer
	if err := json.Unmarshal(body, &subRes); err != nil {
		return nil, fmt.Errorf("gosubsonic: failed to parse response JSON: %s", err.Error())
	}

	// Check for any errors in response object
	if subRes.Response.Error != (APIError{}) {
		// Report error and code
		return nil, fmt.Errorf("gosubsonic: %d: %s", subRes.Response.Error.Code, subRes.Response.Error.Message)
	}

	// Return the response container
	return &subRes, nil
}

// ifaceToString attempts to convert an interface type to its string representation
func ifaceToString(data interface{}) (string, error) {
	// There are many cases in Subsonic's XML-to-JSON converter fails to properly
	// handle certain conversions properly.  To account for this, this function will
	// turn types into their string representation.

	// Some issues found so far:
	//   - Numeric artist/title (311, etc) will return as a float64
	//   - Boolean artist/title (True, false), etc will return as a boolean
	//   - String artist/title, etc must have HTML unescaped
	switch data.(type) {
	case nil:
		return "", nil
	case bool:
		if data.(bool) {
			return "True", nil
		}

		return "False", nil
	case string:
		return html.UnescapeString(data.(string)), nil
	case float64:
		return strconv.FormatInt(int64(data.(float64)), 10), nil
	default:
		return "", fmt.Errorf("gosubsonic: unknown data type %T for ifaceToString", data)
	}
}
