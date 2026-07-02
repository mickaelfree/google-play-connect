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

func TestBundlesList(t *testing.T) {
	discarded := false
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tmp"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tmp/bundles", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.BundlesListResponse{
			Bundles: []*androidpublisher.Bundle{{VersionCode: 7, Sha256: "def456"}},
		})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tmp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			discarded = true
		}
		w.WriteHeader(http.StatusOK)
	})

	root, out := newTestRoot(t, mux)
	root.SetArgs([]string{"bundles", "list", "--app", "com.example.app"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("7")) {
		t.Fatalf("output missing version code: %s", out.String())
	}
	if !discarded {
		t.Fatal("read path must discard its edit")
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

func TestStatusScanAllSkips404Tracks(t *testing.T) {
	mux := http.NewServeMux()
	// Only production has releases; internal/alpha/beta hit the mux default 404.
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/tracks/production/releases", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.ListReleaseSummariesResponse{
			Releases: []*androidpublisher.ReleaseSummary{
				{ReleaseName: "42 (1.2.3)", ReleaseLifecycleState: "ACTIVE", Track: "production"},
			},
		})
	})

	root, out := newTestRoot(t, mux)
	root.SetArgs([]string{"status", "--app", "com.example.app"})
	if err := root.Execute(); err != nil {
		t.Fatalf("scan-all must skip 404 tracks, got: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("ACTIVE")) {
		t.Fatalf("production release missing from output: %s", out.String())
	}
}

func TestStatusScanAllSurfacesNon404Errors(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/tracks/internal/releases", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"code":401,"message":"unauthorized"}}`, http.StatusUnauthorized)
	})

	root, _ := newTestRoot(t, mux)
	root.SetArgs([]string{"status", "--app", "com.example.app"})
	if err := root.Execute(); err == nil {
		t.Fatal("a 401 during scan-all must surface as an error, not an empty success")
	}
}
