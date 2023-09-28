package spotify

import (
	"context"
	"strings"
	"time"

	"github.com/theamniel/spotify-server/config"
	"github.com/theamniel/spotify-server/socket"
	sp "github.com/zmb3/spotify/v2"
	spauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const (
	TOKEN_ENDPOINT           = "https://accounts.spotify.com/api/token"
	NOW_PLAYING_ENDPOINT     = "https://api.spotify.com/v1/me/player/currently-playing"
	RECENTLY_PLAYED_ENDPOINT = "https://api.spotify.com/v1/me/player/recently-played"
	PLAYER_ENDPOINT          = "https://api.spotify.com/v1/me/player"
)

type SpotifyClient struct {
	Socket *socket.Socket[SocketData]
	Client *sp.Client

	pollRate     time.Duration
	refreshToken string
	accessToken  string
	clientID     string
	clientSecret string
}

func New(cfg *config.SpotifyConfig) *SpotifyClient {
	cauth := spauth.New(spauth.WithClientID(cfg.ClientID), spauth.WithClientSecret(cfg.ClientSecret))
	token, err := cauth.RefreshToken(context.Background(), &oauth2.Token{RefreshToken: cfg.RefreshToken})
	if err != nil {
		panic(err)
	}

	return &SpotifyClient{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		accessToken:  token.AccessToken,
		refreshToken: token.RefreshToken,
		Client:       sp.New(cauth.Client(context.Background(), token), sp.WithRetry(true)),
		pollRate:     5,
		Socket:       nil,
	}
}

func (c *SpotifyClient) GetPlayerState() (*PlayerState, error) {
	return c.Client.PlayerState(context.Background())
}

func (c *SpotifyClient) GetNowPlaying() (*CurrentlyPlaying, error) {
	return c.Client.PlayerCurrentlyPlaying(context.Background())
}

func (c *SpotifyClient) GetRecentlyPlayed() ([]RecentlyPlayedItem, error) {
	return c.Client.PlayerRecentlyPlayed(context.Background())
}

func (c *SpotifyClient) GetSpotifyStatus() (*SocketData, error) {
	now, err := c.GetPlayerState()
	if err != nil {
		return nil, err
	}

	if now != nil && now.Playing {
		var artsName []string
		for _, val := range now.Item.Artists {
			artsName = append(artsName, val.Name)
		}
		return &SocketData{
			ID:        now.Item.ID,
			Title:     now.Item.Name,
			URL:       now.Item.ExternalURLs["spotify"],
			IsPlaying: now.Playing,
			Timestamp: &SocketDataTimestamp{
				Progress: now.Progress,
				Duration: now.Item.Duration,
			},
			Artist: &SocketDataArtist{
				Name: strings.Join(artsName, "; "),
				URL:  now.Item.Artists[0].ExternalURLs["spotify"],
			},
			Album: &SocketDataAlbum{
				Name:   now.Item.Album.Name,
				URL:    now.Item.Album.ExternalURLs["spotify"],
				ArtURL: now.Item.Album.Images[0].URL,
			},
		}, nil
	}

	last, err := c.GetRecentlyPlayed()
	if err != nil {
		return nil, err
	}
	track := last[0].Track
	var artsName []string
	for _, val := range track.Artists {
		artsName = append(artsName, val.Name)
	}
	return &SocketData{
		ID:        track.ID,
		Title:     track.Name,
		URL:       track.ExternalURLs["spotify"],
		IsPlaying: false,
		PlayedAt:  &last[0].PlayedAt,
		Artist: &SocketDataArtist{
			Name: strings.Join(artsName, "; "),
			URL:  track.Artists[0].ExternalURLs["spotify"],
		},
		Album: &SocketDataAlbum{
			Name:   track.Album.Name,
			URL:    track.Album.ExternalURLs["spotify"],
			ArtURL: track.Album.Images[0].URL,
		},
	}, nil
}
