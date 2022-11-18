package spotify

type TokenPayload struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type ErrorPayload struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error *ErrorPayload `json:"error"`
}

type SpotifyStatus struct {
	ID        string               `json:"id"`
	Title     string               `json:"title"`
	Artist    *SpotifyStatusArtist `json:"artist"`
	URL       string               `json:"url"`
	IsPlaying bool                 `json:"is_playing"`
	Album     *SpotifyStatusAlbum  `json:"album"`
	Timestamp string               `json:"timestamp,omitempty"`
}

type SpotifyStatusArtist struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type SpotifyStatusAlbum struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	ArtURL string `json:"art_url"`
}

type PlayerState struct {
	IsPlaying  bool   `json:"is_playing"`
	ProgressMs int    `json:"progress_ms"`
	Item       *Track `json:"item"`
	Timestamp  int    `json:"timestamp,omitempty"`
}

type RecentlyPlayedResponse struct {
	Items []RecentlyPlayedTrack
}

type RecentlyPlayedTrack struct {
	Track    *Track `json:"track"`
	PlayedAt string `json:"played_at"`
}

type Album struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Images       []Image           `json:"images"`
	ExternalUrls map[string]string `json:"external_urls"`
}

type Image struct {
	Height int    `json:"height"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
}

type Track struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	DurationMs   int               `json:"duration_ms"`
	Album        *Album            `json:"album"`
	Artists      []Artist          `json:"artists"`
	ExternalUrls map[string]string `json:"external_urls"`
}

type Artist struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	ExternalUrls map[string]string `json:"external_urls"`
}
