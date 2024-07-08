package spotify

import (
	"time"

	sm "github.com/zmb3/spotify/v2"
)

// Aliases for Spotify types
type (
	PlayerState        = sm.PlayerState
	CurrentlyPlaying   = sm.CurrentlyPlaying
	RecentlyPlayedItem = sm.RecentlyPlayedItem
)

/*-------------- SOCKET API ------------*/

// Track represents a Spotify track
type Track struct {
	// Album information
	Album *Album `json:"album"`
	// Artists involved in the track
	Artists []Artist `json:"artists"`
	// Spotify ID of the track
	ID sm.ID `json:"id"`
	// Whether the track is currently playing
	IsPlaying bool `json:"is_playing"`
	// Timestamp when the track was played
	PlayedAt *time.Time `json:"played_at,omitempty"`
	// timestamp information
	Timestamp *Timestamp `json:"timestamp,omitempty"`
	// Track title
	Title string `json:"title"`
	// URL of the track
	URL string `json:"url"`
}

// Timestamp represents timestamp information for a track
type Timestamp struct {
	// Progress of the track in milliseconds
	Progress sm.Numeric `json:"progress"`
	// Duration of the track in milliseconds
	Duration sm.Numeric `json:"duration"`
}

// Artist represents an artist in volved in a track
type Artist struct {
	// Artist name
	Name string `json:"name"`
	// URL of the artist
	URL string `json:"url"`
}

// Album represents an album associated with a track
type Album struct {
	// URL of the album image
	ImageURL string `json:"image_url"`
	// Name of the album
	Name string `json:"name"`
	// Spotify ID of the album
	ID sm.ID `json:"id"`
	// URL of the album
	URL string `json:"url"`
}
