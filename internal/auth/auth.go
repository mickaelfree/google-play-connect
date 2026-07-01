// Package auth resolves Google service-account credentials for gpc without
// ever falling back to Application Default Credentials: the key must come
// from the --service-account flag or a GPC_* environment variable.
package auth

import (
	"errors"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

// Config carries the CLI-supplied override for the service account key path.
type Config struct {
	ServiceAccountPath string
}

// Credentials describes where the service account key was found.
type Credentials struct {
	Path string
	JSON []byte
}

// ErrNoCredentials is returned when no credential source is configured.
var ErrNoCredentials = errors.New("no service account credentials found: set --service-account, GPC_SERVICE_ACCOUNT_KEY_PATH, or GPC_SERVICE_ACCOUNT_KEY_JSON")

// ResolveCredentials picks a credential source in priority order:
//  1. cfg.ServiceAccountPath (the --service-account flag)
//  2. GPC_SERVICE_ACCOUNT_KEY_PATH environment variable
//  3. GPC_SERVICE_ACCOUNT_KEY_JSON environment variable (inline JSON)
func ResolveCredentials(cfg Config, getenv func(string) string) (Credentials, error) {
	if cfg.ServiceAccountPath != "" {
		return Credentials{Path: cfg.ServiceAccountPath}, nil
	}
	if path := getenv("GPC_SERVICE_ACCOUNT_KEY_PATH"); path != "" {
		return Credentials{Path: path}, nil
	}
	if json := getenv("GPC_SERVICE_ACCOUNT_KEY_JSON"); json != "" {
		return Credentials{JSON: []byte(json)}, nil
	}
	return Credentials{}, ErrNoCredentials
}

// ClientOptions builds the option.ClientOption list needed to construct an
// androidpublisher.Service from these credentials.
func (c Credentials) ClientOptions() ([]option.ClientOption, error) {
	switch {
	case len(c.JSON) > 0:
		return []option.ClientOption{
			option.WithAuthCredentialsJSON(option.ServiceAccount, c.JSON),
			option.WithScopes(androidpublisher.AndroidpublisherScope),
		}, nil
	case c.Path != "":
		return []option.ClientOption{
			option.WithAuthCredentialsFile(option.ServiceAccount, c.Path),
			option.WithScopes(androidpublisher.AndroidpublisherScope),
		}, nil
	default:
		return nil, ErrNoCredentials
	}
}
