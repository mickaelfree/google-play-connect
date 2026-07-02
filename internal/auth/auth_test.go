package auth_test

import (
	"errors"
	"testing"

	"github.com/mickaelfree/google-play-connect/internal/auth"
)

func envOf(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestResolveCredentialsFlagWins(t *testing.T) {
	creds, err := auth.ResolveCredentials(
		auth.Config{ServiceAccountPath: "/keys/flag.json"},
		envOf(map[string]string{"GPC_SERVICE_ACCOUNT_KEY_PATH": "/keys/env.json"}),
	)
	if err != nil {
		t.Fatalf("ResolveCredentials: %v", err)
	}
	if creds.Path != "/keys/flag.json" {
		t.Fatalf("got path %q, want /keys/flag.json", creds.Path)
	}
}

func TestResolveCredentialsEnvPath(t *testing.T) {
	creds, err := auth.ResolveCredentials(
		auth.Config{},
		envOf(map[string]string{"GPC_SERVICE_ACCOUNT_KEY_PATH": "/keys/env.json"}),
	)
	if err != nil {
		t.Fatalf("ResolveCredentials: %v", err)
	}
	if creds.Path != "/keys/env.json" {
		t.Fatalf("got path %q, want /keys/env.json", creds.Path)
	}
}

func TestResolveCredentialsInlineJSON(t *testing.T) {
	creds, err := auth.ResolveCredentials(
		auth.Config{},
		envOf(map[string]string{"GPC_SERVICE_ACCOUNT_KEY_JSON": `{"type":"service_account"}`}),
	)
	if err != nil {
		t.Fatalf("ResolveCredentials: %v", err)
	}
	if string(creds.JSON) != `{"type":"service_account"}` {
		t.Fatalf("got JSON %q", creds.JSON)
	}
}

func TestResolveCredentialsMissing(t *testing.T) {
	_, err := auth.ResolveCredentials(auth.Config{}, envOf(nil))
	if !errors.Is(err, auth.ErrNoCredentials) {
		t.Fatalf("got %v, want ErrNoCredentials", err)
	}
}

func TestClientOptionsFromPathAndJSON(t *testing.T) {
	fromPath := auth.Credentials{Path: "/keys/x.json"}
	opts, err := fromPath.ClientOptions()
	if err != nil || len(opts) != 2 {
		t.Fatalf("path options: %v (len %d)", err, len(opts))
	}
	fromJSON := auth.Credentials{JSON: []byte(`{}`)}
	opts, err = fromJSON.ClientOptions()
	if err != nil || len(opts) != 2 {
		t.Fatalf("json options: %v (len %d)", err, len(opts))
	}
	_, err = auth.Credentials{}.ClientOptions()
	if !errors.Is(err, auth.ErrNoCredentials) {
		t.Fatalf("empty creds: got %v, want ErrNoCredentials", err)
	}
}
