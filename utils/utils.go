package utils

import (
	"encoding/base64"
	"os"
	"runtime"
	"strings"
)

func EncodeToBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func ExePathName() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		return strings.Replace(exePath, ".exe", "", 1), nil
	}
	return exePath, nil
}
