// Package playapitest provides a fake androidpublisher.Service backed by an
// httptest.Server, so internal/playapi can be unit tested without real
// network calls or Google credentials.
package playapitest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

// NewService starts an httptest.Server driven by handler and returns an
// androidpublisher.Service pointed at it with authentication disabled. The
// server is closed automatically via t.Cleanup.
func NewService(t *testing.T, handler http.Handler) (*androidpublisher.Service, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	svc, err := androidpublisher.NewService(context.Background(),
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("build test androidpublisher service: %v", err)
	}
	return svc, server
}
