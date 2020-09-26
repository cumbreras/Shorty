package model

import (
	"encoding/json"
	"io"
)

// Model exposes the standard model interface
type Model interface {
	ToJSON(io.Writer) error
	FromJSON(io.Reader) error
}

// ShortenURL is the model for storage
type ShortenURL struct {
	Code string `json:"code"`
	URL  string `json:"url"`
	Model
}

// New creates a new ShortenURL
func New() *ShortenURL {
	return &ShortenURL{}
}

// ToJSON Encodes struct to JSON
func (us *ShortenURL) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(us)
}

// FromJSON Decodes JSON to struct
func (us *ShortenURL) FromJSON(body io.Reader) error {
	e := json.NewDecoder(body)
	return e.Decode(us)
}
