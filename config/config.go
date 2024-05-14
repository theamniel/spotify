package config

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
	"spotify.amniel/utils"
)

func Load() (*Config, error) {
	dir, file, err := utils.Executable()
	if err != nil {
		return nil, err
	}

	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	return LoadFile(dir + file + ".toml")
}

func LoadFile(fileName string) (*Config, error) {
	if fileName[len(fileName)-5:] != ".toml" {
		fileName += ".toml"
	}

	if _, err := os.Stat(fileName); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(utils.ReplaceValues(data), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
