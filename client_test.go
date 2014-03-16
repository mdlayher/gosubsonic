package gosubsonic

import (
	"log"
	"testing"
)

// TestPing verifies that client.Ping() is working properly
func TestPing(t *testing.T) {
	log.Println("TestPing()")

	// Generate mock client
	s, err := New("__MOCK__", "", "")
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

// TestGetLicense verifes that client.GetLicense() is working properly
func TestGetLicense(t *testing.T) {
	log.Println("TestGetLicense()")

	// Generate mock client
	s, err := New("__MOCK__", "", "")
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

// TestGetMusicFolders verifes that client.GetMusicFolders() is working properly
func TestGetMusicFolders(t *testing.T) {
	log.Println("TestGetMusicFolders()")

	// Generate mock client
	s, err := New("__MOCK__", "", "")
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
