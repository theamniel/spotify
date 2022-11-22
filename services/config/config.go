package config

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/theamniel/spotify-server/utils"
)

func Load() (*Config, error) {
	dir, file, err := utils.Executable()
	if err != nil {
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

	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(utils.ReplaceValues(fileBytes), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
