package spotify

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	TOKEN_ENDPOINT           = "https://accounts.spotify.com/api/token"
	NOW_PLAYING_ENDPOINT     = "https://api.spotify.com/v1/me/player/currently-playing"
	RECENTLY_PLAYED_ENDPOINT = "https://api.spotify.com/v1/me/player/recently-played"
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

func (c *SpotifyClient) GetAccessToken() (string, error) {
	f := fiber.AcquireArgs()
	f.Set("grant_type", "refresh_token")
	f.Set("refresh_token", c.refreshToken)

	req := fiber.Post(TOKEN_ENDPOINT).Form(f)
	req.Set("Authorization", fmt.Sprintf("Basic %s", c.encodeBase64(fmt.Sprintf("%s:%s", c.clientID, c.clientSecret))))

	payload := TokenPayload{}
	if code, _, _ := req.Struct(&payload); code != 200 {
		return "", errors.New("Errors unknown")
	}

	c.accessToken = payload.AccessToken
	return payload.AccessToken, nil
}

func (c *SpotifyClient) GetNowPlaying() (*TrackPayload, error) {
	req := fiber.Get(NOW_PLAYING_ENDPOINT)
	req.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	payload := TrackPayload{}
	code, body, _ := req.String()
	if code >= 204 {
		payload.IsPlaying = false
		return &payload, nil
	}
	fmt.Println(body)

	return &payload, nil
}

func (c *SpotifyClient) GetRecentlyPlayed() (string, error) {
	if len(c.accessToken) < 1 {
		_, err := c.GetAccessToken()
		if err != nil {
			return "", err
		}
	}

	req := fiber.Get(RECENTLY_PLAYED_ENDPOINT)
	req.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	_, body, _ := req.String()
	return body, nil
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
			if t != c.accessToken {
				c.accessToken = t
			}
		}
		time.Sleep(time.Second * time.Duration(defaultTimeout))
	}
}

func (c *SpotifyClient) encodeBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
