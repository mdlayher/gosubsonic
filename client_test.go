package gosubsonic

import (
	"log"
	"testing"
)

// TestPing verifies that client.Ping() is working properly
func TestPing(t *testing.T) {
	log.Println("TestPing()")

	// Generate mock client
	s, err := NewMock()
	if err != nil {
		t.Fatalf("Could not generate mock client: %s", err.Error())
	}

	// Ping mock data and get current status
	stat, err := s.Ping()
	if err != nil {
		t.Fatalf("Ping returned error: %s", err.Error())
	}

	// Check for "ok"
	if stat.Status != "ok" {
		t.Fatalf("Ping returned bad status: %s", stat.Status)
	}

	// Check for proper Subsonic xmlns
	if stat.Xmlns != "http://subsonic.org/restapi" {
		t.Fatalf("Ping returned bad xmlns: %s", stat.Xmlns)
	}
}

// TestGetLicense verifies that client.GetLicense() is working properly
func TestGetLicense(t *testing.T) {
	log.Println("TestGetLicense()")

	// Generate mock client
	s, err := NewMock()
	if err != nil {
		t.Fatalf("Could not generate mock client: %s", err.Error())
	}

	// Get license mock data
	license, err := s.GetLicense()
	if err != nil {
		t.Fatalf("GetLicense returned error: %s", err.Error())
	}

	// Check for valid license
	if !license.Valid {
		t.Fatalf("GetLicense returned invalid license")
	}

	// Check for invalid "zero" date
	if license.Date.IsZero() {
		t.Fatalf("GetLicense returned zero date")
	}
}

// TestGetMusicFolders verifies that client.GetMusicFolders() is working properly
func TestGetMusicFolders(t *testing.T) {
	log.Println("TestGetMusicFolders()")

	// Generate mock client
	s, err := NewMock()
	if err != nil {
		t.Fatalf("Could not generate mock client: %s", err.Error())
	}

	// Get music folders mock data
	folders, err := s.GetMusicFolders()
	if err != nil {
		t.Fatalf("GetMusicFolders returned error: %s", err.Error())
	}

	// Check for known ID
	if folders[0].ID != 0 {
		t.Fatalf("GetMusicFolders returned invalid ID: %d", folders[0].ID)
	}

	// Check for known name
	if folders[0].Name != "Music" {
		t.Fatalf("GetMusicFolders returned invalid name: %s", folders[0].Name)
	}
}

// TestGetIndexes verifies that client.GetIndexes() is working properly
func TestGetIndexes(t *testing.T) {
	log.Println("TestGetIndexes()")

	// Generate mock client
	s, err := NewMock()
	if err != nil {
		t.Fatalf("Could not generate mock client: %s", err.Error())
	}

	// Get indexes mock data
	indexes, err := s.GetIndexes(-1, -1)
	if err != nil {
		t.Fatalf("GetIndexes returned error: %s", err.Error())
	}

	// Check for proper index
	if indexes[0].Name != "A" {
		t.Fatalf("GetIndexes returned invalid index name: %s", indexes[0].Name)
	}

	// Check for known ID
	if indexes[0].Artist[0].ID != 1 {
		t.Fatalf("GetIndexes returned invalid ID: %d", indexes[0].Artist[0].ID)
	}

	// Check for known name
	if indexes[1].Artist[0].Name != "Boston" {
		t.Fatalf("GetIndexes returned invalid name: %s", indexes[1].Artist[0].Name)
	}
}

// TestGetMusicDirectory verifies that client.GetMusicDirectory() is working properly
func TestGetMusicDirectory(t *testing.T) {
	log.Println("TestGetMusicDirectory()")

	// Generate mock client
	s, err := NewMock()
	if err != nil {
		t.Fatalf("Could not generate mock client: %s", err.Error())
	}

	// Get music directory mock data
	content, err := s.GetMusicDirectory(1)
	if err != nil {
		t.Fatalf("GetMusicDirectory returned error: %s", err.Error())
	}

	// Check for mock directory ID
	if content.Directories[0].ID != 405 {
		t.Fatalf("GetMusicDirectory returned invalid ID: %d", content.Directories[0].ID)
	}

	// Check for mock artist
	if content.Directories[0].Artist != "Adventure" {
		t.Fatalf("GetMusicDirectory returned invalid artist: %s", content.Directories[0].Artist)
	}
}

// TestScrobble verifies that client.Scrobble() is working properly
func TestScrobble(t *testing.T) {
	log.Println("TestScrobble()")

	// Generate mock client
	s, err := NewMock()
	if err != nil {
		t.Fatalf("Could not generate mock client: %s", err.Error())
	}

	// Get scrobble mock data
	if err := s.Scrobble(1, -1, false); err != nil {
		t.Fatalf("Scrobble returned error: %s", err.Error())
	}
}
