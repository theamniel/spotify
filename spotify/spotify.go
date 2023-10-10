package spotify

import (
	"context"
	"strings"
	"time"

	"github.com/theamniel/spotify-server/config"
	"github.com/theamniel/spotify-server/socket"
	"github.com/zmb3/spotify/v2"
	"github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type SpotifyClient struct {
	Socket *socket.Socket[SocketData]
	Client *spotify.Client

	pollRate    time.Duration
	isConnected bool
}

func New(cfg *config.SpotifyConfig) *SpotifyClient {
	auth := spotifyauth.New(spotifyauth.WithClientID(cfg.ClientID), spotifyauth.WithClientSecret(cfg.ClientSecret))
	token, err := auth.RefreshToken(context.Background(), &oauth2.Token{RefreshToken: cfg.RefreshToken})
	if err != nil {
		panic(err)
	}

	return &SpotifyClient{
		Client:      spotify.New(auth.Client(context.Background(), token), spotify.WithRetry(true)),
		isConnected: len(token.AccessToken) > 0,
		pollRate:    5,
		Socket:      nil,
	}
}

func (sc *SpotifyClient) IsConnected() bool {
	return sc.isConnected
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
