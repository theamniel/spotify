package config

import (
	"bytes"
	"io"
	"os"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/theamniel/spotify-server/utils"
)

var Envars = regexp.MustCompile(`[#]\{([\w\.]+)\}`)

func Load() (*Config, error) {
	exePathName, err := utils.ExePathName()
	if err != nil {
		return nil, err
	}
	exePathName += ".toml"

	if _, err := os.Stat(exePathName); err != nil {
		return nil, err
	}

	file, err := os.Open(exePathName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	dataFile, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	for _, values := range Envars.FindAllSubmatch(dataFile, -1) {
		env := os.Getenv(string(values[1]))
		if env != "" {
			dataFile = bytes.ReplaceAll(dataFile, values[0], []byte(env))
		}
	}

	var cfg Config
	if err := toml.Unmarshal(dataFile, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
