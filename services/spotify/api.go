package spotify

import (
	"time"

	proto "spotify/protocols"

	sm "github.com/zmb3/spotify/v2"
)

// Aliases for Spotify types
type (
	PlayerState        = sm.PlayerState
	CurrentlyPlaying   = sm.CurrentlyPlaying
	RecentlyPlayedItem = sm.RecentlyPlayedItem
)

/*-------------- SOCKET API ------------*/

// Track represents a Spotify track
type Track struct {
	// Album information
	Album *Album `json:"album"`
	// Artists involved in the track
	Artists []Artist `json:"artists"`
	// Spotify ID of the track
	ID sm.ID `json:"id"`
	// Whether the track is currently playing
	IsPlaying bool `json:"is_playing"`
	// Timestamp when the track was played
	PlayedAt *time.Time `json:"played_at,omitempty"`
	// timestamp information
	Timestamp *Timestamp `json:"timestamp,omitempty"`
	// Track title
	Title string `json:"title"`
	// URL of the track
	URL string `json:"url"`
}

func (track *Track) ToProto() *proto.Track {
	var playedAt *int64 = nil
	if track.PlayedAt != nil {
		played := track.PlayedAt.UnixMilli()
		playedAt = &played
	}
	artists := make([]*proto.Artist, len(track.Artists))
	for i, artist := range track.Artists {
		artists[i] = artist.ToProto()
	}

	return &proto.Track{
		Album:     track.Album.ToProto(),
		Artist:    artists,
		ID:        track.ID.String(),
		IsPlaying: track.IsPlaying,
		PlayedAt:  playedAt,
		Timestamp: track.Timestamp.ToProto(),
		Title:     track.Title,
		URL:       track.URL,
	}
}

func FromProtoToTrack(pb *proto.Track) *Track {
	artists := make([]Artist, len(pb.Artist))
	playedAt := &time.Time{}
	if pb.PlayedAt != nil {
		ti := time.UnixMilli(*pb.PlayedAt)
		playedAt = &ti
	} else {
		playedAt = nil
	}
	for i, artist := range pb.Artist {
		artists[i] = Artist{Name: artist.Name, URL: artist.URL}
	}
	return &Track{
		Album: &Album{
			ID:       sm.ID(pb.Album.ID),
			Name:     pb.Album.Name,
			URL:      pb.Album.URL,
			ImageURL: pb.Album.ImageURL,
		},
		Artists:   artists,
		ID:        sm.ID(pb.ID),
		IsPlaying: pb.IsPlaying,
		PlayedAt:  playedAt,
		Timestamp: FromProtoToTimestamp(pb.Timestamp),
		Title:     pb.Title,
		URL:       pb.URL,
	}
}

// Timestamp represents timestamp information for a track
type Timestamp struct {
	// Progress of the track in milliseconds
	Progress sm.Numeric `json:"progress"`
	// Duration of the track in milliseconds
	Duration sm.Numeric `json:"duration"`
}

func (timestamp *Timestamp) ToProto() *proto.Timestamp {
	if timestamp == nil {
		return nil
	}
	return &proto.Timestamp{
		Progress: int64(timestamp.Progress),
		Duration: int64(timestamp.Duration),
	}
}

func FromProtoToTimestamp(pb *proto.Timestamp) *Timestamp {
	if pb == nil || (pb.Duration == 0 && pb.Progress == 0) {
		return nil
	}
	return &Timestamp{
		Duration: sm.Numeric(pb.Duration),
		Progress: sm.Numeric(pb.Progress),
	}
}

// Artist represents an artist in volved in a track
type Artist struct {
	// Artist name
	Name string `json:"name"`
	// URL of the artist
	URL string `json:"url"`
}

func (artist *Artist) ToProto() *proto.Artist {
	return &proto.Artist{
		Name: artist.Name,
		URL:  artist.URL,
	}
}

// Album represents an album associated with a track
type Album struct {
	// URL of the album image
	ImageURL string `json:"image_url"`
	// Name of the album
	Name string `json:"name"`
	// Spotify ID of the album
	ID sm.ID `json:"id"`
	// URL of the album
	URL string `json:"url"`
}

func (album *Album) ToProto() *proto.Album {
	return &proto.Album{
		ImageURL: album.ImageURL,
		Name:     album.Name,
		ID:       album.ID.String(),
		URL:      album.URL,
	}
}
