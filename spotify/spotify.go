package spotify

import (
	"context"
	"time"

	"spotify/config"
	"spotify/socket"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const DefaultPollRate time.Duration = 5

type SpotifyClient struct {
	Socket *socket.Socket[Track]
	Client *spotify.Client

	pollRate    time.Duration
	isConnected bool
}

func New(cfg *config.Config) *SpotifyClient {
	auth := spotifyauth.New(
		spotifyauth.WithClientID(cfg.Spotify.ClientID),
		spotifyauth.WithClientSecret(cfg.Spotify.ClientSecret),
	)
	token, err := auth.RefreshToken(context.Background(), &oauth2.Token{RefreshToken: cfg.Spotify.RefreshToken})
	if err != nil {
		panic(err)
	}

	return &SpotifyClient{
		Client:      spotify.New(auth.Client(context.Background(), token), spotify.WithRetry(true)),
		isConnected: len(token.AccessToken) > 0,
		pollRate:    DefaultPollRate,
		Socket:      nil,
	}
}

func (sc *SpotifyClient) IsConnected() bool {
	return sc.isConnected
}

func (c *SpotifyClient) GetSpotifyStatus() (*Track, error) {
	if now, err := c.GetNowPlaying(false); err != nil {
		return nil, err
	} else {
		if now != nil {
			return now.(*Track), nil
		}
	}

	last, err := c.GetLastPlayed(false)
	if err != nil {
		return nil, err
	}
	return last.(*Track), nil
}

func (c *SpotifyClient) GetNowPlaying(raw bool) (any, error) {
	if now, err := c.Client.PlayerState(context.Background()); err != nil {
		return nil, err
	} else {
		if !raw {
			if now != nil && now.Playing {
				var artists []Artist
				for _, artist := range now.Item.Artists {
					artists = append(artists, Artist{Name: artist.Name, URL: artist.ExternalURLs["spotify"]})
				}
				return &Track{
					ID:        now.Item.ID,
					Title:     now.Item.Name,
					URL:       now.Item.ExternalURLs["spotify"],
					IsPlaying: now.Playing,
					Timestamp: &Timestamp{
						Progress: now.Progress,
						Duration: now.Item.Duration,
					},
					Artists: artists,
					Album: &Album{
						ID:       now.Item.Album.ID,
						ImageURL: now.Item.Album.Images[0].URL,
						Name:     now.Item.Album.Name,
						URL:      now.Item.Album.ExternalURLs["spotify"],
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
			var artists []Artist
			for _, artist := range track.Artists {
				artists = append(artists, Artist{Name: artist.Name, URL: artist.ExternalURLs["spotify"]})
			}
			return &Track{
				ID:        track.ID,
				Title:     track.Name,
				URL:       track.ExternalURLs["spotify"],
				IsPlaying: false,
				PlayedAt:  &last[0].PlayedAt,
				Artists:   artists,
				Album: &Album{
					ID:       track.Album.ID,
					ImageURL: track.Album.Images[0].URL,
					Name:     track.Album.Name,
					URL:      track.Album.ExternalURLs["spotify"],
				},
			}, nil
		}
		return last, nil
	}
}
