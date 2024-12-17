package main

import (
	"time"

	"spotify/services/spotify"
)

func (s *server) pool(id string, onData func(*spotify.Track, *spotify.Track)) {
	for {
		if s.spotify.IsConnected() {
			if !s.hasState() {
				if track, err := s.spotify.GetSpotifyStatus(); err != nil {
					s.spotify.OnError()
					continue
				} else {
					if s.spotify.PollRate > spotify.DefaultPollRate {
						s.spotify.PollRate = spotify.DefaultPollRate
					}
					s.setState(track)
				}
				continue
			} else if len(id) > 0 {
				if track, err := s.spotify.GetSpotifyStatus(); err != nil {
					s.spotify.OnError()
					continue
				} else {
					if track.IsPlaying && s.spotify.PollRate > spotify.DefaultPollRate {
						s.spotify.PollRate = spotify.DefaultPollRate
					}
					oldTrack := s.getState()
					onData(track, oldTrack)
					s.setState(track)
				}
			}
		}
		time.Sleep(s.spotify.PollRate * time.Second)
	}
}
