package config

import (
	"os"

	"spotify/utils"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
)

func Load[T any]() (*T, error) {
	dir, _, err := utils.Executable()
	if err != nil {
		return nil, err
	}

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return LoadFile[T](dir + "config.toml")
}

func LoadFile[T any](fileName string) (*T, error) {
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

	var cfg T
	if err := toml.Unmarshal(utils.ReplaceValues(data), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
