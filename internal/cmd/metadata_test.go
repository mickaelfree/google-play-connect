package cmd_test

import (
	"encoding/json"
	"image"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/mickaelfree/google-play-connect/internal/metadata"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

// writePNG creates a real PNG of the given size at path.
func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
}

func TestMetadataPullWritesListings(t *testing.T) {
	dir := t.TempDir()
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tmp"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tmp/listings", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.ListingsListResponse{
			Listings: []*androidpublisher.Listing{
				{Language: "fr-FR", Title: "Mon App", ShortDescription: "Courte", FullDescription: "Longue"},
			},
		})
	})
	// Image listing for every known type × the one locale; return one icon.
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tmp/listings/fr-FR/icon", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.ImagesListResponse{
			Images: []*androidpublisher.Image{{Id: "i1", Sha1: "aa", Url: "https://x/i1"}},
		})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Other image types → empty list; edit delete → OK.
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			return
		}
		json.NewEncoder(w).Encode(androidpublisher.ImagesListResponse{})
	})

	root, _ := newTestRoot(t, mux)
	root.SetArgs([]string{"metadata", "pull", "--app", "com.example.app", "--dir", dir})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	l, err := metadata.ReadListing(dir, "com.example.app", "fr-FR")
	if err != nil || l.Title != "Mon App" {
		t.Fatalf("listing not pulled: %+v / %v", l, err)
	}
	manifest := filepath.Join(dir, "com.example.app", "images", "fr-FR", "icon.manifest.json")
	if _, err := os.Stat(manifest); err != nil {
		t.Fatalf("icon manifest missing: %v", err)
	}
}

func TestMetadataPushRequiresConfirm(t *testing.T) {
	dir := t.TempDir()
	if err := metadata.WriteListing(dir, "com.example.app", "fr-FR", metadata.Listing{Title: "Mon App"}); err != nil {
		t.Fatal(err)
	}
	root, _ := newTestRoot(t, http.NewServeMux())
	root.SetArgs([]string{"metadata", "push", "--app", "com.example.app", "--dir", dir})
	if err := root.Execute(); err == nil {
		t.Fatal("push without --confirm must fail")
	}
}

func TestMetadataPushUploadsListingsAndImages(t *testing.T) {
	dir := t.TempDir()
	if err := metadata.WriteListing(dir, "com.example.app", "fr-FR", metadata.Listing{Title: "Mon App", ShortDescription: "Courte"}); err != nil {
		t.Fatal(err)
	}
	shotDir := filepath.Join(dir, "com.example.app", "images", "fr-FR", "phoneScreenshots")
	writePNG(t, filepath.Join(shotDir, "1.png"), 1080, 1920)

	var updatedListing, deletedAll, uploaded, committed bool
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx/listings/fr-FR", func(w http.ResponseWriter, r *http.Request) {
		updatedListing = true
		var body androidpublisher.Listing
		_ = json.NewDecoder(r.Body).Decode(&body)
		json.NewEncoder(w).Encode(body)
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx/listings/fr-FR/phoneScreenshots", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deletedAll = true
			json.NewEncoder(w).Encode(androidpublisher.ImagesDeleteAllResponse{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/upload/androidpublisher/v3/applications/com.example.app/edits/tx/listings/fr-FR/phoneScreenshots", func(w http.ResponseWriter, r *http.Request) {
		uploaded = true
		json.NewEncoder(w).Encode(androidpublisher.ImagesUploadResponse{Image: &androidpublisher.Image{Id: "new-1"}})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx:commit", func(w http.ResponseWriter, r *http.Request) {
		committed = true
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})

	root, _ := newTestRoot(t, mux)
	root.SetArgs([]string{"metadata", "push", "--app", "com.example.app", "--dir", dir, "--confirm"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !updatedListing || !deletedAll || !uploaded || !committed {
		t.Fatalf("push incomplete: listing=%v deleteAll=%v upload=%v commit=%v",
			updatedListing, deletedAll, uploaded, committed)
	}
}

func TestMetadataValidateOffline(t *testing.T) {
	dir := t.TempDir()
	if err := metadata.WriteListing(dir, "com.example.app", "en-US", metadata.Listing{Title: "This title is way way way too long for Google Play"}); err != nil {
		t.Fatal(err)
	}
	// No HTTP handlers: validate must never call the API.
	root, out := newTestRoot(t, http.NewServeMux())
	root.SetArgs([]string{"metadata", "validate", "--app", "com.example.app", "--dir", dir})
	err := root.Execute()
	if err == nil {
		t.Fatal("validate with issues must exit non-zero")
	}
	if !json.Valid(out.Bytes()) {
		t.Fatalf("issues must still render as JSON: %s", out.String())
	}
}
