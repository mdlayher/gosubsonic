package gosubsonic

// mockData maps a mock URL to mock data from the mockTable
var mockData map[string][]byte

// mockTable maps a method to mock JSON data for testing
var mockTable = []struct {
	method string
	data   []byte
}{
	{"ping", []byte(`{"subsonic-response":{
		"status": "ok",
		"xmlns": "http://subsonic.org/restapi",
		"version": "1.9.0"
	}}`)},
	{"getLicense", []byte(`{"subsonic-response": {
		"status": "ok",
		"xmlns": "http://subsonic.org/restapi",
		"license": {
			"valid": true,
			"email": "mock@example.com",
			"date": "2014-01-01T00:00:00",
			"key": "abcdef0123456789abcdef0123456789"
		},
		"version": "1.9.0"
	}}`)},
	{"getMusicFolders", []byte(`{"subsonic-response": {
		"status": "ok",
		"xmlns": "http://subsonic.org/restapi",
		"musicFolders": {"musicFolder": {
			"id": 0,
			"name": "Music"
		}},
		"version": "1.9.0"
	}}`)},
	{"getIndexes", []byte(`{"subsonic-response": {
		"status": "ok",
		"indexes": {
			"index": [{
				"name": "A",
				"artist": {
					"id": 1,
					"name": "Adventure"
				}
			},
			{
				"name": "B",
				"artist": {
					"id": 2,
					"name": "Boston"
				}
			}],
			"lastModified": 1395014311154
		},
		"xmlns": "http://subsonic.org/restapi",
		"version": "1.9.0"
	}}`)},
	{"getMusicDirectory", []byte(`{"subsonic-response": {
		"status": "ok",
		"directory": {
			"child": {
				"id": 405,
				"title": "2008 - Adventure",
				"created": "2013-08-12T00:12:24",
				"album": "Adventure",
				"parent": 1,
				"isDir": true,
				"artist": "Adventure",
				"coverArt": 405
			},
		"id": 3,
		"name": "Adventure"
		},
		"xmlns": "http://subsonic.org/restapi",
		"version": "1.9.0"
	}}`)},
}

// mockInit generates the mock data map, so we can test gosubsonic against known, static data
func mockInit(s Client) error {
	// Initialize map
	mockData = map[string][]byte{}

	// Populate map using this client's URLs
	for _, entry := range mockTable {
		// Extra options
		optStr := ""

		// getMusicDirectory - add mock ID
		if entry.method == "getMusicDirectory" {
			optStr = optStr + "&id=1"
		}

		mockData[s.makeURL(entry.method) + optStr] = entry.data
	}

	return nil
}
