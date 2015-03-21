package main

// TODO(CaptainHayashi): Petition for move to baps3-go

// Feature is the type for known feature flags.
type Feature uint8

const (
	// FileLoad represents the FileLoad standard feature.
	FileLoad Feature = iota
	// PlayStop represents the PlayStop standard feature.
	PlayStop
	// Seek represents the Seek standard feature.
	Seek
	// End represents the End standard feature.
	End
	// TimeReport represents the TimeReport standard feature.
	TimeReport
	// Playlist represents the Playlist standard feature.
	Playlist
	// PlaylistAutoAdvance represents the Playlist.AutoAdvance feature.
	PlaylistAutoAdvance
	// PlaylistTextItems represents the Playlist.TextItems feature.
	PlaylistTextItems
)
