package gosubsonic

type Artist struct {
	Id         int
	Album      []Album
	TempAlbum  Album
	AlbumCount int
	CoverArt   string
	Name       string
	Client     SubsonicClient
}

func (a Artist) Albums() ([]Album, error) {
	if len(a.Album) > 0 {
		return a.Album, nil
	}

	// Query for list of all albums for this artist
	a, err := a.Client.GetArtist(a.Id)
	if err != nil {
		return a.Album, err
	}

	// Return list of albums
	return a.Album, nil
}
