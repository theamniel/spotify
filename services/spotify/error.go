package spotify

import "encoding/json"

type SpotifyApiError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type SpotifyApiErrorResponse struct {
	Error *SpotifyApiError `json:"error"`
}

func NewApiErrorFrom(bytes []byte) *SpotifyApiError {
	var data SpotifyApiErrorResponse
	if err := json.Unmarshal(bytes, &data); err != nil {
		data.Error.Status = 500
		data.Error.Message = err.Error()
	}
	return data.Error
}

func NewApiError(code int, message string) *SpotifyApiError {
	return &SpotifyApiError{
		Status:  code,
		Message: message,
	}
}
