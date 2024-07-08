package utils

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"regexp"
)

var valueRegex = regexp.MustCompile(`[#]\{([\w\.]+)\}`)

// Replace values from string in os environement (source)
func ReplaceValues(bsrc []byte) []byte {
	for _, items := range valueRegex.FindAllSubmatch(bsrc, -1) {
		env := os.Getenv(string(items[1]))
		if env != "" {
			bsrc = bytes.ReplaceAll(bsrc, items[0], []byte(env))
		}
	}
	return bsrc
}

// Encode string to base64 (base64)
func EncodeToBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// Get Executable file and dir (dir, file, error)
func Executable() (string, string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", "", err
	}
	dir, file := filepath.Split(executable)
	if file[len(file)-4:] == ".exe" {
		file = file[:len(file)-4]
	}
	return dir, file, nil
}
