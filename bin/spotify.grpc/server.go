package main

import (
	"context"
	"sync"

	"spotify/services/grpc/proto"
	"spotify/services/spotify"

	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedSpotifyServer
	spotify *spotify.SpotifyClient

	state *spotify.Track
	mu    sync.RWMutex
}

func (s *server) setState(value *spotify.Track) {
	s.mu.Lock()
	if s.state == nil {
		s.state = value
	} else {
		*s.state = *value
	}
	s.mu.Unlock()
}

func (s *server) getState() *spotify.Track {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *server) hasState() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state != nil
}

func (s *server) GetTrack(_ context.Context, req *proto.Request) (*proto.Track, error) {
	trackResult := &spotify.Track{}
	for {
		if track, err := s.spotify.GetSpotifyStatus(); err != nil {
			s.spotify.OnError()
			continue
		} else {
			trackResult = track
		}
		break
	}
	return trackResult.ToProto(), nil
}

func (s *server) OnListen(req *proto.Request, stream grpc.ServerStreamingServer[proto.Reponse]) error {
	id := req.GetID()
	s.pool(id, func(track, oldTrack *spotify.Track) {
		if track != nil && oldTrack != nil {
			if track.ID != oldTrack.ID {
				stream.Send(&proto.Reponse{
					ID:       id,
					E:        "CHANGE",
					Track:    track.ToProto(),
					Progress: nil,
				})
			}

			if track.Timestamp != nil {
				progress := int64(track.Timestamp.Progress)
				stream.Send(&proto.Reponse{ID: id, E: "PROGRESS", Track: nil, Progress: &progress})
			}
		}
	})
	return nil
}
