package gosubsonic

type Artist struct {
	Id         int
	Album      []Album
	AlbumCount int
	CoverArt   string
	Name       string
}
