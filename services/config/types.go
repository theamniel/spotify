package config

type (
	Config struct {
		App        *AppConfig        `toml:"app"`
		Middleware *MiddlewareConfig `toml:"middleware"`
		Server     *ServerConfig     `toml:"server"`
		Socket     *SocketConfig     `toml:"socket"`
		Spotify    *SpotifyConfig    `toml:"spotify"`
	}

	AppConfig struct {
		Limit         int  `toml:"limit"`
		Prefork       bool `toml:"prefork"`
		StrictRouting bool `toml:"strictRouting"`
		CaseSensitive bool `toml:"caseSensitive"`
		UnescapePath  bool `toml:"unescapePath"`
	}

	MiddlewareConfig struct {
		Cache    bool `toml:"cache"`
		Compress bool `toml:"compress"`
		Logger   bool `toml:"logger"`
		Recover  bool `toml:"recover"`
	}

	ServerConfig struct {
		Host  string `toml:"host"`
		Port  string `toml:"port"`
		Token string `toml:"token"`
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
