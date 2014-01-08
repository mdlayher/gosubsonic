package gosubsonic

type Song struct {
	Id          int
	Album       string
	AlbumId     int
	Artist      string
	ArtistId    int
	BitRate     int
	ContentType string
	CoverArt    int
	Created     string
	DiscNumber  int
	Duration    int
	Genre       string
	IsDir       bool
	IsVideo     bool
	Parent      int
	Path        string
	Size        int64
	Suffix      string
	Title       string
	Track       int
	Type        string
	Year        int
}
