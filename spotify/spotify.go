package spotify

import (
	"context"
	"time"

	"github.com/theamniel/spotify-server/config"
	"github.com/theamniel/spotify-server/socket"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type SpotifyClient struct {
	Socket *socket.Socket[SocketData]
	Client *spotify.Client

	pollRate    time.Duration
	isConnected bool
}

func New(cfg *config.SpotifyConfig) *SpotifyClient {
	auth := spotifyauth.New(
		spotifyauth.WithClientID(cfg.ClientID),
		spotifyauth.WithClientSecret(cfg.ClientSecret),
	)
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

func (c *SpotifyClient) GetSpotifyStatus() (*SocketData, error) {
	if now, err := c.GetNowPlaying(false); err != nil {
		return nil, err
	} else {
		if now != nil {
			return now.(*SocketData), nil
		}
	}

	last, err := c.GetLastPlayed(false)
	if err != nil {
		return nil, err
	}
	return last.(*SocketData), nil
}

func (c *SpotifyClient) GetNowPlaying(raw bool) (any, error) {
	if now, err := c.Client.PlayerState(context.Background()); err != nil {
		return nil, err
	} else {
		if !raw {
			if now != nil && now.Playing {
				var artists []SocketDataArtist
				for _, artist := range now.Item.Artists {
					artists = append(artists, SocketDataArtist{Name: artist.Name, URL: artist.ExternalURLs["spotify"]})
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
					Artists: artists,
					Album: &SocketDataAlbum{
						Name:   now.Item.Album.Name,
						URL:    now.Item.Album.ExternalURLs["spotify"],
						ArtURL: now.Item.Album.Images[0].URL,
					},
				}, nil
			}
			return nil, nil
		} else {
			rawData, err := c.Client.PlayerCurrentlyPlaying(context.Background())
			return rawData, err
		}
	}
}

func (c *SpotifyClient) GetLastPlayed(raw bool) (any, error) {
	if last, err := c.Client.PlayerRecentlyPlayed(context.Background()); err != nil {
		return nil, err
	} else {
		if !raw {
			track := last[0].Track
			var artists []SocketDataArtist
			for _, artist := range track.Artists {
				artists = append(artists, SocketDataArtist{Name: artist.Name, URL: artist.ExternalURLs["spotify"]})
			}
			return &SocketData{
				ID:        track.ID,
				Title:     track.Name,
				URL:       track.ExternalURLs["spotify"],
				IsPlaying: false,
				PlayedAt:  &last[0].PlayedAt,
				Artists:   artists,
				Album: &SocketDataAlbum{
					Name:   track.Album.Name,
					URL:    track.Album.ExternalURLs["spotify"],
					ArtURL: track.Album.Images[0].URL,
				},
			}, nil
		}
		return last, nil
	}
}
