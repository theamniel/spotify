package spotify

import (
	"time"

	sm "github.com/zmb3/spotify/v2"
)

type (
	PlayerState        = sm.PlayerState
	CurrentlyPlaying   = sm.CurrentlyPlaying
	RecentlyPlayedItem = sm.RecentlyPlayedItem
)

/*-------------- SOCKET API ------------*/
type Track struct {
	Album     *TrackAlbum     `json:"album"`
	Artists   []TrackArtist   `json:"artists"`
	ID        sm.ID           `json:"id"`
	IsPlaying bool            `json:"is_playing"`
	PlayedAt  *time.Time      `json:"played_at,omitempty"`
	Timestamp *TrackTimestamp `json:"timestamp,omitempty"`
	Title     string          `json:"title"`
	URL       string          `json:"url"`
}

type TrackTimestamp struct {
	Progress sm.Numeric `json:"progress"`
	Duration sm.Numeric `json:"duration"`
}

type TrackArtist struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type TrackAlbum struct {
	ImageURL string `json:"image_url"`
	Name     string `json:"name"`
	URL      string `json:"url"`
}
