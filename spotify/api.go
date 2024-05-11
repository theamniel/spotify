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
type SocketData struct {
	Album     *SocketDataAlbum     `json:"album"`
	Artist    *SocketDataArtist    `json:"artist"`
	ID        sm.ID                `json:"id"`
	IsPlaying bool                 `json:"is_playing"`
	PlayedAt  *time.Time           `json:"played_at,omitempty"`
	Timestamp *SocketDataTimestamp `json:"timestamp,omitempty"`
	Title     string               `json:"title"`
	URL       string               `json:"url"`
}

type SocketDataTimestamp struct {
	Progress sm.Numeric `json:"progress"`
	Duration sm.Numeric `json:"duration"`
}

type SocketDataArtist struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type SocketDataAlbum struct {
	ArtURL string `json:"art_url"`
	Name   string `json:"name"`
	URL    string `json:"url"`
}
