package main

// TODO(CaptainHayashi): Petition for move to baps3-go

// Feature is the type for known feature flags.
type Feature uint8

const (
	// FtUnknown represents an unknown feature.
	FtUnknown Feature = iota
	// FtFileLoad represents the FileLoad standard feature.
	FtFileLoad
	// FtPlayStop represents the PlayStop standard feature.
	FtPlayStop
	// FtSeek represents the Seek standard feature.
	FtSeek
	// FtEnd represents the End standard feature.
	FtEnd
	// FtTimeReport represents the TimeReport standard feature.
	FtTimeReport
	// FtPlaylist represents the Playlist standard feature.
	FtPlaylist
	// FtPlaylistAutoAdvance represents the Playlist.AutoAdvance feature.
	FtPlaylistAutoAdvance
	// FtPlaylistTextItems represents the Playlist.TextItems feature.
	FtPlaylistTextItems
)
