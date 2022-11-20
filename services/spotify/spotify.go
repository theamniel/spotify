package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
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
	Socket       *SpotifySocket
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
		Socket:       nil,
	}
}

func (c *SpotifyClient) GetAccessToken() (*Token, *SpotifyApiError) {
	f := fiber.AcquireArgs()
	f.Set("grant_type", "refresh_token")
	f.Set("refresh_token", c.refreshToken)

	req := fiber.Post(TOKEN_ENDPOINT).Form(f)
	req.Set("Authorization", fmt.Sprintf("Basic %s", c.encodeBase64(fmt.Sprintf("%s:%s", c.clientID, c.clientSecret))))

	code, body, _ := req.Bytes()
	if code >= 400 {
		return nil, NewApiErrorFrom(body)
	}

	var payload Token
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, NewApiError(500, err.Error())
	}

	c.accessToken = payload.AccessToken
	return &payload, nil
}

func (c *SpotifyClient) GetNowPlaying() (*PlayerState, *SpotifyApiError) {
	req := fiber.Get(NOW_PLAYING_ENDPOINT)
	req.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	code, body, _ := req.Bytes()
	if code >= 400 {
		return nil, NewApiErrorFrom(body)
	} else if code >= 204 {
		return &PlayerState{IsPlaying: false}, nil
	}
	var payload PlayerState
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, NewApiError(500, err.Error())
	}
	return &payload, nil
}

func (c *SpotifyClient) GetRecentlyPlayed() (*TracksPaged, *SpotifyApiError) {
	req := fiber.Get(RECENTLY_PLAYED_ENDPOINT + "?limit=1")
	req.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	code, body, _ := req.Bytes()
	if code >= 400 {
		return nil, NewApiErrorFrom(body)
	}
	var payload TracksPaged
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, NewApiError(500, err.Error())
	}
	return &payload, nil
}

func (c *SpotifyClient) UpdateAccessTokenAfter() {
	for {
		if t, err := c.GetAccessToken(); err != nil {
			break
		} else {
			if t.AccessToken != c.accessToken {
				c.accessToken = t.AccessToken
			}
		}
		time.Sleep(55 * time.Minute) // by default, access token expires in 1 hour
	}
}

func (c *SpotifyClient) GetSpotifyStatus() (*SocketData, *SpotifyApiError) {
	now, err := c.GetNowPlaying()
	if err != nil {
		return nil, err
	}

	if now.Item != nil {
		var artsName []string
		for _, val := range now.Item.Artists {
			artsName = append(artsName, val.Name)
		}
		return &SocketData{
			ID:        now.Item.ID,
			Title:     now.Item.Name,
			URL:       now.Item.ExternalUrls["spotify"],
			IsPlaying: now.IsPlaying,
			Duration: &SocketDataDuration{
				Start: now.ProgressMs,
				End:   now.Item.DurationMs,
			},
			Artist: &SocketDataArtist{
				Name: strings.Join(artsName, "; "),
				URL:  now.Item.Artists[0].ExternalUrls["spotify"],
			},
			Album: &SocketDataAlbum{
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
	var artsName []string
	for _, val := range track.Artists {
		artsName = append(artsName, val.Name)
	}
	return &SocketData{
		ID:        track.ID,
		Title:     track.Name,
		URL:       track.ExternalUrls["spotify"],
		IsPlaying: false,
		Timestamp: last.Items[0].PlayedAt,
		Artist: &SocketDataArtist{
			Name: strings.Join(artsName, "; "),
			URL:  track.Artists[0].ExternalUrls["spotify"],
		},
		Album: &SocketDataAlbum{
			Name:   track.Album.Name,
			URL:    track.Album.ExternalUrls["spotify"],
			ArtURL: track.Album.Images[0].URL,
		},
	}, nil
}

func (c *SpotifyClient) encodeBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
