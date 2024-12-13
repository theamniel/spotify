package config

type (
	Config struct {
		Server  *ServerConfig  `toml:"server"`
		Grpc    *GrpcConfig    `toml:"grpc"`
		Socket  *SocketConfig  `toml:"socket"`
		Spotify *SpotifyConfig `toml:"spotify"`
	}

	ServerConfig struct {
		Host     string `toml:"host"`
		Port     string `toml:"port"`
		TimeZone string `toml:"timeZone"`
		Prefork  bool   `toml:"prefork"`
	}

	GrpcConfig struct {
		Host string `toml:"host"`
		Port string `toml:"port"`
	}

	SocketConfig struct {
		Origins         []string `toml:"origins"`
		ReadBufferSize  int      `toml:"readBufferSize"`
		WriteBufferSize int      `toml:"writeBufferSize"`
	}

	SpotifyConfig struct {
		ClientID     string `json:"clientID"`
		ClientSecret string `json:"clientSecret"`
		RefreshToken string `json:"refreshToken"`
	}
)
