package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	TOKEN_ENDPOINT           = "https://accounts.spotify.com/api/token"
	NOW_PLAYING_ENDPOINT     = "https://api.spotify.com/v1/me/player/currently-playing"
	RECENTLY_PLAYED_ENDPOINT = "https://api.spotify.com/v1/me/player/recently-played"
	PLAYER_ENDPOINT          = "https://api.spotify.com/v1/me/player"
)

type SpotifyClient struct {
	refreshToken string
	accessToken  string
	clientID     string
	clientSecret string
}

type Config struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
}

func New(cfg Config) *SpotifyClient {
	return &SpotifyClient{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		accessToken:  "",
		refreshToken: cfg.RefreshToken,
	}
}

func (c *SpotifyClient) GetAccessToken() (*TokenPayload, *ErrorResponse) {
	f := fiber.AcquireArgs()
	f.Set("grant_type", "refresh_token")
	f.Set("refresh_token", c.refreshToken)

	req := fiber.Post(TOKEN_ENDPOINT).Form(f)
	req.Set("Authorization", fmt.Sprintf("Basic %s", c.encodeBase64(fmt.Sprintf("%s:%s", c.clientID, c.clientSecret))))

	code, body, _ := req.Bytes()
	if code >= 400 {
		var errRes ErrorResponse
		if err := json.Unmarshal(body, &errRes); err != nil {
			errRes.Error.Status = 500
			errRes.Error.Message = err.Error()
		}
		return nil, &errRes
	}

	var payload TokenPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, &ErrorResponse{
			Error: &ErrorPayload{
				Status:  500,
				Message: err.Error(),
			},
		}
	}

	c.accessToken = payload.AccessToken
	return &payload, nil
}

func (c *SpotifyClient) GetNowPlaying() (*PlayerState, *ErrorResponse) {
	req := fiber.Get(NOW_PLAYING_ENDPOINT)
	req.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	code, body, _ := req.Bytes()
	if code >= 400 {
		var errRes ErrorResponse
		if err := json.Unmarshal(body, &errRes); err != nil {
			errRes.Error.Status = 500
			errRes.Error.Message = err.Error()
		}
		return nil, &errRes
	} else if code >= 204 {
		return &PlayerState{IsPlaying: false}, nil
	}
	var payload PlayerState
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, &ErrorResponse{
			Error: &ErrorPayload{
				Status:  500,
				Message: err.Error(),
			},
		}
	}
	return &payload, nil
}

func (c *SpotifyClient) GetRecentlyPlayed() (*RecentlyPlayedResponse, *ErrorResponse) {
	req := fiber.Get(RECENTLY_PLAYED_ENDPOINT + "?limit=1")
	req.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	code, body, _ := req.Bytes()
	if code >= 400 {
		var errRes ErrorResponse
		if err := json.Unmarshal(body, &errRes); err != nil {
			errRes.Error.Status = 500
			errRes.Error.Message = err.Error()
		}
		return nil, &errRes
	}
	var payload RecentlyPlayedResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, &ErrorResponse{
			Error: &ErrorPayload{
				Status:  500,
				Message: err.Error(),
			},
		}
	}
	return &payload, nil
}

func (c *SpotifyClient) UpdateAccessTokenAfter(timeout ...int) {
	defaultTimeout := 55
	if len(timeout) > 0 {
		defaultTimeout = timeout[0]
	}
	for {
		if t, err := c.GetAccessToken(); err != nil {
			return
		} else {
			if t.AccessToken != c.accessToken {
				c.accessToken = t.AccessToken
			}
		}
		time.Sleep(time.Minute * time.Duration(defaultTimeout))
	}
}

func (c *SpotifyClient) GetSpotifyStatus() (*SpotifyStatus, *ErrorResponse) {
	now, err := c.GetNowPlaying()
	if err != nil {
		return nil, err
	}

	if now.Item != nil {
		art := &SpotifyStatusArtist{
			Name: "",
			URL:  now.Item.Artists[0].ExternalUrls["spotify"],
		}
		for _, val := range now.Item.Artists {
			if len(art.Name) > 0 {
				art.Name += "; "
			}
			art.Name += val.Name
		}
		return &SpotifyStatus{
			ID:        now.Item.ID,
			Title:     now.Item.Name,
			Artist:    art,
			URL:       now.Item.ExternalUrls["spotify"],
			IsPlaying: now.IsPlaying,
			Album: &SpotifyStatusAlbum{
				Name:   now.Item.Album.Name,
				URL:    now.Item.Album.ExternalUrls["spotify"],
				ArtURL: now.Item.Album.Images[0].URL,
			},
		}, nil
	}

	last, err := c.GetRecentlyPlayed()
	if err != nil {
		return nil, err
	}
	track := last.Items[0].Track
	art := &SpotifyStatusArtist{
		Name: "",
		URL:  track.Artists[0].ExternalUrls["spotify"],
	}
	for _, val := range track.Artists {
		if len(art.Name) > 0 {
			art.Name += "; "
		}
		art.Name += val.Name
	}
	return &SpotifyStatus{
		ID:        track.ID,
		Title:     track.Name,
		Artist:    art,
		URL:       track.ExternalUrls["spotify"],
		IsPlaying: false,
		Timestamp: last.Items[0].PlayedAt,
		Album: &SpotifyStatusAlbum{
			Name:   track.Album.Name,
			URL:    track.Album.ExternalUrls["spotify"],
			ArtURL: track.Album.Images[0].URL,
		},
	}, nil
}

func (c *SpotifyClient) encodeBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
