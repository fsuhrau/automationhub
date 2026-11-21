package oauthflow

import (
	"encoding/json"
	"fmt"
	"os"
)

// Credentials are the client id and secret of the OAuth2 application.
type Credentials struct {
	ClientID     string `json:"clientid"`
	ClientSecret string `json:"secret"`
}

// LoadCredentials reads the client credentials from a JSON file.
func LoadCredentials(path string) (Credentials, error) {
	var c Credentials
	data, err := os.ReadFile(path)
	if err != nil {
		return c, fmt.Errorf("read credentials: %w", err)
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("parse credentials: %w", err)
	}
	return c, nil
}
