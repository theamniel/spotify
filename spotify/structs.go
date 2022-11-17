package spotify

type TokenPayload struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type TrackPayload struct {
	IsPlaying bool `json:"is_playing"`
	Progress  int  `json:"progress_ms"`
	Track     `json:"item"`
}

type Track struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Url      string   `json:"uri"`
	Duration int      `json:"duration_ms"`
	Artists  []Artist `json:"artists"`
}

type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
