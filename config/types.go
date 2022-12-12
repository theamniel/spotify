package config

type (
	Config struct {
		Server  *ServerConfig  `toml:"server"`
		Socket  *SocketConfig  `toml:"socket"`
		Spotify *SpotifyConfig `toml:"spotify"`
	}

	ServerConfig struct {
		Host    string `toml:"host"`
		Port    string `toml:"port"`
		Token   string `toml:"token"`
		Prefork bool   `toml:"prefork"`
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
