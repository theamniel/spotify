package spotify

import "time"

// The Token struct describes a item returned from Spotify's API Authorization process
type Token struct {
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
	ExpiresIn   time.Duration `json:"expires_in"`
	Scope       string        `json:"scope"`
}

// The Image struct describes an album, artist, playlist, etc image
type Image struct {
	Height int    `json:"height"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
}

// The Followers struct describes the followers for an artist, playlist, etc
type Followers struct {
	Href  string `json:"href"`
	Total int    `json:"total"`
}

/*--------------- ALBUM --------------*/
// The Album struct describes a "Simple" Album object as defined by the Spotify Web API
type Album struct {
	AlbumType        string            `json:"album"`
	Artists          []Artist          `json:"artists"`
	AvailableMarkets []string          `json:"available_markets"`
	ExternalUrls     map[string]string `json:"external_urls"`
	Href             string            `json:"href"`
	ID               string            `json:"id"`
	Images           []Image           `json:"images"`
	Name             string            `json:"name"`
	Type             string            `json:"type"`
	URI              string            `json:"uri"`
}

/*------------- ARTIST -------------*/
// The Artist struct describes a "Simple" Artist object as defined by the Spotify Web API
type Artist struct {
	ExternalUrls map[string]string `json:"external_urls"`
	Href         string            `json:"href"`
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	URI          string            `json:"uri"`
}

/*--------------- TRACK ---------------*/
// The TrackLink struct describes a TrackLink object as defined by the Spotify Web API
type TrackLink struct {
	ExternalUrls map[string]string `json:"external_urls"`
	Href         string            `json:"href"`
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	URI          string            `json:"uri"`
}

// The Track struct describes a Track object as defined by the Spotify Web API
type Track struct {
	Album            *Album            `json:"album"`
	Artists          []Artist          `json:"artists"`
	AvailableMarkets []string          `json:"available_markets"`
	DiscNumber       int               `json:"disc_number"`
	DurationMs       int               `json:"duration_ms"`
	Explicit         bool              `json:"explicit"`
	ExternalUrls     map[string]string `json:"external_urls"`
	Href             string            `json:"href"`
	ID               string            `json:"id"`
	IsPlayable       bool              `json:"is_playable"`
	LinkedFrom       *TrackLink        `json:"linked_from"`
	Name             string            `json:"name"`
	PreviewURL       string            `json:"preview_url"`
	TrackNumber      int               `json:"track_number"`
	Type             string            `json:"type"`
	URI              string            `json:"uri"`
}

// The TracksPaged struct is a slice of Track objects wrapped in a Spotify paging object
type TracksPaged struct {
	Href     string       `json:"href"`
	Items    []TrackPaged `json:"items"`
	Limit    int          `json:"limit"`
	Next     string       `json:"next"`
	Offset   int          `json:"offset"`
	Previous string       `json:"previous"`
	Total    int          `json:"total"`
}

type TrackPaged struct {
	Track    *Track `json:"track"`
	PlayedAt string `json:"played_at"`
}

/*----------- PLAYER ------------*/
// The PlayerContext struct describes the current context of what is playing on the active device.
type PlayerContext struct {
	Type         string            `json:"type"`
	Href         string            `json:"href"`
	ExternalUrls map[string]string `json:"external_urls"`
	URI          string            `json:"uri"`
}

// The PlayerState struct describes the current playback state of Spotify
type PlayerState struct {
	Device       *Device        `json:"device"`
	RepeatState  string         `json:"repeat_state"`
	ShuffleState bool           `json:"shuffle_state"`
	Context      *PlayerContext `json:"context"`
	Timestamp    int            `json:"timestamp"`
	ProgressMs   int            `json:"progress_ms"`
	IsPlaying    bool           `json:"is_playing"`
	Item         *Track         `json:"item"`
}

// The Device struct describes an available playback device
type Device struct {
	ID            string `json:"id"`
	IsActive      bool   `json:"is_active"`
	IsRestricted  bool   `json:"is_restricted"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	VolumePercent int    `json:"volume_percent"`
}

/*-------------- SOCKET API ------------*/
type SocketData struct {
	Album     *SocketDataAlbum     `json:"album"`
	Artist    *SocketDataArtist    `json:"artist"`
	ID        string               `json:"id"`
	IsPlaying bool                 `json:"is_playing"`
	PlayedAt  string               `json:"played_at,omitempty"`
	Timestamp *SocketDataTimestamp `json:"timestamp,omitempty"`
	Title     string               `json:"title"`
	URL       string               `json:"url"`
}

type SocketDataTimestamp struct {
	Progress int `json:"progress"`
	Duration int `json:"duration"`
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
