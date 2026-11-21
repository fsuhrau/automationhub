// Package oauthflow is the OAuth2 authorization-code sign in shared by the
// GitHub and the generic OAuth2 provider. A Strategy supplies what differs:
// the endpoints and the mapping of the provider's user to a Profile.
package oauthflow

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

// Profile is the provider's user, mapped to what the hub stores.
type Profile struct {
	// ProviderUserID identifies the user at the provider (UserAuth.ProviderUserID).
	ProviderUserID string
	Login          string
	Name           string
	Email          string
	URL            string
}

// Strategy is the provider specific part of the flow.
type Strategy interface {
	// Name is the provider name stored with the user and reported as sign in method.
	Name() string
	// Endpoint holds the authorization and token URLs.
	Endpoint() oauth2.Endpoint
	// Profile fetches the signed-in user; client carries the access token.
	Profile(ctx context.Context, client *http.Client) (Profile, error)
	// SuccessRedirect is where the browser goes after signing in, "" for nowhere.
	SuccessRedirect() string
}

// Settings are the configured values every provider needs.
type Settings struct {
	RedirectURL string
	CredFile    string
	Scopes      []string
	Secret      []byte
}
