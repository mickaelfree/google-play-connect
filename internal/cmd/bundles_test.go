package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

func TestBundlesUpload(t *testing.T) {
	dir := t.TempDir()
	aabPath := filepath.Join(dir, "app.aab")
	if err := os.WriteFile(aabPath, []byte("aab-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})
	mux.HandleFunc("/upload/androidpublisher/v3/applications/com.example.app/edits/tx/bundles", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.Bundle{VersionCode: 42, Sha256: "abc123"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx:commit", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})

	root, out := newTestRoot(t, mux)
	root.SetArgs([]string{"bundles", "upload", "--app", "com.example.app", "--file", aabPath})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("42")) {
		t.Fatalf("output missing version code: %s", out.String())
	}
}

func TestStatusListsReleaseSummaries(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/tracks/production/releases", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.ListReleaseSummariesResponse{
			Releases: []*androidpublisher.ReleaseSummary{
				{ReleaseName: "42 (1.2.3)", ReleaseLifecycleState: "ACTIVE", Track: "production"},
			},
		})
	})

	root, out := newTestRoot(t, mux)
	root.SetArgs([]string{"status", "--app", "com.example.app", "--track", "production"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("ACTIVE")) {
		t.Fatalf("output missing release state: %s", out.String())
	}
}
