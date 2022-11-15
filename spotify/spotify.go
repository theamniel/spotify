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

type Client struct {
	refreshToken string
	accessToken  string
	clientID     string
	clientSecret string
	socket       *ClientSocket
}

type Config struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
}

type AccessTokenPayload struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func New(cfg Config) *Client {
	return &Client{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		accessToken:  "",
		refreshToken: cfg.RefreshToken,
		socket: &ClientSocket{
			send: make(chan *SocketEvent),
			on:   make(<-chan []byte),
		},
	}
}

func (c *Client) GetAccessToken() (string, error) {
	f := fiber.AcquireArgs()
	f.Set("grant_type", "refresh_token")
	f.Set("refresh_token", c.refreshToken)

	req := fiber.Post(TOKEN_ENDPOINT).Form(f)
	req.Set("Authorization", fmt.Sprintf("Basic %s", c.encodeBase64(fmt.Sprintf("%s:%s", c.clientID, c.clientSecret))))

	var payload AccessTokenPayload
	if code, _, _ := req.Struct(&payload); code != 200 {
		return "", errors.New("Errors unknown")
	}

	c.accessToken = payload.AccessToken
	return payload.AccessToken, nil
}

func (c *Client) GetNowPlaying() (string, error) {
	if len(c.accessToken) < 1 {
		_, err := c.GetAccessToken()
		if err != nil {
			return "", err
		}
	}

	req := fiber.Get(NOW_PLAYING_ENDPOINT)
	req.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	_, body, _ := req.String()
	return body, nil
}

func (c *Client) GetRecentlyPlayed() (string, error) {
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

func (c *Client) UpdateAccessTokenAfter(timeout int) {
	for {
		time.Sleep(time.Second * time.Duration(timeout))
		if t, err := c.GetAccessToken(); err != nil {
			return
		} else {
			if t != c.accessToken {
				c.accessToken = t
			}
		}
	}
}

func (c *Client) encodeBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
